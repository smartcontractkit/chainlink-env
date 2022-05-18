package client

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _  = runtime.Caller(0)
	ProjectRoot = filepath.Join(filepath.Dir(b), "../")
	DistRoot    = filepath.Join(ProjectRoot, "dist")
	DiffsDir    = filepath.Join(ProjectRoot, "diffs")
)
