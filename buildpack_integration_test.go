package acceptance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/packit/v2/pexec"
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

		buildImageID   string
		runImageID     string
		builderImageID string
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
			Expect(docker.Image.Remove.Execute(buildImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(runImageID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(builderImageID)).To(Succeed())
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

				buildImageID, runImageID, builderImageID, err = generateBuilder(filepath.Join(root, currentStackRelativePath))
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
					WithBuilder(builderImageID).
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

type Builder struct {
	LocalInfo struct {
		Lifecycle struct {
			Version string `json:"version"`
		} `json:"lifecycle"`
	} `json:"local_info"`
}

func getLifecycleVersion(builderID string) (string, error) {
	buf := bytes.NewBuffer(nil)
	pack := pexec.NewExecutable("pack")
	err := pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"inspect",
			builderID,
			"-o",
			"json",
		},
	})

	if err != nil {
		return "", err
	}

	var builder Builder
	err = json.Unmarshal([]byte(buf.String()), &builder)
	if err != nil {
		return "", err
	}
	return builder.LocalInfo.Lifecycle.Version, nil
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
