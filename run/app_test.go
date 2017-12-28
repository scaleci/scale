package run

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	app, err := Parse("../docs/example-rails-config.toml")
	if err != nil {
		t.Fatal(fmt.Sprintf("Expected nil error, got %+v\n", err))
	}

	testTitle(app, t)
	testConfig(app, t)
	testServices(app, t)
	testStages(app, t)
}

func testTitle(app *App, t *testing.T) {
	expectedTitle := "Rails App"

	if app.Title != expectedTitle {
		t.Fatal(fmt.Sprintf("Expected title to be %s, got %\n", expectedTitle, app.Title))
	}
}

func testConfig(app *App, t *testing.T) {
	expectedBuildWith := "Dockerfile.test"
	expectedVersion := "1.0"
	expectedTestResultsPath := "/tmp/test-results/rspec.xml"
	expectedEnv := map[string]string{
		"RAILS_ENV":      "test",
		"AUTH_URL":       "http://example.com/a/auth",
		"AUTH_CLIENT_ID": "deadbeef",
	}

	// test config
	if app.GlobalConfig.BuildWith != expectedBuildWith {
		t.Fatal(fmt.Sprintf("Expected BuildWith to be %s, got %s\n", expectedBuildWith, app.GlobalConfig.BuildWith))
	}
	if app.GlobalConfig.Version != expectedVersion {
		t.Fatal(fmt.Sprintf("Expected BuildWith to be %s, got %s\n", expectedVersion, app.GlobalConfig.Version))
	}
	if app.GlobalConfig.TestResultsPath != expectedTestResultsPath {
		t.Fatal(fmt.Sprintf("Expected BuildWith to be %s, got %s\n", expectedTestResultsPath, app.GlobalConfig.TestResultsPath))
	}
	if !reflect.DeepEqual(expectedEnv, app.GlobalConfig.Env) {
		t.Fatal(fmt.Sprintf("Expected Env to be %+v, got %+v\n", expectedEnv, app.GlobalConfig.Env))
	}
}

func testServices(app *App, t *testing.T) {
	dbService := app.Services["database"]
	testService(dbService, "database", "scaleci/postgres:9.6", []string{"5432/tcp"}, t)

	redisService := app.Services["redis"]
	testService(redisService, "redis", "scaleci/redis:2.1", []string{"6379/tcp"}, t)
}

func testService(s *Service, id string, image string, ports []string, t *testing.T) {
	if s.ID != id {
		t.Fatal(fmt.Sprintf("Expected s.ID to be %s, got %s\n", id, s.ID))
	}
	if s.Image != image {
		t.Fatal(fmt.Sprintf("Expected s.Image to be %s, got %s\n", image, s.Image))
	}
	if !slicesEqual(s.Ports, ports) {
		t.Fatal(fmt.Sprintf("Expected s.Ports to be %+v, got %+v\n", ports, s.Ports))
	}
}

func testStages(app *App, t *testing.T) {
	dbSetupStage := app.Stages["db.setup"]
	testStage(dbSetupStage, "db.setup", "bundle exec rake db:create db:structure:load", []string{}, int64(1), t)

	rspecStage := app.Stages["rspec"]
	rspecCommand := "bundle exec rspec --profile 10 --format RspecJunitFormatter --out /tmp/test-results/rspec.xml --require ./lib/block_progress_formatter.rb --format BlockProgressFormatter $(scale tests glob \"spec/**/*_spec.rb\" |xargs scale tests split)"
	testStage(rspecStage, "rspec", rspecCommand, []string{"db.setup"}, int64(8), t)

	rubocopStage := app.Stages["rubocop"]
	testStage(rubocopStage, "rubocop", "bundle exec rubocop", []string{"db.setup"}, int64(1), t)
}

func testStage(s *Stage, id string, command string, dependsOn []string, parallelism int64, t *testing.T) {
	t.Logf("stage: %+v\n", s)

	if s.ID != id {
		t.Fatal(fmt.Sprintf("Expected s.ID to be %s, got %s\n", id, s.ID))
	}
	if s.Command != command {
		t.Fatal(fmt.Sprintf("Expected s.Command to be '%s', got '%s'\n", command, s.Command))
	}
	if !slicesEqual(s.DependsOn, dependsOn) {
		t.Fatal(fmt.Sprintf("Expected s.DependsOn to be %+v, got %+v\n", dependsOn, s.DependsOn))
	}
	if s.Parallelism != parallelism {
		t.Fatal(fmt.Sprintf("Expected s.Parallelism to be %d, got %d\n", parallelism, s.Parallelism))
	}
}

func slicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
