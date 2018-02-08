package run

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/moby/moby/client"
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
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	_, _, err = cli.ImageInspectWithRaw(context.Background(), scaleBinaryPath)

	if err != nil && client.IsErrImageNotFound(err) {
		_, err = cli.ImagePull(
			context.Background(),
			scaleBinaryPath,
			types.ImagePullOptions{})

		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func downloadFromContainer(path string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	containerBody, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{Image: scaleBinaryPath},
		nil,
		nil,
		"")

	if err != nil {
		return err
	}

	r, _, err := cli.CopyFromContainer(
		context.Background(),
		containerBody.ID,
		"/app/scale")

	if err != nil {
		return err
	}

	err = untar(path, r)
	if err != nil {
		return err
	}

	return nil
}

// untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
// from: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func untar(dst string, r io.Reader) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()

		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()
			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}
