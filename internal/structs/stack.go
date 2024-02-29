package structs

import (
	"fmt"
)

type Stack struct {
	MajorVersion int
	Path         string
	Engine       string
}

func NewStack(majorVersion int, engine string, rootDir string) Stack {
	return Stack{
		MajorVersion: majorVersion,
		Path:         fmt.Sprintf("build-%s-%d", engine, majorVersion),
		Engine:       engine,
	}
}
