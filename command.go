package main

import (
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

func executeCommand(command string) (*string, error) {
	trimmed := strings.TrimSpace(command)
	parts := strings.Split(trimmed, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Errorf("%s: %v, %s", command, err.Error(), string(out))
	}
	stdout := string(out)
	return &stdout, nil
}

func (b *backup) executePreCommand() (*string, error) {
	return executeCommand(b.PreCommand)
}

func (b *backup) executePostCommand() (*string, error) {
	return executeCommand(b.PostCommand)
}
