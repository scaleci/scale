package run

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func StreamDockerResponse(body io.ReadCloser, key string, errorKey string) error {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		var out map[string]interface{}

		err := json.Unmarshal(scanner.Bytes(), &out)

		if err != nil {
			return err
		}

		if stream, ok := out[key].(string); ok {
			if strings.HasSuffix(stream, "\n") {
				fmt.Printf(stream)
			} else {
				fmt.Println(stream)
			}
		}
		if errorMsg, ok := out[errorKey].(string); ok {
			if strings.HasSuffix(errorMsg, "\n") {
				fmt.Printf(errorMsg)
			} else {
				fmt.Println(errorMsg)
			}
		}
	}

	return nil
}
