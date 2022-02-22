package koyeb

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
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

func loadYaml(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	if isYaml(file) {
		return string(data), nil
	}
	return "", errors.New("Unknown format")
}

func watchDeployment(h *ServiceHandler, deploymentId string) {
	numAttemptsOnUnhealthy := 12
	retryCount := 0
	retryInterval := 5

	for {
		res, resp, err := h.client.DeploymentsApi.GetDeployment(h.ctx, deploymentId).Execute()
		if err != nil {
			fatalApiError(err, resp)
		}
		currentStatus := res.Deployment.GetStatus()

		log.Infof("Service deployment in progress. Deployment status is %s. Next update in %d seconds.", currentStatus, retryInterval)

		if currentStatus == koyeb.DEPLOYMENTSTATUS_ERROR || currentStatus == koyeb.DEPLOYMENTSTATUS_HEALTHY || retryCount >= numAttemptsOnUnhealthy {
			break
		} else if currentStatus == koyeb.DEPLOYMENTSTATUS_UNHEALTHY {
			retryCount += 1
			retryInterval = 10
			time.Sleep(time.Duration(retryInterval) * time.Second)
		} else {
			time.Sleep(time.Duration(retryInterval) * time.Second)
		}
	}
}
