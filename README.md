## Scale

Scale speeds up your automated tests by running them in parallel using Docker. Here are some of
its features:

* It automagically takes care of setting up multiple databases for each container to
connect to (whether that database is Postgres, Redis or so on).
* It autodetects your project's Dockerfile (or attempts to create one
for you) and uses that as the environment for your app
* It records the time taken for each test file and re-balances your
test files on each container based on the execution time of the previous run
* It uses [TOML](https://github.com/toml-lang/toml) to manage
configuration -- easier to read and edit that JSON or YAML

## Config

A Scale Config file for a Rails app looks like [this](docs/example-rails-config.toml)

## CLI Interface

### `scale init`

This initializes a Dockerfile (if a Dockerfile is
present, it doesn't do anything) and a `scale.toml` file in
the root of your project.

It does clever things such as detecting what
language the project is and creating a stock Dockerfile, for Rails/Node
apps it will look at your dependencies and decide whether you need a
Postgres/Redis/ElasticSearch instance and configure your `scale.toml`
accordingly.

### `scale run`

This spins up your environment based on your `scale.toml` file, creates
sandboxed databases (whether that is Redis/Postgres while retaining the
same schema) and runs tests in containers in parallel. The test files
are distributed across these containers based on the time it takes to
run each test file.

### `scale tests glob`

Accepts a glob pattern and expands that into all files in that glob

### `scale tests split`

This accepts a newline separated list of files and uses the JUnit
results which are stored in the file located at `global.test_results_path`
and picks the bucket of tests that correspond to that container "index".
Two implicit arguments to this function are

a) parallelism
b) the index of the container within who's scope these tests are
running

If a JUnit results file is not present at that location, then it
distributes files based on the nunber of lines in the file
