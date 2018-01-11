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

	testName(app, t)
	testConfig(app, t)
	testServices(app, t)
	testStages(app, t)
	testGraph(app, t)
}

func testName(app *App, t *testing.T) {
	expectedName := "Rails App"

	if app.Name != expectedName {
		t.Fatal(fmt.Sprintf("Expected name to be %s, got %s\n", expectedName, app.Name))
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
		"DATABASE_USER":  "postgres",
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
	testService(dbService, "database", "scaleci/postgres:9.6", "5432/tcp", t)

	redisService := app.Services["redis"]
	testService(redisService, "redis", "scaleci/redis:2.1", "6379/tcp", t)
}

func testService(s *Service, id string, image string, port string, t *testing.T) {
	if s.ID != id {
		t.Fatal(fmt.Sprintf("Expected s.ID to be %s, got %s\n", id, s.ID))
	}
	if s.Image != image {
		t.Fatal(fmt.Sprintf("Expected s.Image to be %s, got %s\n", image, s.Image))
	}
	if s.Port != port {
		t.Fatal(fmt.Sprintf("Expected s.Port to be %s, got %s\n", port, s.Port))
	}
}

func testStages(app *App, t *testing.T) {
	dbSetupStage := app.Stages["db.setup"]
	testStage(dbSetupStage, "db.setup", "bundle exec rake db:create db:structure:load && scale tests parallelize postgres --opts user=$DATABASE_USER,host=$DATABASE_HOST,port=$DATABASE_PORT,password=$DATABASE_PASSWORD", []string{}, int64(1), t)

	rspecStage := app.Stages["rspec"]
	rspecCommand := "bundle exec rspec --profile 10 --format RspecJunitFormatter --out /tmp/test-results/rspec.xml --require ./lib/block_progress_formatter.rb --format BlockProgressFormatter $(scale tests glob \"spec/**/*_spec.rb\" |xargs scale tests split)"
	testStage(rspecStage, "rspec", rspecCommand, []string{"db.setup"}, int64(8), t)

	rubocopStage := app.Stages["rubocop"]
	testStage(rubocopStage, "rubocop", "bundle exec rubocop", []string{"db.setup"}, int64(1), t)
}

func testGraph(app *App, t *testing.T) {
	expectedStageGraphLen := 2
	if stageGraphLen := len(app.Graph); stageGraphLen != expectedStageGraphLen {
		t.Fatalf("Expected stage graph length to be %d, got %d\n", expectedStageGraphLen, stageGraphLen)
	}

	expectedFirstStageLen := 1
	if firstStageLen := len(app.Graph[0]); firstStageLen != expectedFirstStageLen {
		t.Fatalf("Expected first stage length to be %d, got %d\n", expectedFirstStageLen, firstStageLen)
	}

	findStageID("db.setup", app.Graph[0], t)

	expectedSecondStageLen := 2
	if secondStageLen := len(app.Graph[1]); secondStageLen != expectedSecondStageLen {
		t.Fatalf("Expected second stage length to be %d, got %d\n", expectedSecondStageLen, secondStageLen)
	}

	findStageID("rspec", app.Graph[1], t)
	findStageID("rubocop", app.Graph[1], t)
}

func findStageID(id string, graph []*Stage, t *testing.T) {
	for _, s := range graph {
		if s.ID == id {
			return
		}
	}

	t.Fatalf("Expected to find %s, not in graph\n", id)
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
