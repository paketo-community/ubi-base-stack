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
	structs "github.com/paketo-community/ubi-base-stack/internal/structs"
	utils "github.com/paketo-community/ubi-base-stack/internal/utils"
)

func testNodejsStackIntegration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		err error

		pack   occam.Pack
		docker occam.Docker
		source string
		name   string

		// These variables probably are not being used
		image     occam.Image
		container occam.Container

		bpUbiRunImageOverrideImageID string
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("When building an app with nodejs stacks", func() {

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

		var stacks []structs.Stack

		for _, nodeMajorVersion := range settings.Config.NodeMajorVersions {
			stacks = append(stacks, structs.NewStack(nodeMajorVersion, "nodejs", root))
		}

		for _, stack := range stacks {
			// Create a copy of the stack to get the value and instead of the pointer
			stack := stack
			it(fmt.Sprintf("it successfully builds an app using Nodejs %d run image", stack.MajorVersion), func() {
				runArchive := filepath.Join(stack.AbsPath, "run.oci")
				bpUbiRunImageOverrideImageID, err = utils.PushFileToLocalRegistry(runArchive, RegistryUrl, fmt.Sprintf("run-%s-%d-%s", stack.Engine, stack.MajorVersion, uuid.NewString()))
				Expect(err).NotTo(HaveOccurred())

				image, _, err = pack.Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.Nodejs.Online,
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
				Eventually(container).Should(Serve(MatchRegexp(fmt.Sprintf(`v%d.*`, stack.MajorVersion))).OnPort(8080).WithEndpoint("/node/version"))
			})
		}
	})
}
