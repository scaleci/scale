package run

import "github.com/BurntSushi/toml"

type App struct {
	Title        string
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
		if s.Parallelism == 0 {
			s.Parallelism = 1
		}
	}

	return &app, nil
}
