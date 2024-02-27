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

// Create a factory for the stack struct
func NewStack(majorVersion int, engine string, rootDir string) Stack {
	return Stack{
		MajorVersion: majorVersion,
		AbsPath:      filepath.Join(rootDir, fmt.Sprintf("build-%s-%d", engine, majorVersion)),
		Engine:       engine,
	}
}
