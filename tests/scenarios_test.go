package scenarios_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infratographer/krakend-endpoints-tool/cmd"
)

//nolint:paralleltest // go test panics when using Setenv with Parallel
func TestHappyPath(t *testing.T) {
	endpointsDir := "scenarios/happy-path"
	cfg := "scenarios/happy-path/krakend.tmpl"
	outf := filepath.Join(t.TempDir(), "krakend.tmpl")
	fcEnable, vhost := false, false
	err := cmd.Generate(endpointsDir, cfg, outf, "$ENDPOINTS$", vhost, fcEnable)
	assert.NoError(t, err, "should not fail")
}

//nolint:paralleltest // go test panics when using Setenv with Parallel
func TestHappyFlexibleConfig(t *testing.T) {
	t.Setenv("FC_SETTINGS", "../internal/flexibleconfig/testData/settings/dev")
	t.Setenv("FC_PARTIALS", "../internal/flexibleconfig/testData/partials")

	endpointsDir := "../internal/flexibleconfig/testData/endpoints"
	cfg := "scenarios/flexibleconfig/krakend.tmpl"
	outf := filepath.Join(t.TempDir(), "krakend.tmpl")
	fcEnable, vhost := true, false
	err := cmd.Generate(endpointsDir, cfg, outf, "$ENDPOINTS$", vhost, fcEnable)
	require.Nil(t, err, "should not fail")

	buf, err := os.ReadFile(outf)
	require.Nil(t, err)
	assert.Contains(t, string(buf), `"method": "GET"`, "GET partial.tmpl should be templated")
	assert.Contains(t, string(buf), `"timeout": "3s"`, "timeout setting should be templated")
	assert.Contains(t, string(buf), `"cache_ttl": "3s"`, "cache_ttl setting should be templated")
	assert.Contains(t, string(buf), `"output_encoding": "json"`, "output_encoding setting should be templated")
	assert.Contains(t, string(buf), `{{ env "KRAKEND_PORT"}}`, "env templating should not be replaced")
	assert.Contains(t, string(buf), `{{ env "TEST_API_URL"}}`, "env templating should not be replaced")
}
