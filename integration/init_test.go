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

}

func by(_ string, f func()) { f() }

func TestAcceptance(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs(".")
	Expect(err).ToNot(HaveOccurred())





	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Metadata", testMetadata)
	suite("BuildpackIntegration", testBuildpackIntegration)

	suite.Run(t)
}
