package run

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		buildArgs[k] = &v
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
	defer imageBuildResponse.Body.Close()

	scanner := bufio.NewScanner(imageBuildResponse.Body)
	for scanner.Scan() {
		var out map[string]interface{}
		err := json.Unmarshal(scanner.Bytes(), &out)

		if err != nil {
			return err
		}

		if stream, ok := out["stream"].(string); ok {
			fmt.Printf(stream)
		}
		if errorMsg, ok := out["error"].(string); ok {
			fmt.Printf("%s\n", errorMsg)
		}
	}

	return nil
}
