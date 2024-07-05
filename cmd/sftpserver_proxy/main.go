package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"sftp_test/pkg/config"
)

var cfgFile string

func init() {
	flag.StringVar(&cfgFile, "config", "config.yml", "config yml file")
}

func LoadConfig(path string) config.Config {
	buf, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var cfg config.Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func main() {
	flag.Parse()
	cfg := LoadConfig(cfgFile)
	slog.Info("config loaded", "config", cfg)
	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	s := Server{
		Addr: listenAddr,
		cfg:  &cfg,
	}
	slog.Info("server starting", "server", s)
	if err := s.Start(); err != nil {
		slog.Error("server failed", "error", err)
	}
}

//go:embed id_ras
var privateKey []byte

func LoadPrivateKey() (ssh.Signer, error) {
	return gossh.ParsePrivateKey(privateKey)
}

type Server struct {
	Addr string
	cfg  *config.Config
}

func (s Server) Start() error {
	slog.Info("server started ", "addr", s.Addr)
	singer, err := LoadPrivateKey()
	if err != nil {
		slog.Error("failed to load private key", "error", err)
		return err
	}
	srv := &ssh.Server{
		Addr:        s.Addr,
		HostSigners: []ssh.Signer{singer},
		PasswordHandler: func(ctx ssh.Context, pass string) bool {
			return pass == s.cfg.Password
		},
		Handler: func(s ssh.Session) {
			fmt.Fprintln(s, "Hello, world!")
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": s.sftpSubsystemHandler,
		},
	}
	return srv.ListenAndServe()
}

func (s Server) sftpSubsystemHandler(sess ssh.Session) {
	sshClient, err := s.createRemoteSSH()
	if err != nil {
		slog.Error("failed to create remote ssh", "error", err)
		return
	}
	defer sshClient.Close()
	sftpClient, err1 := sftp.NewClient(sshClient)
	if err1 != nil {
		slog.Error("failed to create remote sftp", "error", err1)
		return
	}
	remoteSftp := &RemoteSFTP{client: sshClient, sftpClient: sftpClient}
	defer remoteSftp.Close()
	handlers := sftp.Handlers{
		FileGet:  remoteSftp,
		FilePut:  remoteSftp,
		FileCmd:  remoteSftp,
		FileList: remoteSftp,
	}
	sftpSrv := sftp.NewRequestServer(sess, handlers)
	err = sftpSrv.Serve()
	if err != nil {
		slog.Error("sftp server failed", "error", err)
	}
	slog.Info("sftp server closed")

}

func (s Server) createRemoteSSH() (*gossh.Client, error) {

	authMethods := make([]gossh.AuthMethod, 0, 3)
	authMethods = append(authMethods, gossh.Password(s.cfg.DstPassword))
	keyboardAuth := func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		if len(questions) == 0 {
			return []string{}, nil
		}
		return []string{s.cfg.DstPassword}, nil
	}
	authMethods = append(authMethods, gossh.KeyboardInteractive(keyboardAuth))
	gosshCfg := &gossh.ClientConfig{
		User:            s.cfg.DstUsername,
		Auth:            authMethods,
		Timeout:         time.Second * 30,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}
	dstAddr := net.JoinHostPort(s.cfg.DstHost, strconv.Itoa(s.cfg.DstPort))
	client, err := gossh.Dial("tcp", dstAddr, gosshCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type RemoteSFTP struct {
	client     *gossh.Client
	sftpClient *sftp.Client
}

func (r *RemoteSFTP) Fileread(req *sftp.Request) (io.ReaderAt, error) {
	sftpfd, err := r.sftpClient.Open(req.Filepath)
	if err != nil {
		return nil, err
	}
	go func() {
		<-req.Context().Done()
		if err1 := sftpfd.Close(); err1 != nil {
			slog.Error("remote sftp file close error", "error", err1, "file", req.Filepath)
		}
		slog.Info("sftp file read done", "file", req.Filepath)
	}()
	return &LockFile{
		fd: sftpfd,
	}, nil
}

func (r *RemoteSFTP) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	sftpfd, err := r.sftpClient.Create(req.Filepath)
	if err != nil {
		return nil, err
	}
	go func() {
		<-req.Context().Done()
		if err1 := sftpfd.Close(); err1 != nil {
			slog.Error("remote sftp file close error", "error", err1, "file", req.Filepath)
		}
		slog.Info("sftp file write done", "file", req.Filepath)
	}()
	return &LockFile{
		fd: sftpfd,
	}, nil
}

func (r *RemoteSFTP) Filecmd(req *sftp.Request) error {
	switch req.Method {
	case "Setstat":
		return nil
	case "Rename":
		return r.sftpClient.Rename(req.Filepath, req.Target)
	case "Rmdir":
		return r.sftpClient.RemoveDirectory(req.Filepath)
	case "Remove":
		return r.sftpClient.Remove(req.Filepath)
	case "Mkdir":
		return r.sftpClient.MkdirAll(req.Filepath)
	case "Symlink":
		return r.sftpClient.Symlink(req.Filepath, req.Target)
	default:
		return fmt.Errorf("unsupported method: %s", req.Method)

	}
}

func (r *RemoteSFTP) Filelist(req *sftp.Request) (sftp.ListerAt, error) {
	switch req.Method {
	case "List":
		res, err := r.sftpClient.ReadDir(req.Filepath)
		if err != nil {
			return nil, err
		}
		return listerat(res), nil
	case "Stat":
		fsInfo, err := r.sftpClient.Stat(req.Filepath)
		if err != nil {
			return nil, err
		}
		return listerat([]os.FileInfo{fsInfo}), nil
	default:
		return nil, sftp.ErrSshFxOpUnsupported
	}
}

func (r *RemoteSFTP) Close() error {
	_ = r.sftpClient.Close()
	_ = r.client.Close()
	return nil
}

type listerat []os.FileInfo

func (f listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	var n int
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n = copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

type LockFile struct {
	fd *sftp.File
	sync.Mutex
}

func (l *LockFile) WriteAt(p []byte, off int64) (n int, err error) {
	l.Lock()
	defer l.Unlock()
	return l.fd.WriteAt(p, off)
}

func (l *LockFile) ReadAt(p []byte, off int64) (n int, err error) {
	l.Lock()
	defer l.Unlock()
	return l.fd.ReadAt(p, off)
}
