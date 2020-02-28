package koyeb

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
)

func isYaml(file string) bool {
	if strings.HasSuffix(file, ".yaml") {
		return true
	} else if strings.HasSuffix(file, ".yml") {
		return true
	}
	return false
}

func isJson(file string) bool {
	if strings.HasSuffix(file, ".json") {
		return true
	}
	return false
}

func parseFile(file string, item interface{}) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if isYaml(file) {
		d, err := yaml.YAMLToJSON(data)
		if err != nil {
			return err
		}

		return json.Unmarshal(d, item)
	} else if isJson(file) {
		return json.Unmarshal(data, item)
	}
	return errors.New("Unknown format")
}
