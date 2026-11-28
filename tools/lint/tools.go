//go:build tools

// Package lint pins Go static analyzer modules used by Bazel nogo so go.mod
// retains their transitive dependencies for Gazelle's go_deps extension.
package lint

import (
	_ "github.com/gordonklaus/ineffassign/pkg/ineffassign"
	_ "github.com/jaeyeom/godernize/oserrors"
	_ "github.com/securego/gosec/v2/goanalysis"
	_ "github.com/uudashr/iface/identical"
	_ "github.com/uudashr/iface/opaque"
	_ "github.com/uudashr/iface/unexported"
	_ "github.com/uudashr/iface/unused"
	_ "golang.org/x/tools/go/analysis/passes/unusedwrite"
)
