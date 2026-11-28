package main

import (
	"github.com/chorong-sh/chorong/pkg/toolcli"
)

func newGHStorageTool() toolcli.Tool {
	return toolcli.Tool{
		Name:        "gh-storage",
		MainPackage: "github.com/chorong-sh/chorong/gh-nudge/cmd/gh-storage",
		SkillText:   skillText,
	}
}
