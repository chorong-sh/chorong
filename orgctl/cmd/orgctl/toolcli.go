package main

import (
	"github.com/chorong-sh/chorong/pkg/toolcli"
)

func newOrgctlTool() toolcli.Tool {
	return toolcli.Tool{
		Name:        "orgctl",
		MainPackage: "github.com/chorong-sh/chorong/orgctl/cmd/orgctl",
		SkillText:   skillText,
	}
}
