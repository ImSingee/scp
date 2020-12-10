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

func TestWriteMultipleFilesToFolder(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	client, err := NewClient(session)
	tt.AssertIsNil(t, err)

	err = client.Start("/test", false)
	tt.AssertIsNil(t, err)

	err = client.WriteFile("0644", 6, "test01", strings.NewReader("hahaha"))
	tt.AssertIsNil(t, err)

	err = client.WriteFile("0644", 8, "test02", strings.NewReader("xixixixi"))
	tt.AssertIsNil(t, err)

	tt.AssertTrue(t, checkFileExist(t, "/test/test01"))
	tt.AssertTrue(t, checkFileExist(t, "/test/test02"))

	tt.AssertEqual(t, "hahaha", string(readFile(t, "/test/test01")))
	tt.AssertEqual(t, "xixixixi", string(readFile(t, "/test/test02")))
}

func TestWriteDirectoryToFolder(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	client, err := NewClient(session)
	tt.AssertIsNil(t, err)

	/*
		Virtual File Tree
			-- (Folder) a
			  -- (File) b (Content: hahaha)
			  -- (File) c (Content: xixixixi)
			-- (Folder) b
			-- (Folder) c
			  -- (File) d (Empty)
			-- (File) e (Content: root)
	*/

	err = client.Start("/test", true)
	tt.AssertIsNil(t, err)

	// first folder a
	err = client.WriteDirectoryStart("0755", "a")
	tt.AssertIsNil(t, err)

	// in folder a, file b  (Content: hahaha)
	err = client.WriteFile("0644", 6, "b", strings.NewReader("hahaha"))
	tt.AssertIsNil(t, err)
	// in folder a, file c (Content: xixixixi)
	err = client.WriteFile("0644", 8, "c", strings.NewReader("xixixixi"))
	tt.AssertIsNil(t, err)

	// end first folder a
	err = client.WriteDirectoryEnd()
	tt.AssertIsNil(t, err)

	// second folder b
	err = client.WriteDirectoryStart("0755", "b")
	tt.AssertIsNil(t, err)
	err = client.WriteDirectoryEnd()
	tt.AssertIsNil(t, err)

	// third folder c
	err = client.WriteDirectoryStart("0755", "c")
	tt.AssertIsNil(t, err)
	// in folder c, empty file d
	err = client.WriteFile("0644", 0, "d", strings.NewReader(""))
	tt.AssertIsNil(t, err)
	err = client.WriteDirectoryEnd()
	tt.AssertIsNil(t, err)

	// last file e (Content: root)
	err = client.WriteFile("0644", 4, "e", strings.NewReader("root"))
	tt.AssertIsNil(t, err)

	// Write End, Now Check

	tt.AssertTrue(t, checkFileExist(t, "/test/a"))
	tt.AssertTrue(t, checkFileExist(t, "/test/a/b"))
	tt.AssertTrue(t, checkFileExist(t, "/test/a/c"))
	tt.AssertTrue(t, checkFileExist(t, "/test/b"))
	tt.AssertTrue(t, checkFileExist(t, "/test/c"))
	tt.AssertTrue(t, checkFileExist(t, "/test/c/d"))
	tt.AssertTrue(t, checkFileExist(t, "/test/e"))

	tt.AssertEqual(t, "hahaha", string(readFile(t, "/test/a/b")))
	tt.AssertEqual(t, "xixixixi", string(readFile(t, "/test/a/c")))
	tt.AssertEqual(t, "", string(readFile(t, "/test/c/d")))
	tt.AssertEqual(t, "root", string(readFile(t, "/test/e")))
}
