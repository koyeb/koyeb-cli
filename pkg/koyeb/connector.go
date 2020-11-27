package koyeb

import (
	"errors"
	"fmt"
	"github.com/go-openapi/swag"
	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/client/connectors"
	"github.com/koyeb/koyeb-cli/pkg/gen/kclient/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"strconv"
	"strings"
)

var (
	createConnectorCommand = &cobra.Command{
		Use:     "connectors [resource]",
		Aliases: []string{"connector"},
		Short:   "Create connectors",
		RunE:    createConnectors,
	}
	getConnectorCommand = &cobra.Command{
		Use:     "connectors [resource]",
		Aliases: []string{"connector"},
		Short:   "Get connectors",
		RunE:    getConnectors,
	}
	describeConnectorCommand = &cobra.Command{
		Use:     "connectors [resource]",
		Aliases: []string{"connector"},
		Short:   "Describe connectors",
		RunE:    getConnectors,
	}
	updateConnectorCommand = &cobra.Command{
		Use:     "connectors [resource]",
		Aliases: []string{"connector"},
		Short:   "Update connectors",
		RunE:    updateConnectors,
	}
	deleteConnectorCommand = &cobra.Command{
		Use:     "connectors [resource]",
		Aliases: []string{"connector"},
		Short:   "Delete connectors",
		RunE:    deleteConnectors,
	}
)

func deleteConnectors(cmd *cobra.Command, args []string) error {
	client := getApiClient().Connectors

	if len(args) > 0 {
		for _, arg := range args {
			p := connectors.NewConnectorsDeleteConnectorParams().WithIDOrName(arg)
			resp, err := client.ConnectorsDeleteConnector(p, getAuth())
			if err != nil {
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			log.Infof("Stack %s deleted", p.IDOrName)
		}
	}
	return nil
}

func updateConnectors(cmd *cobra.Command, args []string) error {
	client := getApiClient().Connectors
	var all StorageConnectorUpsertBody

	log.Debugf("Loading file %s", file)
	err := loadMultiple(file, &all, "connectors")
	if err != nil {
		er(err)
	}
	log.Debugf("Content loaded %v", all.Connectors)

	for _, cn := range all.Connectors {
		p := connectors.NewConnectorsUpdateConnectorParams().WithIDOrName(cn.Name).WithBody(cn)
		resp, err := client.ConnectorsUpdateConnector(p, getAuth())
		if err != nil {
			logApiError(err)
			continue
		}
		log.Debugf("got response: %v", resp)
	}
	return nil
}

func getConnectors(cmd *cobra.Command, args []string) error {
	client := getApiClient().Connectors
	format := "table"
	if cmd.Parent().Name() == "describe" {
		fmt.Printf("Connector:\n")
		format = "yaml"
	}

	if len(args) > 0 {
		var all = make([]models.StorageConnector, len(args))
		for i, arg := range args {
			p := connectors.NewConnectorsGetConnectorParams().WithIDOrName(arg)
			resp, err := client.ConnectorsGetConnector(p, getAuth())
			if err != nil {
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
			all[i] = *resp.GetPayload().Connector
		}
		render(ConnectorListDetails{Elts: all}, format)
	} else {
		var all []models.StorageConnectorListItem
		page := 0
		limit := 10
		offset := 0

		for {
			strOffset := strconv.Itoa(offset)
			strLimit := strconv.Itoa(limit)
			p := connectors.NewConnectorsListConnectorsParams().
				WithLimit(&strLimit).
				WithOffset(&strOffset)

			resp, err := client.ConnectorsListConnectors(p, getAuth())
			if err != nil {
				fatalApiError(err)
			}
			if all == nil {
				all = make([]models.StorageConnectorListItem, resp.Payload.Count)
			}
			log.Debugf("got response: %v", resp)
			for i, v := range resp.GetPayload().Connectors {
				all[i+offset] = *v
			}
			page += 1
			offset = page * limit
			if int64(offset) >= resp.GetPayload().Count {
				break
			}
		}
		render(ConnectorList{Elts: all}, format)
	}
	return nil
}

func createConnectors(cmd *cobra.Command, args []string) error {
	client := getApiClient().Connectors

	var all StorageConnectorUpsertBody

	if file != "" {
		log.Debugf("Loading file %s", file)
		err := loadMultiple(file, &all, "stacks")
		if err != nil {
			er(err)
		}
		log.Debugf("Content loaded %v", all.Connectors)

		for _, c := range all.Connectors {
			resp, err := client.ConnectorsNewConnector(connectors.NewConnectorsNewConnectorParams().WithBody(c), getAuth())
			if err != nil {
				logApiError(err)
				continue
			}
			log.Debugf("got response: %v", resp)
		}
	} else {
		return errors.New("Missing file")
	}

	return nil
}

type StorageConnectorUpsertBody struct {
	Connectors []*models.StorageConnectorUpsert `json:"connectors"`
}

func (a *StorageConnectorUpsertBody) Append(item interface{}) {
	connector := item.(*models.StorageConnectorUpsert)
	a.Connectors = append(a.Connectors, connector)
}

func (a *StorageConnectorUpsertBody) New() interface{} {
	return &models.StorageConnectorUpsert{}
}

type ConnectorListDetails struct {
	Elts []models.StorageConnector `json:"connectors"`
}

func (c ConnectorListDetails) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "type", "url", "created_at", "updated_at", "filter", "mapper"},
	}
	for _, item := range c.Elts {
		var fields []string
		for _, field := range res.headers {
			var res string
			if field == "url" {
				res = genFullUrl(string(item.Type), item.URL)
			} else if field == "filter" || field == "mapper" {
				if e := item.With[field]; e != nil {
					res = item.With[field].(string)
				} else {
					res = ""
				}
			} else {
				res = getField(item, field)
			}
			fields = append(fields, res)
		}
		res.fields = append(res.fields, fields)
	}
	return res
}

func (c ConnectorListDetails) MarshalBinary() ([]byte, error) {
	if len(c.Elts) == 1 {
		return swag.WriteJSON(c.Elts[0])
	} else {
		return swag.WriteJSON(c)
	}
}

type ConnectorList struct {
	Elts []models.StorageConnectorListItem `json:"connectors"`
}

func genFullUrl(kind, path string) string {
	r, err := url.Parse(apiurl)
	if err != nil {
		panic(err)
	}
	host := "connectors.prod.koyeb.com"
	if strings.HasPrefix(r.Host, "staging") {
		host = "connectors.staging.koyeb.com"
	}
	return fmt.Sprintf("https://%s/%s/%s", host, kind, path)
}

func (c ConnectorList) GetTable() TableInfo {
	res := TableInfo{
		headers: []string{"id", "name", "type", "url", "created_at", "updated_at"},
	}
	for _, item := range c.Elts {
		var fields []string
		for _, field := range res.headers {
			var res string
			// TODO remove once the API always returns full URLs
			if field == "url" && !strings.HasPrefix(item.URL, "https://") {
				res = genFullUrl(string(item.Type), item.URL)
			} else {
				res = getField(item, field)
			}
			fields = append(fields, res)
		}
		res.fields = append(res.fields, fields)
	}
	return res
}

func (c ConnectorList) MarshalBinary() ([]byte, error) {
	if len(c.Elts) == 1 {
		return swag.WriteJSON(c.Elts[0])
	} else {
		return swag.WriteJSON(c)
	}
}
