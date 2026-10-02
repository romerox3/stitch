package shell

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Result holds output from a shell command.
type Result struct {
	Success bool
	Stdout  string
	Stderr  string
	Code    int
}

// pathPrefix returns the PATH entries RunCmd prepends on the given OS:
// Homebrew locations on macOS, the conventional binary locations elsewhere.
func pathPrefix(goos string) string {
	if goos == "darwin" {
		return "/opt/homebrew/bin:/usr/local/bin"
	}
	return "/usr/local/bin:/usr/bin:/bin"
}

// RunCmd executes a shell command, optionally in workingDir.
func RunCmd(command string, workingDir string) Result {
	cmd := exec.Command("sh", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	env := os.Environ()
	pathVar := ""
	pathIdx := -1
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			pathVar = strings.TrimPrefix(e, "PATH=")
			pathIdx = i
			break
		}
	}

	extraPaths := pathPrefix(runtime.GOOS)
	if pathVar != "" {
		pathVar = extraPaths + ":" + pathVar
	} else {
		pathVar = extraPaths
	}

	envEntry := "PATH=" + pathVar
	if pathIdx >= 0 {
		env[pathIdx] = envEntry
	} else {
		env = append(env, envEntry)
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			code = exitError.ExitCode()
		} else {
			code = 1
		}
	}

	return Result{
		Success: err == nil,
		Stdout:  strings.TrimSpace(stdout.String()),
		Stderr:  strings.TrimSpace(stderr.String()),
		Code:    code,
	}
}

// GetRepoInfoFromGit resolves GitHub owner and repo name from a local clone.
func GetRepoInfoFromGit(repoPath string) (string, string) {
	res := RunCmd("git remote get-url origin", repoPath)
	if res.Success {
		url := res.Stdout
		re := regexp.MustCompile(`github\.com[:/]([^/]+)\/([^.]+)(?:\.git)?`)
		match := re.FindStringSubmatch(url)
		if len(match) >= 3 {
			return match[1], match[2]
		}
	}
	return "", filepath.Base(repoPath)
}
