package acceptance_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	root                  string
	buildImageID          string
	runImageID            string
	builderConfigFilepath string
	builderImageID        string
	RegistryUrl           string
)

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

func by(_ string, f func()) { f() }

func TestAcceptance(t *testing.T) {
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

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NPMInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.NPMInstall)
	Expect(err).ToNot(HaveOccurred())

	buildImageID, runImageID, builderConfigFilepath, builderImageID, err = generateBuilder(root)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Acceptance", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Metadata", testMetadata)
	suite("NodejsStackIntegration", testNodejsStackIntegration)
	suite.Run(t)

	/** Cleanup **/
	err = removeLifecycleImage(docker, builderImageID)
	Expect(err).NotTo(HaveOccurred())

	err = removeBuilderImages(docker, builderImageID, buildImageID, runImageID, builderConfigFilepath)
	Expect(err).NotTo(HaveOccurred())
}

func removeLifecycleImage(docker occam.Docker, builderImageID string) error {

	lifecycleVersion, err := getLifecycleVersion(builderImageID)
	if err != nil {
		return err
	}

	err = docker.Image.Remove.Execute(fmt.Sprintf("buildpacksio/lifecycle:%s", lifecycleVersion))
	if err != nil {
		return err
	}

	return nil
}

// Remove builder run image and build image
func removeBuilderImages(docker occam.Docker, builderImageID string, buildImageID string, runImageID string, builderConfigFilepath string) error {

	var err error

	err = docker.Image.Remove.Execute(buildImageID)
	if err != nil {
		return err
	}

	err = docker.Image.Remove.Execute(runImageID)
	if err != nil {
		return err
	}

	err = docker.Image.Remove.Execute(fmt.Sprintf("%s/%s", RegistryUrl, runImageID))
	if err != nil {
		return err
	}

	err = docker.Image.Remove.Execute(builderImageID)
	if err != nil {
		return err
	}

	os.RemoveAll(builderConfigFilepath)

	return nil
}
