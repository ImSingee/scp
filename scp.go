package scp

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Copy `from` to `target`
func Copy(session *ssh.Session, from, target string) error {
	return DefaultOptions.Copy(session, from, target)
}

func CopyFile(session *ssh.Session, file io.Reader, filename string, size int64, mode os.FileMode, target string) error {
	return DefaultOptions.CopyFile(session, file, filename, size, mode, target)
}

func CopyFolder(session *ssh.Session, from string, mode os.FileMode, target string) error {
	return DefaultOptions.CopyFolder(session, from, mode, target)
}

// Copy `from` to `target`
func (o *Options) Copy(session *ssh.Session, from, target string) error {
	stat, err := os.Stat(from)
	if err != nil {
		return fmt.Errorf("cannot stat file %s: %w", from, err)
	}

	if stat.IsDir() {
		return o.CopyFolder(session, from, stat.Mode(), target)
	}

	f, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", from, err)
	}
	defer f.Close()

	return o.CopyFile(session, f, stat.Name(), stat.Size(), stat.Mode(), target)
}

func (o *Options) CopyFile(session *ssh.Session, file io.Reader, filename string, size int64, mode os.FileMode, target string) error {
	client, err := o.NewClient(session)
	if err != nil {
		return err
	}

	err = client.Start(target, false)
	if err != nil {
		return err
	}

	return o.copyFile(client, ConvertFileModeToPermString(mode), size, filename, file)
}

func (o *Options) copyFile(client *RemoteClient, perm string, size int64, filename string, file io.Reader) error {
	err := client.WriteFile(perm, size, filename, file)
	if err != nil {
		return err
	}

	return nil
}

func (o *Options) CopyFolder(session *ssh.Session, from string, mode os.FileMode, target string) error {
	client, err := o.NewClient(session)
	if err != nil {
		return err
	}

	err = client.Start(target, true)
	if err != nil {
		return err
	}

	return o.copyFolder(client, ConvertFileModeToPermString(mode), from)
}

func (o *Options) copyFolder(client *RemoteClient, perm string, path string) error {
	err := client.WriteDirectoryStart(perm, filepath.Base(path))
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(path)

	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = o.copyFolder(client, ConvertFileModeToPermString(file.Mode()), filepath.Join(path, file.Name()))

			if err != nil {
				return err
			}
		} else {
			f, err := os.Open(filepath.Join(path, file.Name()))
			if err != nil {
				return err
			}

			err = o.copyFile(client, ConvertFileModeToPermString(file.Mode()), file.Size(), file.Name(), f)
			f.Close()

			if err != nil {
				return err
			}
		}
	}

	err = client.WriteDirectoryEnd()
	if err != nil {
		return err
	}

	return nil
}
