package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/infratographer/krakend-endpoints-tool/internal/flexibleconfig"
	"gopkg.in/yaml.v3"
)

var (
	debugMode  *bool
	exceptions = []string{"api.test.v1"}
)

type endpointType string

type execFunc func(path string, t endpointType, obj any, prefix string) error

const (
	arrayEndpoint   endpointType = "array"
	objectEndpoint  endpointType = "object"
	unknownEndpoint endpointType = "unknown"
)

func debug(frmt string, args ...interface{}) {
	if *debugMode {
		fmt.Printf(frmt, args...)
		fmt.Printf("\n")
	}
}

func info(frmt string, args ...interface{}) {
	fmt.Printf(frmt, args...)
	fmt.Printf("\n")
}

// parses an endpoint and returns the type of endpoint
// and the parsed JSON object.
func getEndpointAndType(path string) (endpointType, any, error) {
	if os.Getenv("FC_ENABLE") == "1" {
		// template requested endpoint file
		ext := filepath.Ext(path)
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("KrakenD_parsed_template_*%s", ext))
		if err != nil {
			return unknownEndpoint, nil, err
		}

		defer func() {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
		}()

		p, err := flexibleconfig.NewTemplateParser(flexibleconfig.Config{
			SettingsPath: os.Getenv("FC_SETTINGS"),
			PartialsPath: os.Getenv("FC_PARTIALS"),
		})

		if err != nil {
			return unknownEndpoint, nil, err
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return unknownEndpoint, nil, err
		}

		buf, err := p.Parse(bytes.NewBuffer(b))
		if err != nil {
			return unknownEndpoint, nil, err
		}

		_, err = tmpFile.Write(buf.Bytes())
		if err != nil {
			return unknownEndpoint, nil, err
		}

		if err := tmpFile.Close(); err != nil {
			return unknownEndpoint, nil, err
		}

		path = tmpFile.Name()
	}

	// open the file
	f, err := os.Open(path)
	if err != nil {
		return unknownEndpoint, nil, err
	}

	defer f.Close()

	// check if the endpoint is an object
	if isObj, obj := castEndpointTypeObject(f); isObj {
		return objectEndpoint, obj, nil
	}

	if _, err := f.Seek(0, 0); err != nil {
		return unknownEndpoint, nil, err
	}

	// check if the endpoint is an array
	if isArray, arr := castEndpointTypeArray(f); isArray {
		return arrayEndpoint, arr, nil
	}

	return unknownEndpoint, nil, nil
}

func castEndpointTypeArray(f *os.File) (ok bool, endpointArray any) {
	var decoder interface{ Decode(any) error }

	if strings.HasSuffix(f.Name(), ".yaml") {
		decoder = yaml.NewDecoder(f)
	} else {
		decoder = json.NewDecoder(f)
	}

	arrayProbe := []any{}

	err := decoder.Decode(&arrayProbe)
	if err != nil {
		return false, nil
	}

	return true, arrayProbe
}

func castEndpointTypeObject(f *os.File) (ok bool, endpointObj any) {
	var decoder interface{ Decode(any) error }

	if strings.HasSuffix(f.Name(), ".yaml") {
		decoder = yaml.NewDecoder(f)
	} else {
		decoder = json.NewDecoder(f)
	}

	objProbe := map[string]any{}

	err := decoder.Decode(&objProbe)
	if err != nil {
		return false, nil
	}

	return true, objProbe
}

// checks whether a file or directory should be skipped
// when processing endpoints.
func shouldBeSkipped(path string, d fs.DirEntry, exceptions []string) bool {
	// skip directories
	if d.IsDir() {
		return true
	}

	// skip hidden files
	if strings.HasPrefix(d.Name(), ".") {
		return true
	}

	if !strings.HasSuffix(d.Name(), ".json") && !strings.HasSuffix(d.Name(), ".yaml") {
		return true
	}

	for _, e := range exceptions {
		if strings.Contains(path, e) {
			return true
		}
	}

	return false
}

func WalkEndpoints(endpoints string, exceptions []string, f execFunc) error {
	return filepath.WalkDir(endpoints, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if shouldBeSkipped(path, d, exceptions) {
			debug("skipping %s", yellow(path))
			return nil
		}

		debug("processing %s", green(path))

		typ, obj, err := getEndpointAndType(path)
		if err != nil {
			return err
		}

		dir := filepath.Dir(path)
		prefix := strings.TrimPrefix(dir, endpoints)

		return f(path, typ, obj, prefix)
	})
}

func getOutputFile(outf string) (*os.File, error) {
	if outf == "" {
		return nil, fmt.Errorf("output file is required")
	}

	if outf == "-" {
		return os.Stdout, nil
	}

	return os.Create(outf)
}

func persistJSON(w io.Writer, endpts []any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(endpts)
}
