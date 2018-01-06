package exec

import (
	"io"
	"os"
	"os/exec"
)

func Run(cmdName string, cmdArgs []string, logPrefix string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	stdOutReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdErrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	streamOutput(logPrefix, stdOutReader)
	streamOutput(logPrefix, stdErrReader)

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func streamOutput(logPrefix string, reader io.Reader) {
	go func() {
		buf := make([]byte, 1024, 1024)
		for {
			n, err := reader.Read(buf[:])
			if n > 0 {
				d := buf[:n]

				_, err := os.Stdout.Write(d)
				if err != nil {
					return
				}
			}
			if err != nil {
				// Read returns io.EOF at the end of file, which is not an error for us
				if err == io.EOF {
					err = nil
				}
				return
			}
		}
	}()
}
