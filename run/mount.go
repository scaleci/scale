package run

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scaleci/scale/exec"
)

var scaleBinaryPath string
var scaleBinaryType string
var scaleVersion string

func DownloadScaleBinary() (string, error) {
	if scaleBinaryType == "docker" {
		// http://larstechnica.com/2017/05/docker-copy-files-from-image-to-local
		path := filepath.Join("/", "tmp", fmt.Sprintf("scale-%s", scaleVersion))
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return "", err
		}

		err = dockerPull()
		if err != nil {
			return "", err
		}

		err = downloadFromContainer(path)
		if err != nil {
			return "", err
		}
		return filepath.Join(path, "scale"), nil
	}

	return scaleBinaryPath, nil
}

func dockerPull() error {
	cmdName := "docker"
	cmdArgs := []string{
		"pull",
		scaleBinaryPath,
	}

	return exec.Run(cmdName, cmdArgs, "docker.pull")
}

func downloadFromContainer(path string) error {
	cmdName := "docker"
	cmdArgs := []string{
		"run",
		"--entrypoint",
		"",
		"--rm",
		"-v",
		fmt.Sprintf("%s:/tmp", path),
		scaleBinaryPath,
		"sh", "-c",
		"cp -r scale /tmp",
	}

	return exec.Run(cmdName, cmdArgs, "download.scale")
}
