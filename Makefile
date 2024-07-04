SFTPSERVER=sftpserver
TCPPROXY=sftptcpproxy

BUILDDIR=build

define make_artifact_full
	@echo "build $(1)-$(2)"
	GOOS=$(1) GOARCH=$(2) go build -o $(BUILDDIR)/${SFTPSERVER}-$(1)-$(2) ./cmd/sftpserver_proxy
	GOOS=$(1) GOARCH=$(2) go build -o $(BUILDDIR)/${TCPPROXY}-$(1)-$(2) ./cmd/tcp_proxy
endef

all:
	$(call make_artifact_full,darwin,amd64)
	$(call make_artifact_full,darwin,arm64)
	$(call make_artifact_full,linux,amd64)
	$(call make_artifact_full,linux,arm64)
	$(call make_artifact_full,linux,mips64le)
	$(call make_artifact_full,linux,ppc64le)
	$(call make_artifact_full,linux,s390x)
	$(call make_artifact_full,linux,riscv64)

darwin-amd64:
	$(call make_artifact_full,darwin,amd64)

darwin-arm64:
	$(call make_artifact_full,darwin,arm64)

linux-amd64:
	$(call make_artifact_full,linux,amd64)

linux-arm64:
	$(call make_artifact_full,linux,arm64)

linux-loong64:
	$(call make_artifact_full,linux,loong64)

linux-mips64le:
	$(call make_artifact_full,linux,mips64le)

linux-ppc64le:
	$(call make_artifact_full,linux,ppc64le)

linux-s390x:
	$(call make_artifact_full,linux,s390x)

linux-riscv64:
	$(call make_artifact_full,linux,riscv64)

clean:
	-rm -rf $(BUILDDIR)