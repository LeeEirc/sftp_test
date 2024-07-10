module sftp_test

go 1.22

require (
	github.com/gliderlabs/ssh v0.3.7
	github.com/pkg/sftp v1.13.6
	golang.org/x/crypto v0.24.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)

replace github.com/gliderlabs/ssh => github.com/jumpserver-dev/ssh v0.3.10
