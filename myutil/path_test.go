package myutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectRootPath_PrefersCWDWhenPublicExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "public"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if got := ProjectRootPath(); got != dir {
		t.Errorf("ProjectRootPath() = %v, want %v", got, dir)
	}
}

func TestProjectRootPath_FallsBackToExecutableDir(t *testing.T) {
	dir := t.TempDir()
	// No "public" folder here, so ProjectRootPath should not report this
	// directory even though it's the current working directory.
	t.Chdir(dir)

	if got := ProjectRootPath(); got == dir {
		t.Errorf("ProjectRootPath() = %v, want it to fall back away from a public-less cwd", got)
	}
	if got, want := ProjectRootPath(), ExecutableDir(); got != want {
		t.Errorf("ProjectRootPath() = %v, want ExecutableDir() = %v", got, want)
	}
}

func TestStoragePathAndPublicPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "public"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if got, want := PublicPath(), filepath.Join(dir, "public"); got != want {
		t.Errorf("PublicPath() = %v, want %v", got, want)
	}
	if got, want := StoragePath(), filepath.Join(dir, "storage"); got != want {
		t.Errorf("StoragePath() = %v, want %v", got, want)
	}
}
