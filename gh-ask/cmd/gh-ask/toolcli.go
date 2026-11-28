package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	ghask "github.com/chorong-sh/chorong/gh-ask"
	"github.com/chorong-sh/chorong/pkg/toolcli"
)

func newGHAskTool() toolcli.Tool {
	return toolcli.Tool{
		Name:        "gh-ask",
		MainPackage: "github.com/chorong-sh/chorong/gh-ask/cmd/gh-ask",
		SkillText:   skillText,
		ConfigPath: func() (string, bool) {
			p, err := ghask.DefaultConfigPath(os.Environ())
			return p, err == nil
		},
		CheckConfig: func() error {
			p, err := ghask.DefaultConfigPath(os.Environ())
			if err != nil {
				return fmt.Errorf("resolve gh-ask config: %w", err)
			}
			if _, err := ghask.LoadConfig(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("load gh-ask config: %w", err)
			}
			return nil
		},
	}
}
