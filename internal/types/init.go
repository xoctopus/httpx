package types

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xoctopus/x/misc/must"
)

var (
	PkgRoot     string
	ExposedRoot string
)

func init() {
	_, file, _, ok := runtime.Caller(0)
	must.BeTrueF(ok, "failed to trace stack caller")

	dir := filepath.Dir(file)
	for {
		var (
			filename = filepath.Join(dir, "go.mod")
			f        *os.File
			err      error
			info     os.FileInfo
		)

		if info, err = os.Stat(filename); err != nil || info.IsDir() {
			goto Next
		}

		f, err = os.Open(filename)
		if err != nil {
			goto Next
		}
		if scanner := bufio.NewScanner(f); scanner.Scan() {
			_, after, found := strings.Cut(scanner.Text(), "module ")
			if found {
				PkgRoot = strings.TrimSpace(after)
			}
		}
		_ = f.Close()

		if len(PkgRoot) > 0 {
			break
		}
	Next:
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	must.BeTrueF(len(PkgRoot) > 0, "failed to find module path")
	ExposedRoot = filepath.Join(PkgRoot, "pkg", "httpx")
}
