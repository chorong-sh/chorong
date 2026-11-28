package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chorong-sh/chorong/gh-nudge/internal/config"
	"github.com/chorong-sh/chorong/pkg/toolcli"
)

func defaultGHSlackConfigPath() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	base := filepath.Join(home, ".config", "gh-nudge", "config")
	if _, err := os.Stat(base + ".pkl"); err == nil {
		return base + ".pkl", true
	}
	return base + ".yaml", true
}

func newGHSlackTool() toolcli.Tool {
	return toolcli.Tool{
		Name:        "gh-slack",
		MainPackage: "github.com/chorong-sh/chorong/gh-nudge/cmd/gh-slack",
		SkillText:   skillText,
		ConfigPath:  defaultGHSlackConfigPath,
		CheckConfig: func() error {
			if _, err := config.LoadConfig(""); err != nil {
				return fmt.Errorf("load gh-nudge config: %w", err)
			}
			return nil
		},
	}
}
