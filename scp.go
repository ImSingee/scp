package scp

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
)

// Copy: Copy `from` to `target`
func Copy(session *ssh.Session, from, target string) error {
	stat, err := os.Stat(from)
	if err != nil {
		return fmt.Errorf("cannot stat file %s: %w", from, err)
	}

	if stat.IsDir() {
		return CopyFolder(session, from, target)
	}

	f, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", from, err)
	}
	defer f.Close()

	return CopyFile(session, f, stat.Name(), stat.Size(), stat.Mode(), target)
}

func CopyFile(session *ssh.Session, file io.Reader, filename string, size int64, mode os.FileMode, target string) error {
	client, err := NewClient(session)
	if err != nil {
		return err
	}

	err = client.Start(target, false)
	if err != nil {
		return err
	}

	err = client.WriteFile(ConvertFileModeToPermString(mode), size, filename, file)
	if err != nil {
		return err
	}

	return nil
}

func CopyFolder(session *ssh.Session, from, target string) error {
	return nil
}
