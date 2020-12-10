package scp

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/alessio/shellescape"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strconv"
	"strings"
)

type RemoteClient struct {
	session *ssh.Session

	stdout io.Reader
	stdin  io.WriteCloser

	verbose bool
}

func NewClient(session *ssh.Session) (*RemoteClient, error) {
	stdout, err := session.StdoutPipe()

	if err != nil {
		return nil, fmt.Errorf("cannot catch stdout: %w", err)
	}

	stdin, err := session.StdinPipe()

	if err != nil {
		return nil, fmt.Errorf("cannot set stdin: %w", err)
	}

	return &RemoteClient{session: session, stdout: stdout, stdin: stdin}, nil
}

func (c *RemoteClient) Start(filename string, isDirectory bool) error {
	commandBuilder := strings.Builder{}

	commandBuilder.WriteString("scp -t")

	if isDirectory {
		commandBuilder.WriteString(" -r")
	}
	if c.verbose {
		commandBuilder.WriteString(" -v")

		// just for debug
		stderr, err := c.session.StderrPipe()
		if err != nil {
			fmt.Println("Cannot connect to stderr")
		} else {
			f, err := os.Create("protocol.log")

			if err != nil {
				fmt.Println("Cannot create test file")
			} else {
				go func() {
					buffer := make([]byte, 1024)
					for {
						read, _ := stderr.Read(buffer)
						if read != 0 {
							_, err := f.Write(buffer[:read])

							if err != nil {
								fmt.Println("Cannot write verbose info")
							}
						}
					}
				}()
			}

		}
	}

	commandBuilder.WriteByte(' ')

	commandBuilder.WriteString(shellescape.Quote(filename))

	command := commandBuilder.String()

	err := c.session.Start(command)

	if err != nil {
		return err
	}

	return c.checkResponse()
}

func (c *RemoteClient) checkResponse() error {
	buffer := make([]byte, 1)

	_, err := c.stdout.Read(buffer)
	if err != nil {
		return fmt.Errorf("cannot read from remote stdout: %w", err)
	}

	if buffer[0] == 0 {
		return nil
	}

	errorMessageBuilder := strings.Builder{}

	if buffer[0] == 1 {
		errorMessageBuilder.WriteString("Error (1): ")
	} else if buffer[0] == 2 {
		errorMessageBuilder.WriteString("Fatal (2): ")
	} else {
		errorMessageBuilder.WriteString("UnknownError (")
		errorMessageBuilder.WriteString(strconv.Itoa(int(buffer[0])))
		errorMessageBuilder.WriteString("): ")
	}

	reader := bufio.NewReader(c.stdout)
	all, err := reader.ReadString('\n')

	if err != nil {
		errorMessageBuilder.WriteString("[ERROR during get reason] (error: ")
		errorMessageBuilder.WriteString(err.Error())
		errorMessageBuilder.WriteString(")")
	} else {
		errorMessageBuilder.WriteString(all)
	}

	return errors.New(errorMessageBuilder.String())
}

func (c *RemoteClient) WriteFile(perm string, size int64, filename string, data io.Reader) error {
	_, err := fmt.Fprintln(c.stdin, "C"+perm, size, filename)
	if err != nil {
		return err
	}

	err = c.checkResponse()
	if err != nil {
		return err
	}

	_, err = io.Copy(c.stdin, data)
	if err != nil {
		return err
	}

	_, err = c.stdin.Write([]byte{0})
	if err != nil {
		return err
	}

	err = c.checkResponse()
	if err != nil {
		return err
	}

	return nil
}

func (c *RemoteClient) WriteDirectoryStart(perm string, filename string) error {
	if len(filename) > 1024 {
		// Unknown reason, just as same as https://github.com/openssh/openssh-portable/blob/master/scp.c#L1204
		filename = filename[:1024]
	}

	_, err := fmt.Fprintln(c.stdin, "D"+perm, 0, filename)
	if err != nil {
		return err
	}

	err = c.checkResponse()
	if err != nil {
		return err
	}

	return nil
}

func (c *RemoteClient) WriteDirectoryEnd() error {
	_, err := fmt.Fprintln(c.stdin, "E")
	if err != nil {
		return err
	}

	err = c.checkResponse()
	if err != nil {
		return err
	}

	return nil
}
