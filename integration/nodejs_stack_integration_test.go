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

func testNodejsStackIntegration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		err error

		builderConfigFilepath string

		pack   occam.Pack
		docker occam.Docker
		source string
		name   string

		image     occam.Image
		container occam.Container
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()

		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(builderConfigFilepath)).To(Succeed())
	})

	context("When using Node.js 16 as run image", func() {
		var (
			err                error
			builder            string
			buildImageID       string
			runImageID         string
			runNodejs16ImageID string
			nodeMajorVersion   int
		)

		it.After(func() {
			Expect(docker.Image.Remove.Execute(builder)).To(Succeed())
			Expect(docker.Image.Remove.Execute(buildImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runNodejs16ImageID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it.Before(func() {
			nodeMajorVersion = 16

			buildImageID, runImageID, builderConfigFilepath, builder, err = generateBuilder(root)
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "nodejs_simple_app"))
			Expect(err).NotTo(HaveOccurred())

			//Creating and pushing the run image to local registry
			runNodejs16Archive := filepath.Join(root, fmt.Sprintf("../build-nodejs-%d", nodeMajorVersion), "run.oci")
			runNodejs16ImageID = fmt.Sprintf("run-nodejs-%d-%s", nodeMajorVersion, uuid.NewString())
			Expect(err).NotTo(HaveOccurred())

			Expect(archiveToDaemon(runNodejs16Archive, runNodejs16ImageID)).To(Succeed())
		})

		it("sucessfully builds an app", func() {

			image, _, err = pack.Build.
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.NPMInstall.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithBuilder(builder).
				WithNetwork("host").
				WithEnv(map[string]string{"BP_NODE_RUN_EXTENSION": runNodejs16ImageID}).
				WithPullPolicy("always").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			container, err = docker.Container.Run.
				WithPublish("8080").
				WithCommand("node server.js").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(Serve("Hello World!"))
		})
	})

	context("When using Node.js 18 as run image", func() {
		var (
			err                error
			builder            string
			buildImageID       string
			runImageID         string
			runNodejs18ImageID string
			nodeMajorVersion   int
		)

		it.After(func() {
			Expect(docker.Image.Remove.Execute(builder)).To(Succeed())
			Expect(docker.Image.Remove.Execute(buildImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runNodejs18ImageID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it.Before(func() {
			nodeMajorVersion = 18

			buildImageID, runImageID, builderConfigFilepath, builder, err = generateBuilder(root)
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "nodejs_simple_app"))
			Expect(err).NotTo(HaveOccurred())

			//Creating and pushing the run image to local registry
			runNodejs18Archive := filepath.Join(root, fmt.Sprintf("../build-nodejs-%d", nodeMajorVersion), "run.oci")
			runNodejs18ImageID = fmt.Sprintf("run-nodejs-%d-%s", nodeMajorVersion, uuid.NewString())
			Expect(err).NotTo(HaveOccurred())

			Expect(archiveToDaemon(runNodejs18Archive, runNodejs18ImageID)).To(Succeed())
		})

		it("sucessfully builds an app", func() {

			image, _, err = pack.Build.
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.NPMInstall.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithBuilder(builder).
				WithNetwork("host").
				WithEnv(map[string]string{"BP_NODE_RUN_EXTENSION": runNodejs18ImageID}).
				WithPullPolicy("always").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			container, err = docker.Container.Run.
				WithPublish("8080").
				WithCommand("node server.js").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(Serve("Hello World!"))
		})
	})

	context("When using Node.js 20 as run image", func() {
		var (
			err                error
			builder            string
			buildImageID       string
			runImageID         string
			runNodejs20ImageID string
			nodeMajorVersion   int
		)

		it.After(func() {
			Expect(docker.Image.Remove.Execute(builder)).To(Succeed())
			Expect(docker.Image.Remove.Execute(buildImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runNodejs20ImageID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it.Before(func() {
			nodeMajorVersion = 20

			buildImageID, runImageID, builderConfigFilepath, builder, err = generateBuilder(root)
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "nodejs_simple_app"))
			Expect(err).NotTo(HaveOccurred())

			//Creating and pushing the run image to local registry
			runNodejs20Archive := filepath.Join(root, fmt.Sprintf("../build-nodejs-%d", nodeMajorVersion), "run.oci")
			runNodejs20ImageID = fmt.Sprintf("run-nodejs-%d-%s", nodeMajorVersion, uuid.NewString())
			Expect(err).NotTo(HaveOccurred())

			Expect(archiveToDaemon(runNodejs20Archive, runNodejs20ImageID)).To(Succeed())
		})

		it("sucessfully builds an app", func() {

			image, _, err = pack.Build.
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.NPMInstall.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithBuilder(builder).
				WithNetwork("host").
				WithEnv(map[string]string{"BP_NODE_RUN_EXTENSION": runNodejs20ImageID}).
				WithPullPolicy("always").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			container, err = docker.Container.Run.
				WithPublish("8080").
				WithCommand("node server.js").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(Serve("Hello World!"))
		})
	})

}

func archiveToDaemon(path, id string) error {
	skopeo := pexec.NewExecutable("skopeo")

	return skopeo.Execute(pexec.Execution{
		Args: []string{
			"copy",
			fmt.Sprintf("oci-archive:%s", path),
			fmt.Sprintf("docker-daemon:%s:latest", id),
		},
	})
}

func createBuilder(config string, name string) (string, error) {
	buf := bytes.NewBuffer(nil)

	pack := pexec.NewExecutable("pack")
	err := pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"create",
			name,
			fmt.Sprintf("--config=%s", config),
			"--publish",
		},
	})
	return buf.String(), err
}

func pushFileToLocalRegistry(filePath string, localRegistryUrl string, imageName string) (string, error) {
	buf := bytes.NewBuffer(nil)

	var imageURL = fmt.Sprintf("%s/%s", localRegistryUrl, imageName)

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

func generateBuilder(root string) (BImageID string, RImageID string, BConfigFilepath string, builder string, err error) {
	buildArchive := filepath.Join(root, "../build", "build.oci")
	buildImageID := fmt.Sprintf("build-nodejs-%s", uuid.NewString())
	buildImageURL, err := pushFileToLocalRegistry(buildArchive, LocalRegistryUrl, buildImageID)
	if err != nil {
		return "", "", "", "", err
	}

	runArchive := filepath.Join(root, "../build", "run.oci")
	runImageID := fmt.Sprintf("run-nodejs-%s", uuid.NewString())
	runImageURL, err := pushFileToLocalRegistry(runArchive, LocalRegistryUrl, runImageID)
	if err != nil {
		return "", "", "", "", err
	}

	//Pushing Builder's stack images
	err = archiveToDaemon(buildArchive, buildImageID)
	if err != nil {
		return "", "", "", "", err
	}

	err = archiveToDaemon(runArchive, runImageID)
	if err != nil {
		return "", "", "", "", err
	}

	//Creating builder file
	builderConfigFile, err := os.CreateTemp("", "builder.toml")
	if err != nil {
		return "", "", "", "", err
	}

	builderConfigFilepath := builderConfigFile.Name()

	_, err = fmt.Fprintf(builderConfigFile, `
			[stack]
			  id = "io.buildpacks.stacks.ubi8"
			  build-image = "%s:latest"
			  run-image = "%s:latest"
			`, buildImageURL, runImageURL)

	if err != nil {
		return "", "", "", "", err
	}

	//Pushing builder to local registry
	builder = fmt.Sprintf("%s/builder-%s", LocalRegistryUrl, uuid.NewString())
	_, err = createBuilder(builderConfigFilepath, builder)
	if err != nil {
		return "", "", "", "", err
	}

	return buildImageID, runImageID, builderConfigFilepath, builder, nil
}
