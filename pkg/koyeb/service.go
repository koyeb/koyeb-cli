package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

type GetServiceReply struct {
	koyeb.GetServiceReply
}

func (a *GetServiceReply) MarshalBinary() ([]byte, error) {
	return a.GetServiceReply.GetService().MarshalJSON()
}

func (a *GetServiceReply) Title() string {
	return "Service"
}

func (a *GetServiceReply) Headers() []string {
	return []string{"id", "name", "version", "status", "updated_at"}
}

func (a *GetServiceReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.GetService()
	fields := map[string]string{}
	for _, field := range a.Headers() {
		switch field {
		case "status":
			fields[field] = GetField(item, "state.status")
		default:
			fields[field] = GetField(item, field)
		}
	}
	res = append(res, fields)
	return res
}

type ListServicesReply struct {
	koyeb.ListServicesReply
}

func (a *ListServicesReply) Title() string {
	return "Services"
}

func (a *ListServicesReply) MarshalBinary() ([]byte, error) {
	return a.ListServicesReply.MarshalJSON()
}

func (a *ListServicesReply) Headers() []string {
	return []string{"id", "name", "version", "status", "updated_at"}
}

func (a *ListServicesReply) Fields() []map[string]string {
	res := []map[string]string{}
	for _, item := range a.GetServices() {
		fields := map[string]string{}
		for _, field := range a.Headers() {
			switch field {
			case "status":
				fields[field] = GetField(item, "state.status")
			default:
				fields[field] = GetField(item, field)
			}
		}
		res = append(res, fields)
	}
	return res
}
