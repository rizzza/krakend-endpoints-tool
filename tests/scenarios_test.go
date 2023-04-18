package scenarios_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infratographer/krakend-endpoints-tool/cmd"
)

//nolint:paralleltest // environment may clobber
func TestHappyPath(t *testing.T) {
	endpointsDir := "scenarios/happy-path"
	cfg := "scenarios/happy-path/krakend.tmpl"
	outf := filepath.Join(t.TempDir(), "krakend.tmpl")
	err := cmd.Generate(endpointsDir, cfg, outf, "$ENDPOINTS$", false)
	assert.NoError(t, err, "should not fail")
}

//nolint:paralleltest // environment may clobber
func TestHappyFlexibleConfig(t *testing.T) {
	t.Setenv("FC_ENABLE", "1")
	t.Setenv("FC_SETTINGS", "../internal/flexibleconfig/testData/settings/dev")
	t.Setenv("FC_PARTIALS", "../internal/flexibleconfig/testData/partials")

	endpointsDir := "../internal/flexibleconfig/testData/endpoints"
	cfg := "scenarios/flexibleconfig/krakend.tmpl"
	outf := filepath.Join(t.TempDir(), "krakend.tmpl")
	err := cmd.Generate(endpointsDir, cfg, outf, "$ENDPOINTS$", false)
	require.Nil(t, err, "should not fail")

	buf, err := os.ReadFile(outf)
	require.Nil(t, err)
	assert.Contains(t, string(buf), `"method": "GET"`)
	assert.Contains(t, string(buf), `"timeout": "3s"`)
	assert.Contains(t, string(buf), `"cache_ttl": "3s"`)
	assert.Contains(t, string(buf), `"output_encoding": "json"`)
	assert.Contains(t, string(buf), `"port": {{ env "KRAKEND_PORT"}}`, "env templating should not be replaced")
}
