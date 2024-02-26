package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2/vacation"
	"github.com/sclevine/spec"

	. "github.com/paketo-buildpacks/occam/matchers"
)

func testMetadata(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tmpDir string
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	it("builds base stack", func() {
		var (
			buildReleaseDate time.Time
			runReleaseDate   time.Time
		)

		by("confirming that the build image is correct", func() {
			index, manifests, err := getImageIndexAndManifests(tmpDir, filepath.Join(root, "./build", "build.oci"))
			Expect(err).NotTo(HaveOccurred())

			Expect(manifests).To(HaveLen(1))
			Expect(manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "base build ubi8 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			buildReleaseDate, err = time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(buildReleaseDate).NotTo(BeZero())

			Expect(image).To(SatisfyAll(
				HaveFileWithContent("/etc/group", ContainSubstring("cnb:x:1000:")),
				HaveFileWithContent("/etc/passwd", ContainSubstring("cnb:x:1002:1000::/home/cnb:/bin/bash")),
				HaveDirectory("/home/cnb"),
			))

			Expect(file.Config.User).To(Equal("1002:1000"))

			Expect(file.Config.Env).To(ContainElements(
				"CNB_USER_ID=1002",
				"CNB_GROUP_ID=1000",
				"CNB_STACK_ID=io.buildpacks.stacks.ubi8",
			))
		})

		by("confirming that the run image is correct", func() {
			index, manifests, err := getImageIndexAndManifests(tmpDir, filepath.Join(root, "./build", "run.oci"))
			Expect(err).NotTo(HaveOccurred())

			Expect(manifests).To(HaveLen(1))
			Expect(manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "base run ubi8 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDate, err = time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDate).NotTo(BeZero())

			Expect(file.Config.User).To(Equal("1001:1000"))

			Expect(image).To(SatisfyAll(
				HaveFileWithContent("/etc/group", ContainSubstring("cnb:x:1000:")),
				HaveFileWithContent("/etc/passwd", ContainSubstring("cnb:x:1001:1000::/home/cnb:/bin/bash")),
				HaveDirectory("/home/cnb"),
			))

			Expect(image).To(HaveFileWithContent("/etc/os-release", SatisfyAll(
				ContainLines(MatchRegexp(`PRETTY_NAME=\"Red Hat Enterprise Linux 8\.\d+ \(Ootpa\)\"`)),
				ContainSubstring(`HOME_URL="https://github.com/paketo-community/ubi-base-stack"`),
				ContainSubstring(`SUPPORT_URL="https://github.com/paketo-community/ubi-base-stack/blob/main/README.md"`),
				ContainSubstring(`BUG_REPORT_URL="https://github.com/paketo-community/ubi-base-stack/issues/new"`),
			)))
		})

		var engines = []struct {
			majorVersion int
			typeName     string
		}{
			{
				majorVersion: 16,
				typeName:     "nodejs",
			},
			{
				majorVersion: 18,
				typeName:     "nodejs",
			},
			{
				majorVersion: 20,
				typeName:     "nodejs",
			},
			{
				majorVersion: 8,
				typeName:     "java",
			},
			{
				majorVersion: 11,
				typeName:     "java",
			},
			{
				majorVersion: 17,
				typeName:     "java",
			},
			{
				majorVersion: 21,
				typeName:     "java",
			},
		}

		for _, engine := range engines {
			by(fmt.Sprintf("confirming that the run %s-%d image is correct", engine.typeName, engine.majorVersion), func() {

				index, manifests, err := getImageIndexAndManifests(tmpDir, filepath.Join(root, fmt.Sprintf("./build-%s-%d", engine.typeName, engine.majorVersion), "run.oci"))
				Expect(err).NotTo(HaveOccurred())

				Expect(manifests).To(HaveLen(1))
				Expect(manifests[0].Platform).To(Equal(&v1.Platform{
					OS:           "linux",
					Architecture: "amd64",
				}))

				image, err := index.Image(manifests[0].Digest)
				Expect(err).NotTo(HaveOccurred())

				file, err := image.ConfigFile()
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Config.Labels).To(SatisfyAll(
					HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
					HaveKeyWithValue("io.buildpacks.stack.description", fmt.Sprintf("ubi8 %s-%d image to support buildpacks", engine.typeName, engine.majorVersion)),
					HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
					HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
					HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
					HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
					HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
				))

				runNodejsReleaseDate, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
				Expect(err).NotTo(HaveOccurred())
				Expect(runNodejsReleaseDate).NotTo(BeZero())

				Expect(file.Config.User).To(Equal("1001:1000"))

				Expect(image).To(SatisfyAll(
					HaveFileWithContent("/etc/group", ContainSubstring("cnb:x:1000:")),
					HaveFileWithContent("/etc/passwd", ContainSubstring("cnb:x:1001:1000::/home/cnb:/bin/bash")),
					HaveDirectory("/home/cnb"),
				))

				Expect(image).To(HaveFileWithContent("/etc/os-release", SatisfyAll(
					ContainLines(MatchRegexp(`PRETTY_NAME=\"Red Hat Enterprise Linux 8\.\d+ \(Ootpa\)\"`)),
					ContainSubstring(`HOME_URL="https://github.com/paketo-community/ubi-base-stack"`),
					ContainSubstring(`SUPPORT_URL="https://github.com/paketo-community/ubi-base-stack/blob/main/README.md"`),
					ContainSubstring(`BUG_REPORT_URL="https://github.com/paketo-community/ubi-base-stack/issues/new"`),
				)))
			})
		}

		Expect(runReleaseDate).To(Equal(buildReleaseDate))
	})
}

func getImageIndexAndManifests(tmpDir string, ociImageFilePath string) (index v1.ImageIndex, manifests []v1.Descriptor, err error) {

	dir := filepath.Join(tmpDir, uuid.New().String())
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, []v1.Descriptor{}, err
	}

	archive, err := os.Open(ociImageFilePath)
	if err != nil {
		return nil, []v1.Descriptor{}, err
	}
	defer archive.Close()

	err = vacation.NewArchive(archive).Decompress(dir)
	if err != nil {
		return nil, []v1.Descriptor{}, err
	}

	path, err := layout.FromPath(dir)
	if err != nil {
		return nil, []v1.Descriptor{}, err
	}

	index, err = path.ImageIndex()
	if err != nil {
		return nil, []v1.Descriptor{}, err
	}

	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, []v1.Descriptor{}, err
	}

	return index, indexManifest.Manifests, nil
}
