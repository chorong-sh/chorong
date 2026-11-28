package main

import (
	"github.com/chorong-sh/chorong/pkg/toolcli"
)

func newGHCodeownersTool() toolcli.Tool {
	return toolcli.Tool{
		Name:        "gh-codeowners",
		MainPackage: "github.com/chorong-sh/chorong/gh-nudge/cmd/gh-codeowners",
		SkillText:   skillText,
	}
}
