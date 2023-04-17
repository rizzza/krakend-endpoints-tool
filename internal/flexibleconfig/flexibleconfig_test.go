package flexibleconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexibleConfig(t *testing.T) {
	t.Run("substitutes settings", func(t *testing.T) {
		cfg := Config{
			SettingsPath: "./testData/settings/dev",
		}

		p, err := NewTemplateParser(cfg)
		require.Nil(t, err)

		buf, err := p.Parse("./testData/endpoints/test_api_v1.yaml")
		require.Nil(t, err)
		assert.NotNil(t, buf)
		assert.Contains(t, buf.String(), "&hosts [http://host.docker.internal:7608]")
	})

	t.Run("substitutes partials", func(t *testing.T) {
		cfg := Config{
			PartialsPath: "./testData/partials",
		}

		p, err := NewTemplateParser(cfg)
		require.Nil(t, err)

		buf, err := p.Parse("./testData/endpoints/test_api_v1.yaml")
		require.Nil(t, err)
		assert.NotNil(t, buf)
		assert.Contains(t, buf.String(), "method: GET")
	})
}
