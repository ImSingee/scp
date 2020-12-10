package scp

import (
	"fmt"
	"github.com/ImSingee/tt"
	"github.com/alessio/shellescape"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"testing"
)

// Underlying client cannot be closed
func getSshSession(t *testing.T) *ssh.Session {
	client, err := ssh.Dial("tcp", "localhost:22", &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Getenv("SSH_PASSWORD")),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback:  ssh.BannerDisplayStderr(),
	})
	tt.AssertIsNil(t, err)

	session, err := client.NewSession()
	tt.AssertIsNil(t, err)

	return session
}

func reset(t *testing.T) {
	session := getSshSession(t)
	defer session.Close()

	err := session.Run("rm -rf /test; mkdir /test")
	tt.AssertIsNil(t, err)
}

func checkFileExist(t *testing.T, filename string) bool {
	session := getSshSession(t)
	defer session.Close()

	output, err := session.Output("stat " + shellescape.Quote(filename))

	if err != nil {
		return false
	}

	return len(output) != 0
}

func readFile(t *testing.T, filename string) []byte {
	session := getSshSession(t)
	defer session.Close()

	output, err := session.Output("cat " + shellescape.Quote(filename))
	tt.AssertIsNil(t, err)

	return output
}

func TestWriteFileToFile(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	client, err := NewClient(session)
	tt.AssertIsNil(t, err)

	err = client.Start("/test/some", false)
	tt.AssertIsNil(t, err)

	// filename (test01) can be difference from real (some)
	err = client.WriteFile("0644", 6, "test01", strings.NewReader("hahaha"))
	tt.AssertIsNil(t, err)

	tt.AssertTrue(t, checkFileExist(t, "/test/some"))

	file := readFile(t, "/test/some")

	tt.AssertEqual(t, "hahaha", string(file))
}

func TestWriteFileToFolder(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	client, err := NewClient(session)
	tt.AssertIsNil(t, err)

	err = client.Start("/test", false)
	tt.AssertIsNil(t, err)

	err = client.WriteFile("0644", 6, "test01", strings.NewReader("hahaha"))
	tt.AssertIsNil(t, err)

	tt.AssertTrue(t, checkFileExist(t, "/test/test01"))

	file := readFile(t, "/test/test01")

	tt.AssertEqual(t, "hahaha", string(file))
}

func TestWriteFileToUnExistFolder(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	client, err := NewClient(session)
	tt.AssertIsNil(t, err)

	err = client.Start("/test/not/exist/folder", false)
	tt.AssertIsNil(t, err)

	err = client.WriteFile("0644", 6, "test01", strings.NewReader("hahaha"))
	tt.AssertIsNotNil(t, err)

	fmt.Println(err)

	tt.AssertFalse(t, checkFileExist(t, "/test/not/exist/folder"))
}
