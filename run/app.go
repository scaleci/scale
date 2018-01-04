package run

import (
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type App struct {
	Name         string `toml:"title"`
	GlobalConfig Config `toml:"global"`
	Services     map[string]*Service
	Stages       map[string]*Stage
}

func Parse(path string) (*App, error) {
	app := App{}

	if _, err := toml.DecodeFile(path, &app); err != nil {
		return nil, err
	}

	for id, s := range app.Services {
		s.ID = id
	}
	for id, s := range app.Stages {
		s.ID = id
		s.ParentApp = &app
		if s.Parallelism == 0 {
			s.Parallelism = 1
		}
		s.Command = strings.TrimSpace(s.Command)
	}

	if app.GlobalConfig.BuildWith == "" {
		app.GlobalConfig.BuildWith = "Dockerfile"
	}

	app.GlobalConfig.Tag = fmt.Sprintf("%d", time.Now().UnixNano())
	return &app, nil
}

func (a *App) ImageName() string {
	return fmt.Sprintf("scale-%s:%s", a.Name, a.GlobalConfig.Tag)
}

func (a *App) Env() map[string]string {
	finalEnv := map[string]string{}

	for k, v := range a.GlobalConfig.Env {
		finalEnv[k] = v
	}

	for _, s := range a.Services {
		for key, value := range s.Env() {
			finalEnv[key] = value
		}
	}

	return finalEnv
}
