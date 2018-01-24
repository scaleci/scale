package exec

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
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

// Capture output:
// https://stackoverflow.com/questions/10385551/get-exit-code-go
func RunAndCaptureOutput(cmdName string, cmdArgs []string, outbuf *bytes.Buffer, errbuf *bytes.Buffer) int {
	// default failure code is 1
	exitCode := 1

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	err := cmd.Run()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			log.Printf("Could not get exit code for failed program: %v, %v", cmdName, cmdArgs)
			if errbuf.String() == "" {
				errbuf = bytes.NewBufferString(err.Error())
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	return exitCode
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
