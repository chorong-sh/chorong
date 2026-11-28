package lint_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var (
	skipDirs = map[string]bool{
		".git":         true,
		"vendor":       true,
		".cache":       true,
		".claude":      true,
		".tmp":         true,
		"node_modules": true,
		"dist":         true,
	}
	goRuleStartRE = regexp.MustCompile(`(?m)^go_[a-z_]+\s*\(`)
	quotedGoRE    = regexp.MustCompile(`"(:?[^"\n]+\.go)"`)
	globCallRE    = regexp.MustCompile(`glob\s*\(([^)]*)\)`)
	quotedStrRE   = regexp.MustCompile(`"([^"\n]+)"`)
)

func shouldSkipDir(name string) bool {
	return skipDirs[name] || strings.HasPrefix(name, "bazel-") || strings.HasPrefix(name, ".bazel-")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	cur, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		cand := filepath.Join(cur, "go.mod")
		if _, err := os.Stat(cand); err == nil {
			resolved, err := filepath.EvalSymlinks(cand)
			if err != nil {
				t.Fatalf("eval symlinks %s: %v", cand, err)
			}
			return filepath.Dir(resolved)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			t.Fatal("go.mod not found")
		}
		cur = parent
	}
}

func findGoFiles(t *testing.T, root string) map[string]bool {
	t.Helper()
	found := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasPrefix(name, "test_") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		// tools/lint/tools.go uses //go:build tools solely to pin nogo deps in go.mod.
		if rel == "tools/lint/tools.go" {
			return nil
		}
		found[rel] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return found
}

func extractBalancedBlock(src string, openParenIdx int) string {
	depth := 0
	inStr := false
	escaped := false
	for i := openParenIdx; i < len(src); i++ {
		c := src[i]
		if inStr {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return src[openParenIdx : i+1]
			}
		}
	}
	return src[openParenIdx:]
}

func findBazelGoFiles(t *testing.T, root string) map[string]bool {
	t.Helper()
	rootFS, err := os.OpenRoot(root)
	if err != nil {
		t.Fatalf("open root %s: %v", root, err)
	}
	defer rootFS.Close()

	found := map[string]bool{}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "BUILD.bazel" && d.Name() != "BUILD" {
			return nil
		}
		relBuild, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := rootFS.ReadFile(relBuild)
		if err != nil {
			return err
		}
		pkgDir := filepath.Dir(path)
		relDir, err := filepath.Rel(root, pkgDir)
		if err != nil {
			return err
		}
		content := string(data)
		for _, loc := range goRuleStartRE.FindAllStringIndex(content, -1) {
			openIdx := strings.IndexByte(content[loc[0]:], '(')
			if openIdx < 0 {
				continue
			}
			block := extractBalancedBlock(content, loc[0]+openIdx)
			for _, m := range quotedGoRE.FindAllStringSubmatch(block, -1) {
				src := strings.TrimPrefix(m[1], ":")
				rel := src
				if relDir != "." {
					rel = filepath.ToSlash(filepath.Join(relDir, src))
				}
				found[rel] = true
			}
			for _, gm := range globCallRE.FindAllStringSubmatch(block, -1) {
				for _, pm := range quotedStrRE.FindAllStringSubmatch(gm[1], -1) {
					matches, _ := filepath.Glob(filepath.Join(pkgDir, pm[1]))
					for _, match := range matches {
						if strings.HasSuffix(match, ".go") {
							if r, err := filepath.Rel(root, match); err == nil {
								found[filepath.ToSlash(r)] = true
							}
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk BUILD files: %v", err)
	}
	return found
}

// TestRepoGoFilesInBazel verifies every checked .go file in the checkout is
// listed in a BUILD.bazel go_* target.
func TestRepoGoFilesInBazel(t *testing.T) {
	root := repoRoot(t)
	fsFiles := findGoFiles(t, root)
	if len(fsFiles) == 0 {
		t.Fatalf("no .go files found under %s", root)
	}
	bazelFiles := findBazelGoFiles(t, root)
	var missing []string
	for f := range fsFiles {
		if !bazelFiles[f] {
			missing = append(missing, f)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("Go files not listed in any BUILD.bazel go_* target (run `bazel run //:gazelle` or add them to BUILD.bazel):\n  - %s",
			strings.Join(missing, "\n  - "))
	}
}
