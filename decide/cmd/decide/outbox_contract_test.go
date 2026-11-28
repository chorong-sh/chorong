package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chorong-sh/chorong/decide"
)

// TestOutboxAskKeepsOptions runs the real decide CLI the way an approval
// outbox asks (pull delivery, owner-replies-only, explicit options), then
// verifies a second sender's plain reply is rejected without clearing options,
// and that the owner can re-attach options via decide reply.
func TestOutboxAskKeepsOptions(t *testing.T) {
	state := t.TempDir()
	cfgDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("DECIDE_TEST_CHILD", "1")

	const owner = "outbox-bot"
	stdout, stderr, code := runDecideChildAsSender(t, owner,
		"ask",
		"--class", "approval",
		"--kind", "relay",
		"--verbatim", "Ship the note.",
		"--owner", owner,
		"--notify-pane", decide.SchemePull,
		"--owner-replies-only",
		"--option", "yes:Yes",
		"--option", "no:No",
		"--option", "edit:Edit",
		"Post this note as written?",
	)
	if code != 0 {
		t.Fatalf("ask failed (%d):\n%s\n%s", code, stdout, stderr)
	}
	id := strings.TrimSpace(stdout)
	item := loadDecide(t, state, id)
	if !item.OwnerRepliesOnly {
		t.Fatal("ask did not set ownerRepliesOnly")
	}
	if item.NotifyPane != decide.SchemePull {
		t.Fatalf("notifyPane = %q, want %q", item.NotifyPane, decide.SchemePull)
	}
	if !optionKeys(item.Options, "yes", "no", "edit") {
		t.Fatalf("options = %+v", item.Options)
	}

	stdout, stderr, code = runDecideChildAsSender(t, "other-agent", "reply", id, "plain reply from someone else")
	if code == 0 {
		t.Fatalf("second sender was accepted\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "accepts replies only from its owner "+owner) {
		t.Fatalf("stderr = %q", stderr)
	}
	if got := loadDecide(t, state, id); !optionKeys(got.Options, "yes", "no", "edit") || len(got.Messages) != 1 {
		t.Fatalf("second sender changed the item: options %+v messages %d", got.Options, len(got.Messages))
	}

	clearDecideOptions(t, state, id)
	stdout, stderr, code = runDecideChildAsSender(t, owner,
		"reply",
		"--option", "yes:Yes",
		"--option", "no:No",
		"--option", "edit:Edit",
		id,
		"Re-attached approval options.",
	)
	if code != 0 {
		t.Fatalf("owner re-attach reply failed (%d):\n%s\n%s", code, stdout, stderr)
	}
	got := loadDecide(t, state, id)
	if !optionKeys(got.Options, "yes", "no", "edit") {
		t.Fatalf("options after re-attach = %+v", got.Options)
	}
	if len(got.Messages) != 2 {
		t.Fatalf("messages = %d, want the ask and one re-attach", len(got.Messages))
	}
	last := got.Messages[len(got.Messages)-1]
	if last.By != owner {
		t.Fatalf("re-attach by = %q, want %q", last.By, owner)
	}
}

func TestOutboxQuietDoneClosesWithoutUnread(t *testing.T) {
	state := t.TempDir()
	cfgDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("DECIDE_TEST_CHILD", "1")

	const owner = "outbox-bot"
	stdout, stderr, code := runDecideChildAsSender(t, owner,
		"ask",
		"--class", "approval",
		"--kind", "relay",
		"--verbatim", "Ship the note.",
		"--owner", owner,
		"--notify-pane", decide.SchemePull,
		"--option", "yes:Yes",
		"--option", "no:No",
		"Post this as written?",
	)
	if code != 0 {
		t.Fatalf("ask failed (%d):\n%s\n%s", code, stdout, stderr)
	}
	id := strings.TrimSpace(stdout)

	stdout, stderr, code = runDecideChildAsSender(t, owner, "reply", id, "Post stuck.")
	if code != 0 {
		t.Fatalf("reply failed (%d):\n%s\n%s", code, stdout, stderr)
	}
	before := loadDecide(t, state, id)

	const outcome = "Posted.\n\n- https://example.com/posts/1"
	stdout, stderr, code = runDecideChildAsSender(t, owner, "done", "--quiet", id, outcome)
	if code != 0 {
		t.Fatalf("done --quiet failed (%d):\n%s\n%s", code, stdout, stderr)
	}
	got := loadDecide(t, state, id)
	if got.State != decide.StateDone || got.Unread || got.Outcome != outcome {
		t.Fatalf("item = state %s unread %v outcome %q", got.State, got.Unread, got.Outcome)
	}
	if len(got.Messages) != len(before.Messages) {
		t.Fatalf("messages = %d, want %d (no extra message)", len(got.Messages), len(before.Messages))
	}
}

func loadDecide(t *testing.T, state, id string) decide.Item {
	t.Helper()
	led, err := decide.Open(filepath.Join(state, "decide"))
	if err != nil {
		t.Fatal(err)
	}
	item, err := led.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func clearDecideOptions(t *testing.T, state, id string) {
	t.Helper()
	path := filepath.Join(state, "decide", "ledger.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Next  int           `json:"next"`
		Items []decide.Item `json:"items"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	var found bool
	for i := range doc.Items {
		if doc.Items[i].ID == id {
			doc.Items[i].Options = nil
			found = true
		}
	}
	if !found {
		t.Fatalf("ledger has no %s", id)
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

// runDecideChildAsSender runs the decide CLI as a named agentbus sender.
func runDecideChildAsSender(t *testing.T, sender string, args ...string) (string, string, int) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "agentbus")
	body := "#!/bin/sh\nprintf '%s\\nwhen: other\\n' " + sender + "\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil { //nolint:gosec // G306: the fake agentbus must be executable.
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], args...) //nolint:gosec // G204: the test re-execs itself; no shell.
	cmd.Env = append(withoutEnv("HERDR_PANE_ID", "DECIDE_AS_HANDLE", "SSH_CONNECTION", "DECIDE_BUS_NAMES", "DECIDE_TAB_LABEL", "PATH"),
		"HERDR_PANE_ID=w9:p9",
		"PATH="+dir,
	)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

func withoutEnv(keys ...string) []string {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	var env []string
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if drop[key] {
			continue
		}
		env = append(env, kv)
	}
	return env
}
