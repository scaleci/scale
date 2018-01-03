package exec

import (
	"bufio"
	"fmt"
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

	stdOutScanner := bufio.NewScanner(stdOutReader)
	streamOutput(logPrefix, stdOutScanner)

	stdErrScanner := bufio.NewScanner(stdErrReader)
	streamOutput(logPrefix, stdErrScanner)

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

func streamOutput(logPrefix string, scanner *bufio.Scanner) {
	go func() {
		for scanner.Scan() {
			fmt.Printf("[%s] %s\n", logPrefix, scanner.Text())
		}
	}()
}
