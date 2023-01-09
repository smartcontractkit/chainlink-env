package client

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func ExecCmd(command string) error {
	c := strings.Split(command, " ")
	cmd := exec.Command(c[0], c[1:]...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go readStdPipe(stderr)
	go readStdPipe(stdout)
	return cmd.Wait()
}

// readStdPipe continuously read a pipe from the command
func readStdPipe(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
}
