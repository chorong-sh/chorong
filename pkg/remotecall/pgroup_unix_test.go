//go:build unix

package remotecall

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// sessionHelperEnv makes the test binary print whether it leads its own
// session, which means it has no controlling terminal.
const sessionHelperEnv = "REMOTECALL_SESSION_HELPER"

func TestMain(m *testing.M) {
	if os.Getenv(sessionHelperEnv) == "1" {
		sid, err := unix.Getsid(0)
		fmt.Println(err == nil && sid == os.Getpid())
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestRunIONoTerminalStartsANewSession(t *testing.T) {
	env := append(os.Environ(), sessionHelperEnv+"=1")
	var stdout, stderr strings.Builder
	res, err := RunIOEnvNoTerminal(context.Background(), env, nil, &stdout, &stderr, os.Args[0])
	if err != nil || res.Code != 0 {
		t.Fatalf("RunIOEnvNoTerminal = %+v, %v; stderr %q", res, err, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "true" {
		t.Fatalf("child leads its own session = %s, want true", got)
	}

	stdout.Reset()
	if _, err := RunIOEnv(context.Background(), env, nil, &stdout, &stderr, os.Args[0]); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "false" {
		t.Fatalf("RunIOEnv child leads its own session = %s, want false", got)
	}
}

func TestStartWithETXTBSYRetry(t *testing.T) {
	t.Run("retries transient ETXTBSY then succeeds", func(t *testing.T) {
		attempts := 0
		got, err := startWithETXTBSYRetry(context.Background(), func() (string, error) {
			attempts++
			if attempts < 3 {
				return "", fmt.Errorf("fork/exec: %w", unix.ETXTBSY)
			}
			return "ok", nil
		})
		if err != nil || got != "ok" || attempts != 3 {
			t.Fatalf("got=(%q, %v) attempts=%d, want (ok, nil) attempts=3", got, err, attempts)
		}
	})

	t.Run("does not retry non-ETXTBSY errors", func(t *testing.T) {
		attempts := 0
		boom := errors.New("boom")
		_, err := startWithETXTBSYRetry(context.Background(), func() (string, error) {
			attempts++
			return "", boom
		})
		if !errors.Is(err, boom) || attempts != 1 {
			t.Fatalf("err=%v attempts=%d, want boom attempts=1", err, attempts)
		}
	})

	t.Run("stops when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0
		_, err := startWithETXTBSYRetry(ctx, func() (string, error) {
			attempts++
			cancel()
			return "", unix.ETXTBSY
		})
		if !errors.Is(err, unix.ETXTBSY) || attempts != 1 {
			t.Fatalf("err=%v attempts=%d, want ETXTBSY attempts=1", err, attempts)
		}
	})
}
