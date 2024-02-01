package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	utils "github.com/paketo-community/ubi-base-stack/utils"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackIntegration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		err error

		pack   occam.Pack
		docker occam.Docker
		source string
		name   string

		image     occam.Image
		container occam.Container

		buildImageID    string
		runImageID      string
		runImageUrl     string
		builderImageUrl string
	)

	stackRelativePaths := []string{
		"build",
		"build-nodejs-16",
		"build-nodejs-18",
		"build-nodejs-20",
	}

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("When building an app", func() {

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			err = utils.RemoveImages(docker, []string{buildImageID, runImageID, runImageUrl, builderImageUrl})
			Expect(err).NotTo(HaveOccurred())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it.Before(func() {
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("integration", "testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		for _, srp := range stackRelativePaths {
			currentStackRelativePath := srp

			it(fmt.Sprintf("should be a success when using run image from stack %s", currentStackRelativePath), func() {
				buildImageID, _, runImageID, runImageUrl, builderImageUrl, err = utils.GenerateBuilder(filepath.Join(root, currentStackRelativePath), RegistryUrl)
				Expect(err).NotTo(HaveOccurred())

				image, _, err = pack.WithNoColor().Build.
					WithBuildpacks(
						settings.Buildpacks.GoDist.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL": "DEBUG",
					}).
					WithPullPolicy("if-not-present").
					WithBuilder(builderImageUrl).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				container, err = docker.Container.Run.
					WithDirect().
					WithCommand("go").
					WithCommandArgs([]string{"run", "main.go"}).
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(MatchRegexp(`go1.*`)).OnPort(8080))

			})
		}
	})
}
