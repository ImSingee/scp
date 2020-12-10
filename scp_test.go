package scp

import (
	"fmt"
	"github.com/ImSingee/tt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func generateTempFile(t *testing.T, data string) string {
	p := filepath.Join(os.TempDir(), fmt.Sprintf("scp-test-%d", time.Now().UnixNano()))

	f, err := os.Create(p)
	tt.AssertIsNil(t, err)
	defer f.Close()

	_, err = f.WriteString(data)
	tt.AssertIsNil(t, err)

	return p
}

func TestCopyFileToFile(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	p := generateTempFile(t, "abcdefg")

	err := Copy(session, p, "/test/test01")
	tt.AssertIsNil(t, err)

	tt.AssertTrue(t, checkFileExist(t, "/test/test01"))
	tt.AssertEqual(t, "abcdefg", string(readFile(t, "/test/test01")))
}

func TestCopyFileToFolder(t *testing.T) {
	reset(t)

	session := getSshSession(t)
	defer session.Close()

	p := generateTempFile(t, "abcdefg")

	err := Copy(session, p, "/test")
	tt.AssertIsNil(t, err)

	remotePath := fmt.Sprintf("/test/%s", filepath.Base(p))

	tt.AssertTrue(t, checkFileExist(t, remotePath))
	tt.AssertEqual(t, "abcdefg", string(readFile(t, remotePath)))
}
