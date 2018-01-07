package run

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/deckarep/golang-set"
)

type App struct {
	Name         string `toml:"title"`
	GlobalConfig Config `toml:"global"`
	Services     map[string]*Service
	Stages       map[string]*Stage

	Graph [][]*Stage
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
		s.Command = strings.TrimSpace(s.Command)
		s.ParentApp = &app

		if s.Parallelism == 0 {
			s.Parallelism = 1
		}
	}
	err := app.ResolveDependencies()
	if err != nil {
		return nil, err
	}

	if app.GlobalConfig.BuildWith == "" {
		app.GlobalConfig.BuildWith = "Dockerfile"
	}

	app.GlobalConfig.Tag = fmt.Sprintf("%d", time.Now().UnixNano())
	return &app, nil
}

func (a *App) ResolveDependencies() error {
	stageDependencies := make(map[string]mapset.Set)

	for id, stage := range a.Stages {
		dependencySet := mapset.NewSet()
		for _, depId := range stage.DependsOn {
			dependencySet.Add(depId)
		}
		stageDependencies[id] = dependencySet
	}

	currentIndex := 0

	for len(stageDependencies) != 0 {
		readySet := mapset.NewSet()

		for id, deps := range stageDependencies {
			if deps.Cardinality() == 0 {
				readySet.Add(id)
			}
		}

		if readySet.Cardinality() == 0 {
			return errors.New("circular dependency detected")
		}

		a.Graph = append(a.Graph, []*Stage{})

		for id := range readySet.Iter() {
			idString := id.(string)
			delete(stageDependencies, idString)
			a.Graph[currentIndex] = append(a.Graph[currentIndex], a.Stages[idString])
		}

		for id, deps := range stageDependencies {
			diff := deps.Difference(readySet)
			stageDependencies[id] = diff
		}

		currentIndex += 1
	}

	return nil
}

func (a *App) ImageName() string {
	return fmt.Sprintf("scale-%s:%s", a.Name, a.GlobalConfig.Tag)
}
