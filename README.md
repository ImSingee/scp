# scp in Go

## Install

```bash
go get -u github.com/ImSingee/scp
```

## Usage

[![Go Reference](https://pkg.go.dev/badge/github.com/ImSingee/scp.svg)](https://pkg.go.dev/github.com/ImSingee/scp)

### Simple Copy

```go
package main
import "golang.org/x/crypto/ssh"
import "github.com/ImSingee/scp"

func main() {
    // make session
    client, _ := ssh.Dial("tcp", "remote-addr:22", &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback:  ssh.BannerDisplayStderr(),
	})
    defer client.Close()

	session, _ := client.NewSession()
    defer session.Close()
    
    // Copy!
    err := scp.Copy(session, "/path/to/local/file", "/path/to/remote/file")
    
    if err != nil {
    panic(err)
    }
}
```

The behavior is same as `scp [-r] /path/to/local/file remote-addr:/path/to/remote/file`

### Custom Use

Please read [protocol_test.go](https://github.com/ImSingee/scp/blob/master/protocol_test.go)

