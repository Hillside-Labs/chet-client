package lib

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

var DefaultMatches = map[string]*regexp.Regexp{
	"go_test":  regexp.MustCompile("(.*)/?go test(.*)"),
	"go_build": regexp.MustCompile("(.*)/?go build(.*)"),
}

type CmdMatcher struct {
	matchers map[string]*regexp.Regexp
}

func (cm CmdMatcher) FindMatch(cmd string) string {
	for name, m := range cm.matchers {
		if m.MatchString(cmd) {
			return name
		}
	}

	// no match so we take the first arg.
	first := strings.Split(cmd, " ")
	relativeName := strings.Split(first[0], "/")
	return relativeName[len(relativeName)-1]
}

func NewCmdMatcher(matchers map[string]*regexp.Regexp) CmdMatcher {
	return CmdMatcher{matchers: matchers}
}

// CreateLabelFromCommand takes a command and distills it into a label
// that we measure. We want to know the command and the package it
// pertains to.
func CreateLabelFromCommand(cmd *exec.Cmd, cm CmdMatcher) (string, error) {
	if len(cmd.Args) == 0 {
		return "", errors.New("no command args")
	}

	command := strings.Join(cmd.Args, " ")
	return cm.FindMatch(command), nil
}
