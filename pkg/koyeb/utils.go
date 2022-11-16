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
	now := time.Now()
	prevStatus := koyeb.DEPLOYMENTSTATUS_PENDING
	retryInterval := 5 * time.Second
	timeoutAt := time.Minute * 10

	for time.Since(now) < timeoutAt {
		res, resp, err := h.client.DeploymentsApi.GetDeployment(h.ctx, deploymentId).Execute()
		if err != nil {
			fatalApiError(err, resp)
		}
		currentStatus := res.Deployment.GetStatus()

		log.Infof("Service deployment in progress. Deployment status is %q. Next update in %s.", currentStatus, retryInterval)

		if currentStatus == koyeb.DEPLOYMENTSTATUS_ERROR || currentStatus == koyeb.DEPLOYMENTSTATUS_HEALTHY {
			if currentStatus == koyeb.DEPLOYMENTSTATUS_ERROR {
				log.Infof("Service deployment failed. Please check the logs.")
			}
			return
		} else if currentStatus == koyeb.DEPLOYMENTSTATUS_UNHEALTHY && prevStatus != koyeb.DEPLOYMENTSTATUS_UNHEALTHY {
			timeoutAt = time.Minute * 5
			now = time.Now()
		}
		time.Sleep(retryInterval)
		prevStatus = currentStatus
	}

	log.Infof("Service deployment didn't pass health checks. Last status was %q", prevStatus)
}
