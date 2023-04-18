package flexibleconfig

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexibleConfig(t *testing.T) {
	b, err := os.ReadFile("./testData/endpoints/test_api_v1.yaml")
	require.Nil(t, err)

	t.Run("templates settings config", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			SettingsPath: "./testData/settings/dev",
		}

		p, err := NewTemplateParser(cfg)
		require.Nil(t, err)

		buf, err := p.Parse(bytes.NewBuffer(b))
		require.Nil(t, err)
		assert.NotNil(t, buf)
		assert.Contains(t, buf.String(), "&hosts [http://host.docker.internal:7608]")
	})

	t.Run("templates partials config", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			PartialsPath: "./testData/partials",
		}

		p, err := NewTemplateParser(cfg)
		require.Nil(t, err)

		buf, err := p.Parse(bytes.NewBuffer(b))
		require.Nil(t, err)
		assert.NotNil(t, buf)
		assert.Contains(t, buf.String(), "method: GET")
	})
}
