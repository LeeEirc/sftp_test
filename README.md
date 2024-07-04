## golang 通过不同方式的代理 sftp server 对比

### 下载 Release 的二进制文件
```bash

sftptcpproxy-linux-amd64

sftpserver-linux-amd64

```

### 创建两个不同的 config1.yml 和 config2.yml 配置文件
```yaml

PORT: 2234  ## 监听端口
PASSWORD: 123456 ## sftpserver 的登陆
DST_HOST: xxx.xxx.xxx.xxx ##  sftp 代理机器的 ip
DST_PORT: 22 ## sftp 代理机器的端口
DST_USERNAME: root ## sftp 代理机器的用户名
DST_PASSWORD: passwd ## sftp 代理机器的密码
```

### 分别启动 sftptcpproxy 和 sftpserver
```bash

./sftptcpproxy-linux-amd64 -config config1.yml

./sftpserver-linux-amd64 -config config2.yml
```