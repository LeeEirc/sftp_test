package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
	"sftp_test/pkg/config"
)

var cfgFile string

func init() {
	flag.StringVar(&cfgFile, "config", "config.yml", "config yml file")
}

func main() {
	flag.Parse()
	cfg := LoadConfig(cfgFile)
	slog.Info("config loaded", "config", cfg)
	dstAddr := net.JoinHostPort(cfg.DstHost, strconv.Itoa(cfg.DstPort))
	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	s := Server{
		Addr:    listenAddr,
		DstAddr: dstAddr,
	}
	slog.Info("server starting", "server", s)
	if err := s.Start(); err != nil {
		slog.Error("server failed", "error", err)
	}

}

type Server struct {
	Addr    string
	DstAddr string
}

func (s Server) Start() error {
	slog.Info("server started ", "addr", s.Addr, "dstAddr", s.DstAddr)
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	for {
		conn, err1 := ln.Accept()
		if err1 != nil {
			return err1
		}
		go s.handleConn(conn)
	}
}

func (s Server) handleConn(conn net.Conn) {
	defer conn.Close()
	dstConn, err := net.Dial("tcp", s.DstAddr)
	if err != nil {
		slog.Error("failed to dial destination", "error", err)
		return
	}
	defer dstConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err1 := io.Copy(dstConn, conn)
		if err1 != nil {
			slog.Error("failed to copy to destination", "error", err1)
		}
	}()

	go func() {
		defer wg.Done()
		_, err2 := io.Copy(conn, dstConn)
		if err2 != nil {
			slog.Error("failed to copy from destination", "error", err2)
		}
	}()
	wg.Wait()
	slog.Info("connection closed")
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
