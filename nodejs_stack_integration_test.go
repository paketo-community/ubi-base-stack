package acceptance_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/packit/v2/pexec"
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
				bpUbiRunImageOverrideImageID, err = pushFileToLocalRegistry(runNodejsArchive, RegistryUrl, fmt.Sprintf("run-nodejs-%d-%s", nodeMajorVersion, uuid.NewString()))
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
					WithBuilder(builderImageID).
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

func pushFileToLocalRegistry(filePath string, registryUrl string, imageName string) (string, error) {
	buf := bytes.NewBuffer(nil)

	var imageURL = fmt.Sprintf("%s/%s", registryUrl, imageName)

	skopeo := pexec.NewExecutable("skopeo")

	err := skopeo.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"copy",
			fmt.Sprintf("oci-archive:%s", filePath),
			fmt.Sprintf("docker://%s:latest", imageURL),
			"--dest-tls-verify=false",
		},
	})

	if err != nil {
		return buf.String(), err
	} else {
		return imageURL, nil
	}
}

func generateBuilder(stackPath string) (BImageID string, RImageID string, builderImageID string, err error) {
	buildArchive := filepath.Join(stackPath, "build.oci")
	buildImageID := fmt.Sprintf("build-nodejs-%s", uuid.NewString())
	buildImageURL, err := pushFileToLocalRegistry(buildArchive, RegistryUrl, buildImageID)
	if err != nil {
		return "", "", "", err
	}

	runArchive := filepath.Join(stackPath, "run.oci")
	runImageID := fmt.Sprintf("run-nodejs-%s", uuid.NewString())
	runImageURL, err := pushFileToLocalRegistry(runArchive, RegistryUrl, runImageID)
	if err != nil {
		return "", "", "", err
	}

	err = archiveToDaemon(buildArchive, buildImageID)
	if err != nil {
		return "", "", "", err
	}

	err = archiveToDaemon(runArchive, runImageID)
	if err != nil {
		return "", "", "", err
	}

	// Creating builder file
	builderConfigFile, err := os.CreateTemp("", "builder.toml")
	if err != nil {
		return "", "", "", err
	}

	builderConfigFilepath := builderConfigFile.Name()

	_, err = fmt.Fprintf(builderConfigFile, `
			[stack]
			  id = "io.buildpacks.stacks.ubi8"
			  build-image = "%s:latest"
			  run-image = "%s:latest"
			`, buildImageURL, runImageURL)

	if err != nil {
		return "", "", "", err
	}

	// naming builder and pushing it to registry with pack cli
	builderImageID = fmt.Sprintf("%s/builder-%s", RegistryUrl, uuid.NewString())

	buf := bytes.NewBuffer(nil)

	pack := pexec.NewExecutable("pack")
	err = pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"create",
			builderImageID,
			fmt.Sprintf("--config=%s", builderConfigFilepath),
			"--publish",
		},
	})

	if err != nil {
		return "", "", "", err
	}

	err = os.RemoveAll(builderConfigFilepath)
	if err != nil {
		return "", "", "", err
	}

	return buildImageID, runImageID, builderImageID, nil
}
