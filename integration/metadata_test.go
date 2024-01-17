package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
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
			dir := filepath.Join(tmpDir, "build-index")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build", "build.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
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
			dir := filepath.Join(tmpDir, "run-index")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
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

		by("confirming that the run nodejs-16 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-nodejs-16")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build-nodejs-16", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 nodejs-16 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateNodejs16, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateNodejs16).NotTo(BeZero())

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

		by("confirming that the run nodejs-18 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-nodejs-18")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build-nodejs-18", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 nodejs-18 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateNodejs18, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateNodejs18).NotTo(BeZero())

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

		by("confirming that the run nodejs-20 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-nodejs-20")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build-nodejs-20", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 nodejs-20 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateNodejs20, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateNodejs20).NotTo(BeZero())

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

		by("confirming that the run java-8 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-java-8")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build-java-8", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 java-8 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateJava8, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateJava8).NotTo(BeZero())

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

		by("confirming that the run java-11 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-java-11")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build-java-11", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 java-11 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateJava11, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateJava11).NotTo(BeZero())

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

		by("confirming that the run java-17 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-java-17")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(filepath.Join(root, "../build-java-17", "run.oci"))
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 java-17 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateJava17, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateJava17).NotTo(BeZero())

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

		by("confirming that the run java-21 image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index-java-21")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(stack.RunJava21Archive)
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.ubi8"),
				HaveKeyWithValue("io.buildpacks.stack.description", "ubi8 java-21 image to support buildpacks"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "rhel"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", MatchRegexp(`8\.\d+`)),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-community/ubi-base-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Community"),
				HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			))

			runReleaseDateJava21, err := time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			Expect(runReleaseDateJava21).NotTo(BeZero())

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

		Expect(runReleaseDate).To(Equal(buildReleaseDate))
	})
}
