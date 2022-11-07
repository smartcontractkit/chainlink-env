package client

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func ExecCmd(command string) error {
	c := strings.Split(command, " ")
	cmd := exec.Command(c[0], c[1:]...)
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	scanner2 := bufio.NewScanner(stdout)
	scanner2.Split(bufio.ScanLines)
	for scanner2.Scan() {
		m := scanner2.Text()
		fmt.Println(m)
	}
	return cmd.Wait()
}
