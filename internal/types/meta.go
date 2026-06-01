package types

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xoctopus/x/misc/must"
)

type ServerMeta struct {
	Name    string
	Version string
}

func (m ServerMeta) UA() string {
	if len(m.Version) == 0 {
		return m.Name
	}
	return m.Name + "@" + m.Version
}

type RequestMeta struct {
	ID     string
	Method string
	Route  string
}

type OperationMeta struct {
	ServerMeta
	RequestMeta
}

func (m OperationMeta) UA() string {
	id := m.ID
	if len(id) == 0 {
		id = "unknown"
	}
	return m.ServerMeta.UA() + "(" + id + ")"
}

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
