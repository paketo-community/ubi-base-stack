package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/occam"
	structs "github.com/paketo-community/ubi-base-stack/internal/structs"
	utils "github.com/paketo-community/ubi-base-stack/internal/utils"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var root string
var RegistryUrl string

var builder struct {
	imageUrl      string
	buildImageID  string
	buildImageUrl string
	runImageID    string
	runImageUrl   string
}

var settings struct {
	Buildpacks struct {
		Nodejs struct {
			Online string
		}
		BuildPlan struct {
			Online string
		}
		GoDist struct {
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
		Nodejs             string `json:"nodejs"`
		GoDist             string `json:"go-dist"`
		NodeMajorVersions  []int  `json:"nodejs-major-versions"`
	}

	Stacks []structs.Stack
}

func by(_ string, f func()) { f() }

func TestAcceptance(t *testing.T) {

	for _, nodeMajorVersion := range settings.Config.NodeMajorVersions {
		settings.Stacks = append(settings.Stacks, structs.NewStack(nodeMajorVersion, "nodejs", root))
	}

	var err error
	Expect := NewWithT(t).Expect

	docker := occam.NewDocker()

	RegistryUrl = os.Getenv("REGISTRY_URL")
	Expect(RegistryUrl).NotTo(Equal(""))

	root, err = filepath.Abs(".")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("./integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	settings.Extensions.UbiNodejsExtension.Online, err = buildpackStore.Get.
		Execute(settings.Config.UbiNodejsExtension)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.Nodejs.Online, err = buildpackStore.Get.
		Execute(settings.Config.Nodejs)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.GoDist.Online, err = buildpackStore.Get.
		Execute(settings.Config.GoDist)
	Expect(err).NotTo(HaveOccurred())

	builder.buildImageID, builder.buildImageUrl, builder.runImageID, builder.runImageUrl, builder.imageUrl, err = utils.GenerateBuilder(root, "build", RegistryUrl)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(120 * time.Second)

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Metadata", testMetadata)
	suite("NodejsStackIntegration", testNodejsStackIntegration)
	suite("buildpackIntegration", testBuildpackIntegration)
	suite.Run(t)

	/** Cleanup **/
	lifecycleImageID, err := utils.GetLifecycleImageID(docker, builder.imageUrl)
	Expect(err).NotTo(HaveOccurred())

	err = utils.RemoveImages(docker, []string{builder.buildImageID, builder.runImageID, lifecycleImageID, builder.runImageUrl, builder.imageUrl})
	Expect(err).NotTo(HaveOccurred())

}
