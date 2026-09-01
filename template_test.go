package migrate

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

func Test_applyEnvironmentTemplate(t *testing.T) {
	_ = os.Setenv("WAREHOUSE_DB", "WH_STAGING")
	envMap = sync.OnceValue(buildEnvMap)

	migration := io.NopCloser(bytes.NewBuffer([]byte(`SELECT * FROM {{.WAREHOUSE_DB}}.STD.INVOICES`)))

	gotReader, err := applyEnvironmentTemplate(migration)
	if err != nil {
		t.Fatalf("expected no error applying template")
	}

	gotBytes, err := io.ReadAll(gotReader)
	if err != nil {
		t.Fatalf("expected no error reading")
	}
	got := string(gotBytes)
	want := `SELECT * FROM WH_STAGING.STD.INVOICES`
	if got != want {
		t.Fatalf("expected [%s] but got [%s]", want, got)
	}
}

// Migrations are prefetched concurrently, so the first envMap() calls can
// race; run under -race this catches an unsynchronized cache.
func Test_applyEnvironmentTemplate_concurrent(t *testing.T) {
	envMap = sync.OnceValue(buildEnvMap)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			migration := io.NopCloser(bytes.NewBufferString(`{{.PATH}}`))
			gotReader, err := applyEnvironmentTemplate(migration)
			if err != nil {
				t.Error(err)
				return
			}
			_, _ = io.ReadAll(gotReader)
		}()
	}
	wg.Wait()
}
