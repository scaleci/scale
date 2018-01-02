package tests

import (
	"fmt"
	"os"

	"github.com/scaleci/scale/exec"
)

func Parallelize(protocol string, parallelCount int, opts map[string]string) error {
	if protocol == "postgres" {
		var userName string
		var host string
		var port string
		var database string
		var ok bool

		if userName, ok = opts["user"]; !ok {
			userName = os.Getenv("POSTGRES_USER")
		}
		if host, ok = opts["host"]; !ok {
			host = os.Getenv("POSTGRES_HOST")
		}
		if port, ok = opts["port"]; !ok {
			port = os.Getenv("POSTGRES_PORT")
		}
		if database, ok = opts["database"]; !ok {
			database = os.Getenv("POSTGRES_DATABASE")
		}

		for i := 0; i < parallelCount; i++ {
			err := exec.Run("psql", []string{
				"-U", userName,
				"-h", host,
				"-p", port,
				"-c", fmt.Sprintf("CREATE DATABASE scale%d template '%s'", i, database),
			}, "")

			if err != nil {
				return err
			}
		}
	}

	return nil
}
