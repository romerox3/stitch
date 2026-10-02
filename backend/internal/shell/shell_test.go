package shell

import (
	"runtime"
	"testing"
)

// TestRunCmdPrependsOSAppropriatePaths verifies that RunCmd prepends the
// Homebrew paths on macOS and the conventional Linux locations elsewhere,
// while preserving any PATH already present in the environment (appended
// after the prepended entries instead of being overwritten).
func TestRunCmdPrependsOSAppropriatePaths(t *testing.T) {
	// Keep /usr/bin reachable so "sh" itself stays resolvable, but lead with
	// a distinctive entry to prove the original PATH is preserved verbatim.
	const orig = "/custom/bin:/usr/bin:/bin"
	t.Setenv("PATH", orig)

	res := RunCmd("sh -c 'echo $PATH'", "")
	if !res.Success {
		t.Fatalf("command failed: %s", res.Stderr)
	}

	want := "/usr/local/bin:/usr/bin:/bin:" + orig
	if runtime.GOOS == "darwin" {
		want = "/opt/homebrew/bin:/usr/local/bin:" + orig
	}
	if res.Stdout != want {
		t.Errorf("PATH = %q, want %q", res.Stdout, want)
	}
}

// TestPathPrefixPerGOOS pins the per-OS prefix table used by RunCmd.
func TestPathPrefixPerGOOS(t *testing.T) {
	cases := map[string]string{
		"darwin":  "/opt/homebrew/bin:/usr/local/bin",
		"linux":   "/usr/local/bin:/usr/bin:/bin",
		"windows": "/usr/local/bin:/usr/bin:/bin",
	}
	for goos, want := range cases {
		if got := pathPrefix(goos); got != want {
			t.Errorf("pathPrefix(%q) = %q, want %q", goos, got, want)
		}
	}
}
