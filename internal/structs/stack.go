package structs

import (
	"fmt"
	"path/filepath"
)

type Stack struct {
	MajorVersion int
	AbsPath      string
	Engine       string
}

func NewStack(majorVersion int, engine string, rootDir string) Stack {
	return Stack{
		MajorVersion: majorVersion,
		AbsPath:      filepath.Join(rootDir, fmt.Sprintf("build-%s-%d", engine, majorVersion)),
		Engine:       engine,
	}
}
