package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	utils "github.com/paketo-community/ubi-base-stack/utils"
)

var nodeMajorVersions = []struct {
	nodeMajorVersion int
}{
	{
		nodeMajorVersion: 16,
	},
	{
		nodeMajorVersion: 18,
	},
	{
		nodeMajorVersion: 20,
	},
}

func testNodejsStackIntegration(t *testing.T, context spec.G, it spec.S) {
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

		bpUbiRunImageOverrideImageID string
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("When building an app", func() {

		it.Before(func() {

			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("integration", "testdata", "nodejs_simple_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
			Expect(docker.Image.Remove.Execute(bpUbiRunImageOverrideImageID)).To(Succeed())
		})

		for _, nmvs := range nodeMajorVersions {
			nodeMajorVersion := nmvs.nodeMajorVersion
			it(fmt.Sprintf("it successfully builds an app using Nodejs %d run image", nodeMajorVersion), func() {

				//Creating and pushing the run image to registry
				runNodejsArchive := filepath.Join(root, fmt.Sprintf("./build-nodejs-%d", nodeMajorVersion), "run.oci")
				bpUbiRunImageOverrideImageID, err = utils.PushFileToLocalRegistry(runNodejsArchive, RegistryUrl, fmt.Sprintf("run-nodejs-%d-%s", nodeMajorVersion, uuid.NewString()))
				Expect(err).NotTo(HaveOccurred())

				image, _, err = pack.Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.NPMInstall.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithBuilder(builder.imageUrl).
					WithNetwork("host").
					WithEnv(map[string]string{"BP_UBI_RUN_IMAGE_OVERRIDE": bpUbiRunImageOverrideImageID}).
					WithPullPolicy("always").
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithCommand("npm start").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve("Hello World!"))
			})
		}
	})
}
