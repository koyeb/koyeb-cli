package koyeb

import (
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

type ClaimReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.PoolClaimReply
	full   bool
}

func NewClaimReply(mapper *idmapper.Mapper, value *koyeb.PoolClaimReply, full bool) *ClaimReply {
	return &ClaimReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ClaimReply) Title() string {
	return "Claim"
}

func (r *ClaimReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

// Claim IDs render in full: the claim API only accepts complete UUIDs and
// claims have no client-side ID resolution.
func (r *ClaimReply) Headers() []string {
	return []string{"claim_id", "service_id", "prewarmed"}
}

func (r *ClaimReply) Fields() []map[string]string {
	fields := map[string]string{
		"claim_id":   r.value.GetClaimId(),
		"service_id": renderer.FormatID(r.value.GetServiceId(), r.full),
		"prewarmed":  fmt.Sprintf("%t", r.value.GetPrewarmed()),
	}

	return []map[string]string{fields}
}

type ListPoolClaimsReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.ListPoolClaimReply
	full   bool
}

func NewListPoolClaimsReply(mapper *idmapper.Mapper, value *koyeb.ListPoolClaimReply, full bool) *ListPoolClaimsReply {
	return &ListPoolClaimsReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (ListPoolClaimsReply) Title() string {
	return "Claims"
}

func (r *ListPoolClaimsReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *ListPoolClaimsReply) Headers() []string {
	return []string{"id", "request_id", "status", "service_id", "created_at"}
}

func (r *ListPoolClaimsReply) Fields() []map[string]string {
	claims := r.value.GetClaims()
	resp := make([]map[string]string, 0, len(claims))

	for _, claim := range claims {
		fields := map[string]string{
			"id":         claim.GetId(),
			"request_id": claim.GetRequestId(),
			"status":     string(claim.GetStatus()),
			"service_id": renderer.FormatID(claim.GetServiceId(), r.full),
			"created_at": renderer.FormatTime(claim.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}

type GetPoolClaimReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetPoolClaimReply
	full   bool
}

func NewGetPoolClaimReply(mapper *idmapper.Mapper, value *koyeb.GetPoolClaimReply, full bool) *GetPoolClaimReply {
	return &GetPoolClaimReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetPoolClaimReply) Title() string {
	return "Claim"
}

func (r *GetPoolClaimReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *GetPoolClaimReply) Headers() []string {
	return []string{"id", "pool_id", "request_id", "status", "service_id", "created_at", "fulfilled_at"}
}

func (r *GetPoolClaimReply) Fields() []map[string]string {
	claim := r.value.GetClaim()
	fields := map[string]string{
		"id":           claim.GetId(),
		"pool_id":      renderer.FormatID(claim.GetPoolId(), r.full),
		"request_id":   claim.GetRequestId(),
		"status":       string(claim.GetStatus()),
		"service_id":   renderer.FormatID(claim.GetServiceId(), r.full),
		"created_at":   renderer.FormatTime(claim.GetCreatedAt()),
		"fulfilled_at": renderer.FormatTime(claim.GetFulfilledAt()),
	}

	return []map[string]string{fields}
}

func newPoolClaimsCmd() *cobra.Command {
	claimsCmd := &cobra.Command{
		Use:   "claims ACTION",
		Short: "Manage pool claims",
	}
	claimsCmd.AddCommand(newPoolClaimsListCmd())
	claimsCmd.AddCommand(newPoolClaimsGetCmd())

	return claimsCmd
}
