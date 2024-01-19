package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var root string
var settings struct {
	Buildpacks struct {
		NodeEngine struct {
			Online string
		}
		NPMInstall struct {
			Online string
		}
		BuildPlan struct {
			Online string
		}
	}

	Extensions struct {
		UbiNodejsExtension struct {
			Online string
		}
	}

	Config struct {
		BuildPlan          string `json:"build-plan"`
		UbiNodejsExtension string `json:"ubi-nodejs-extension"`
		NodeEngine         string `json:"node-engine"`
		NPMInstall         string `json:"npm-install"`
	}
}

var RegistryUrl string

// func init() {
// 	flag.StringVar(&LocalRegistryUrl, "registry-url", "", "")
// }

func by(_ string, f func()) { f() }

//TODO how do we test the run.oci, because i dont know when it is applied.
//TODO on java base stasck, there is no extension.toml

func TestAcceptance(t *testing.T) {
	var err error

	Expect := NewWithT(t).Expect

	// RegistryUrl := "localhost:5000"
	RegistryUrl = os.Getenv("REGISTRY_URL")
	// flag.Parse()
	Expect(RegistryUrl).NotTo(Equal(""))

	root, err = filepath.Abs(".")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("./integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	// settings.Extensions.UbiNodejsExtension.Online, err = buildpackStore.Get.
	// 	Execute(settings.Config.UbiNodejsExtension)
	// Expect(err).ToNot(HaveOccurred())

	settings.Extensions.UbiNodejsExtension.Online, err = buildpackStore.Get.
		WithVersion("0.0.1").
		// Execute("github.com/paketo-community/ubi-nodejs-extension")
		Execute("/home/costas/RedHat/buildpacks/ubi-nodejs-extension")
	Expect(err).NotTo(HaveOccurred())

	// settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
	// 	Execute(settings.Config.NodeEngine)
	// Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		WithVersion("0.0.1").
		Execute("/home/costas/RedHat/buildpacks/node-engine")
	Expect(err).NotTo(HaveOccurred())

	// settings.Buildpacks.NPMInstall.Online, err = buildpackStore.Get.
	// 	Execute(settings.Config.NPMInstall)
	// Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NPMInstall.Online, err = buildpackStore.Get.
		WithVersion("0.0.1").
		Execute("/home/costas/RedHat/buildpacks/npm-install")
	Expect(err).NotTo(HaveOccurred())

	// settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
	// 	Execute(settings.Config.BuildPlan)
	// Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		WithVersion("0.0.1").
		Execute("/home/costas/RedHat/buildpacks/build-plan")
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	// suite("Metadata", testMetadata)
	suite("NodejsStackIntegration", testNodejsStackIntegration)
	suite.Run(t)
}
