package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/v2/pexec"
)

func GenerateBuilder(stackPath string, registryUrl string) (buildImageID string, buildImageUrl string, runImageID string, runImageUrl string, builderImageUrl string, err error) {

	buildArchive := filepath.Join(stackPath, "build.oci")
	buildImageID = fmt.Sprintf("build-nodejs-%s", uuid.NewString())
	err = archiveToDaemon(buildArchive, buildImageID)
	if err != nil {
		return "", "", "", "", "", err
	}

	buildImageUrl, err = PushFileToLocalRegistry(buildArchive, registryUrl, buildImageID)
	if err != nil {
		return "", "", "", "", "", err
	}

	runArchive := filepath.Join(stackPath, "run.oci")
	runImageID = fmt.Sprintf("run-nodejs-%s", uuid.NewString())
	err = archiveToDaemon(runArchive, runImageID)
	if err != nil {
		return "", "", "", "", "", err
	}
	runImageUrl, err = PushFileToLocalRegistry(runArchive, registryUrl, runImageID)
	if err != nil {
		return "", "", "", "", "", err
	}

	// Creating builder file
	builderConfigFile, err := os.CreateTemp("", "builder.toml")
	if err != nil {
		return "", "", "", "", "", err
	}

	builderConfigFilepath := builderConfigFile.Name()

	_, err = fmt.Fprintf(builderConfigFile, `
			[stack]
			  id = "io.buildpacks.stacks.ubi8"
			  build-image = "%s:latest"
			  run-image = "%s:latest"
			`, buildImageUrl, runImageUrl)

	if err != nil {
		return "", "", "", "", "", err
	}

	// naming builder and pushing it to registry with pack cli
	builderImageUrl = fmt.Sprintf("%s/builder-%s", registryUrl, uuid.NewString())

	buf := bytes.NewBuffer(nil)

	pack := pexec.NewExecutable("pack")
	err = pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"create",
			builderImageUrl,
			fmt.Sprintf("--config=%s", builderConfigFilepath),
			"--publish",
		},
	})

	if err != nil {
		return "", "", "", "", "", err
	}

	err = os.RemoveAll(builderConfigFilepath)
	if err != nil {
		return "", "", "", "", "", err
	}

	return buildImageID, buildImageUrl, runImageID, runImageUrl, builderImageUrl, nil
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

func PushFileToLocalRegistry(filePath string, registryUrl string, imageName string) (string, error) {
	buf := bytes.NewBuffer(nil)

	imageURL := fmt.Sprintf("%s/%s", registryUrl, imageName)

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

func RemoveImages(docker occam.Docker, imageIDs []string) error {

	for _, imageID := range imageIDs {
		err := docker.Image.Remove.Execute(imageID)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetLifecycleImageID(docker occam.Docker, builderImageUrl string) (lifecycleImageID string, err error) {

	lifecycleVersion, err := GetLifecycleVersion(builderImageUrl)
	if err != nil {
		return "", err
	}

	lifecycleImageID = fmt.Sprintf("buildpacksio/lifecycle:%s", lifecycleVersion)

	return lifecycleImageID, nil
}

type Builder struct {
	LocalInfo struct {
		Lifecycle struct {
			Version string `json:"version"`
		} `json:"lifecycle"`
	} `json:"remote_info"`
}

func GetLifecycleVersion(builderUrl string) (string, error) {
	buf := bytes.NewBuffer(nil)
	pack := pexec.NewExecutable("pack")
	err := pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"inspect",
			builderUrl,
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
