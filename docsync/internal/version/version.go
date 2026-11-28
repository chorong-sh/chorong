// Package version holds the docsync version string.
package version

import pkgversion "github.com/chorong-sh/chorong/pkg/version"

// Version is the docsync version. Overridable at link time with
// -X github.com/chorong-sh/chorong/docsync/internal/version.Version.
var Version = "" //nolint:pkgvar // set with -ldflags -X

// Resolve returns the effective docsync version string.
func Resolve() string {
	return pkgversion.ResolveWithOverride(Version)
}
