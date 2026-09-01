package migrate

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"
)

func buildEnvMap() map[string]string {
	env := make(map[string]string)
	for _, kvp := range os.Environ() {
		kvParts := strings.SplitN(kvp, "=", 2)
		env[kvParts[0]] = kvParts[1]
	}
	return env
}

// envMap returns the process environment as a template context. Computed
// once, synchronized: migrations are prefetched concurrently and each
// applyEnvironmentTemplate call executes its template on a goroutine, so an
// unsynchronized lazy cache is a concurrent map write.
var envMap = sync.OnceValue(buildEnvMap)

func applyEnvironmentTemplate(body io.ReadCloser) (io.ReadCloser, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	defer func() {
		_ = body.Close()
	}()

	tmpl, err := template.New("migration").Parse(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	r, w := io.Pipe()

	go func() {
		_ = tmpl.Execute(w, envMap())
		_ = w.Close()
	}()

	return r, nil
}
