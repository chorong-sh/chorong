package lint_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var (
	shaPinRE = regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+(?:/[a-zA-Z0-9_./-]+)?@[0-9a-f]{40}$`)
	semverRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

func readRepoFile(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return data
}

func TestReleasePleaseManifestValid(t *testing.T) {
	var manifest map[string]string
	if err := json.Unmarshal(readRepoFile(t, ".release-please-manifest.json"), &manifest); err != nil {
		t.Fatalf("parse .release-please-manifest.json: %v", err)
	}
	ver, ok := manifest["."]
	if !ok || !semverRE.MatchString(ver) {
		t.Fatalf("invalid root version in .release-please-manifest.json: %q", ver)
	}
}

func TestReleasePleaseConfigValid(t *testing.T) {
	var cfg struct {
		ReleaseType         string         `json:"release-type"`
		BumpMinorPreMajor   bool           `json:"bump-minor-pre-major"`
		IncludeComponentTag bool           `json:"include-component-in-tag"`
		IncludeVInTag       bool           `json:"include-v-in-tag"`
		Packages            map[string]any `json:"packages"`
		ChangelogSections   []struct {
			Type    string `json:"type"`
			Section string `json:"section"`
			Hidden  bool   `json:"hidden"`
		} `json:"changelog-sections"`
	}
	if err := json.Unmarshal(readRepoFile(t, "release-please-config.json"), &cfg); err != nil {
		t.Fatalf("parse release-please-config.json: %v", err)
	}
	if cfg.ReleaseType != "go" {
		t.Errorf("release-type = %q, want go", cfg.ReleaseType)
	}
	if !cfg.BumpMinorPreMajor {
		t.Error("bump-minor-pre-major must be true")
	}
	if cfg.IncludeComponentTag {
		t.Error("include-component-in-tag must be false")
	}
	if !cfg.IncludeVInTag {
		t.Error("include-v-in-tag must be true")
	}
	if _, ok := cfg.Packages["."]; !ok || len(cfg.Packages) != 1 {
		t.Errorf("packages = %v, want single root package", cfg.Packages)
	}
	hidden := map[string]bool{}
	for _, s := range cfg.ChangelogSections {
		hidden[s.Type] = s.Hidden
	}
	for _, vis := range []string{"feat", "fix", "perf", "docs"} {
		if hidden[vis] {
			t.Errorf("section %q should be visible", vis)
		}
	}
	for _, hid := range []string{"build", "ci", "chore", "refactor", "style", "test"} {
		if !hidden[hid] {
			t.Errorf("section %q should be hidden", hid)
		}
	}
}

func TestChangelogHasHeaderAndBaselineRelease(t *testing.T) {
	changelog := string(readRepoFile(t, "CHANGELOG.md"))
	if !strings.HasPrefix(changelog, "# Changelog\n") {
		t.Error("CHANGELOG.md must start with '# Changelog\\n'")
	}
	if !strings.Contains(changelog, "0.1.0") {
		t.Error("CHANGELOG.md missing 0.1.0 baseline entry")
	}
}

func TestGoReleaserConfigValidYAML(t *testing.T) {
	var cfg struct {
		Version int `yaml:"version"`
		Builds  []struct {
			ID      string   `yaml:"id"`
			Main    string   `yaml:"main"`
			Binary  string   `yaml:"binary"`
			Env     []string `yaml:"env"`
			LDFlags []string `yaml:"ldflags"`
		} `yaml:"builds"`
		Brews []struct {
			Name       string `yaml:"name"`
			Repository struct {
				Owner string `yaml:"owner"`
				Name  string `yaml:"name"`
			} `yaml:"repository"`
		} `yaml:"brews"`
	}
	if err := yaml.Unmarshal(readRepoFile(t, ".goreleaser.yaml"), &cfg); err != nil {
		t.Fatalf("parse .goreleaser.yaml: %v", err)
	}
	if cfg.Version != 2 {
		t.Fatalf("goreleaser version = %d, want 2", cfg.Version)
	}
	if len(cfg.Builds) == 0 {
		t.Fatal("goreleaser config has no builds")
	}
	for _, b := range cfg.Builds {
		if b.ID == "" || b.Main == "" || b.Binary == "" {
			t.Errorf("incomplete build entry: %+v", b)
		}
		if len(b.LDFlags) == 0 {
			t.Errorf("build %s missing ldflags", b.ID)
		}
	}
	if len(cfg.Brews) == 0 {
		t.Fatal("goreleaser config has no brews")
	}
	for _, br := range cfg.Brews {
		if br.Repository.Owner != "chorong-sh" || br.Repository.Name != "homebrew-tap" {
			t.Errorf("brew %s targets %s/%s, want chorong-sh/homebrew-tap",
				br.Name, br.Repository.Owner, br.Repository.Name)
		}
	}
}

// TestAllGitHubWorkflowsPinActionSHAs enforces 40-char commit SHA pinning
// across every workflow in .github/workflows/*.yml.
func TestAllGitHubWorkflowsPinActionSHAs(t *testing.T) {
	root := repoRoot(t)
	matches, err := filepath.Glob(filepath.Join(root, ".github/workflows/*.yml"))
	if err != nil {
		t.Fatalf("glob workflows: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("no workflow files found under .github/workflows/")
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var doc map[string]any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		jobs, _ := doc["jobs"].(map[string]any)
		for jobName, rawJob := range jobs {
			job, _ := rawJob.(map[string]any)
			if uses, ok := job["uses"].(string); ok {
				if !shaPinRE.MatchString(uses) {
					t.Errorf("%s job %q uses unpinned reusable workflow: %s", filepath.Base(path), jobName, uses)
				}
			}
			steps, _ := job["steps"].([]any)
			for _, rawStep := range steps {
				step, _ := rawStep.(map[string]any)
				if uses, ok := step["uses"].(string); ok {
					if !shaPinRE.MatchString(uses) {
						t.Errorf("%s job %q uses unpinned action: %s", filepath.Base(path), jobName, uses)
					}
				}
			}
		}
	}
}
