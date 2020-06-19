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

func loadMultiple(file string, item UpdateApiResources, root string) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	buffers := strings.Split(string(raw), "---")

	for _, buf := range buffers {
		data := []byte(buf)
		rootDict := make(map[string]interface{})

		err = tryLoad(file, data, &rootDict)
		if err != nil {
			return err
		}
		if _, ok := rootDict[root]; ok {
			err = tryLoad(file, data, item)
			if err != nil {
				return err
			}
			return nil
		}

		ne := item.New()
		err = tryLoad(file, data, ne)
		if err != nil {
			return err
		}
		item.Append(ne)
	}

	return nil
}

func tryLoad(file string, data []byte, item interface{}) error {
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
