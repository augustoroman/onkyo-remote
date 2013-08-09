package main

import (
	"fmt"
	"github.com/augustoroman/onkyo-remote"
)

type Value struct {
	Code        string
	Name        string
	Description string
}

type CommandInfo struct {
	Zone        string
	Code        string
	Name        string
	Description string
	Values      []Value
}

var codeToCmd = map[string]*CommandInfo{}

func init() {
	for i := range AllKnownCommands {
		cmd := &AllKnownCommands[i]
		codeToCmd[cmd.Code] = cmd
	}
}

func MessageToCommand(m eiscp.Message) (*CommandInfo, *Value) {
	cmd := codeToCmd[m.Code()]
	if cmd == nil {
		return nil, nil
	}
	for i := range cmd.Values {
		if cmd.Values[i].Code == m.Value() {
			return cmd, &cmd.Values[i]
		}
	}
	return cmd, nil
}

func pretty(m eiscp.Message) string {
	cmd, val := MessageToCommand(m)
	if cmd == nil {
		return fmt.Sprintf("Unknown code: [%s] Val: [%s]", m.Code(), m.Value())
	}
	if val == nil {
		s := ""
		for i := range cmd.Values {
			c := '!'
			if cmd.Values[i].Code == m.Value() {
				c = '+'
			}
			s += fmt.Sprintf("%c%s ", c, cmd.Values[i].Code)
		}
		return fmt.Sprintf("%s (%s): Unknown val: [%q] : %s", cmd.Name, cmd.Code, m.Value(), s)
	}
	return fmt.Sprintf("%s: %s    (%s: %s)", cmd.Name, val.Name, cmd.Code, val.Code)
}
