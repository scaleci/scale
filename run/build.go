package run

import (
	"bytes"
	"context"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/mholt/archiver"
)

func Build(app *App) error {
	buildArgs := map[string]*string{}

	dockerContext := new(bytes.Buffer)
	dockerFile := app.GlobalConfig.BuildWith
	path := filepath.Dir(dockerFile)

	err := archiver.Tar.Write(dockerContext, []string{path})
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	for k, v := range app.GlobalConfig.Env {
		argValue := v
		buildArgs[k] = &argValue
	}

	imageBuildResponse, err := cli.ImageBuild(
		context.Background(),
		dockerContext,
		types.ImageBuildOptions{
			Tags:       []string{app.ImageName()},
			Context:    dockerContext,
			Dockerfile: dockerFile,
			BuildArgs:  buildArgs})

	if err != nil {
		return err
	}

	if err := StreamDockerResponse(imageBuildResponse.Body, "stream", "error"); err != nil {
		return err
	}

	return nil
}
