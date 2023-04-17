// flexibleconfig based on https://github.com/krakendio/krakend-flexibleconfig/blob/master/template.go
package flexibleconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

type FlexibleConfig struct {
	Vars     map[string]interface{}
	Partials string
	funcMap  template.FuncMap
}

type Config struct {
	SettingsPath string
	PartialsPath string
}

func NewTemplateParser(cfg Config) (*FlexibleConfig, error) {
	t := &FlexibleConfig{
		Partials: cfg.PartialsPath,
		Vars:     map[string]interface{}{},
	}

	if cfg.SettingsPath != "" {
		files, err := os.ReadDir(cfg.SettingsPath)
		if err != nil {
			files = []os.DirEntry{}
		}

		for _, settingsFile := range files {
			var v map[string]interface{}

			ext := filepath.Ext(settingsFile.Name())
			if ext != ".json" && ext != ".yaml" {
				return nil, fmt.Errorf("settings file %q read from %q is invalid: must be of type .json or .yaml", settingsFile.Name(), cfg.SettingsPath)
			}

			b, err := os.ReadFile(filepath.Join(cfg.SettingsPath, settingsFile.Name()))
			if err != nil {
				return nil, err
			}

			switch ext {
			case ".json":
				if err := json.Unmarshal(b, &v); err != nil {
					return nil, fmt.Errorf("faield to unmarshal %s: %w", settingsFile.Name(), err)
				}
				t.Vars[strings.TrimSuffix(filepath.Base(settingsFile.Name()), ".json")] = v
			case ".yaml":
				if err := yaml.Unmarshal(b, &v); err != nil {
					return nil, fmt.Errorf("faield to unmarshal %s: %w", settingsFile.Name(), err)
				}
				t.Vars[strings.TrimSuffix(filepath.Base(settingsFile.Name()), ".yaml")] = v
			}
		}
	}

	t.funcMap = sprig.GenericFuncMap()
	t.funcMap["marshal"] = t.marshal
	t.funcMap["include"] = t.include

	return t, nil
}

func (t FlexibleConfig) Parse(endpointFile string) (bytes.Buffer, error) {
	var (
		buf bytes.Buffer
	)

	tmpl, err := template.New("endpoint").Funcs(t.funcMap).ParseFiles(endpointFile)
	if err != nil {
		return buf, fmt.Errorf("failed to parse endpoint file %s: %w", endpointFile, err)
	}

	err = tmpl.ExecuteTemplate(&buf, filepath.Base(endpointFile), t.Vars)
	if err != nil {
		return buf, fmt.Errorf("failure executing template: %w", err)
	}

	return buf, nil
}

func (FlexibleConfig) marshal(v interface{}) string {
	a, _ := json.Marshal(v)
	return string(a)
}

func (t FlexibleConfig) include(v interface{}) string {
	a, _ := os.ReadFile(path.Join(t.Partials, v.(string)))
	return string(a)
}
