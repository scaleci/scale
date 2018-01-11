package run

type Config struct {
	BuildWith       string `toml:"build_with"`
	Tag             string
	Version         string
	TestResultsPath string `toml:"test_results_path"`
	Env             map[string]string
}
