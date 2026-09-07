package myutil

import (
	"os"
	"path/filepath"
)

// ExecutableDir returns the directory containing the running binary. Falling
// back to it (see ProjectRootPath) lets ptt-alertor find its own public/ and
// storage/ folders even when launched with a working directory that doesn't
// match where the binary lives - e.g. a Windows shortcut, Task Scheduler
// entry, or NSSM service that doesn't "cd" there first.
func ExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		dir, _ := os.Getwd()
		return dir
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Dir(exe)
}

// ProjectRootPath returns the directory ptt-alertor's bundled resources
// (public/, storage/) live under. It prefers the current working directory
// (so `go run .` / `docker run` with WORKDIR set keep working unchanged),
// and only falls back to the executable's own directory when "public"
// isn't found there.
func ProjectRootPath() string {
	if dir, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(dir, "public")); err == nil {
			return dir
		}
	}
	return ExecutableDir()
}

func StoragePath() string {
	return filepath.Join(ProjectRootPath(), "storage")
}

func PublicPath() string {
	return filepath.Join(ProjectRootPath(), "public")
}
