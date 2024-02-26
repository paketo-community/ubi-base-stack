package acceptance_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var stack struct {
	BuildArchive       string
	RunArchive         string
	BuildImageID       string
	RunImageID         string
	RunNodejs16Archive string
	RunNodejs18Archive string
	RunNodejs20Archive string
	RunJava8Archive    string
	RunJava11Archive   string
	RunJava17Archive   string
	RunJava21Archive   string
}

func by(_ string, f func()) { f() }

func TestAcceptance(t *testing.T) {
	format.MaxLength = 0
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs(".")
	Expect(err).ToNot(HaveOccurred())

	stack.BuildArchive = filepath.Join(root, "build", "build.oci")
	stack.BuildImageID = fmt.Sprintf("stack-build-%s", uuid.NewString())

	stack.RunArchive = filepath.Join(root, "build", "run.oci")
	stack.RunImageID = fmt.Sprintf("stack-run-%s", uuid.NewString())

	stack.RunNodejs16Archive = filepath.Join(root, "build-nodejs-16", "run.oci")
	stack.RunNodejs18Archive = filepath.Join(root, "build-nodejs-18", "run.oci")
	stack.RunNodejs20Archive = filepath.Join(root, "build-nodejs-20", "run.oci")
	stack.RunJava8Archive = filepath.Join(root, "build-java-8", "run.oci")
	stack.RunJava11Archive = filepath.Join(root, "build-java-11", "run.oci")
	stack.RunJava17Archive = filepath.Join(root, "build-java-17", "run.oci")
	stack.RunJava21Archive = filepath.Join(root, "build-java-21", "run.oci")

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Metadata", testMetadata)
	suite("BuildpackIntegration", testBuildpackIntegration)

	suite.Run(t)
}
