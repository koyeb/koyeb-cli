package koyeb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"

	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func (h *ServiceHandler) ShowUnappliedChanges(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceName, err := h.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	service, err := h.ResolveServiceArgs(ctx, serviceName)
	if err != nil {
		return err
	}

	var stashedDeployment *koyeb.DeploymentListItem
	var previousNonStashedDeployment *koyeb.DeploymentListItem

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	// To display unapplied changes, we need to compare the last deployment of
	// the service (which needs to be stashed0 with the next non-stashed
	// deployment.
	for {
		res, resp, err := ctx.Client.DeploymentsApi.
			ListDeployments(ctx.Context).
			ServiceId(service).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error while listing the deployments of the service `%s`", serviceName),
				err,
				resp,
			)
		}

		// After the first iteration, stashedDeployment is set so we don't enter this condition again
		if stashedDeployment == nil {
			if len(res.Deployments) == 0 {
				return &errors.CLIError{
					What:       "Unable to show unapplied changes",
					Why:        "we couldn't find the latest deployment of your service",
					Additional: []string{},
					Orig:       nil,
					Solution:   "Use `koyeb service update` to update your service",
				}
			}

			// Last deployment is not stashed, render an empty diff
			if res.Deployments[0].GetStatus() != koyeb.DEPLOYMENTSTATUS_STASHED {
				showUnappliedChangesReply := NewShowDeploymentsDiff(nil, nil)
				ctx.Renderer.Render(showUnappliedChangesReply)
				return nil
			}
			stashedDeployment = &res.Deployments[0]
		}

		for _, deployment := range res.Deployments {
			if deployment.GetStatus() != koyeb.DEPLOYMENTSTATUS_STASHED {
				previousNonStashedDeployment = &deployment
				break
			}
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	lhs, _ := json.Marshal(previousNonStashedDeployment.GetDefinition())
	rhs, _ := json.Marshal(stashedDeployment.GetDefinition())

	diff, err := gojsondiff.New().Compare(lhs, rhs)
	if err != nil {
		return &errors.CLIError{
			What:       "Unable to show unapplied changes",
			Why:        "unable to create the JSON diff",
			Additional: []string{},
			Orig:       err,
			Solution:   "Please, create an issue on https://github.com/koyeb/koyeb-cli/issues/new and provide your service ID",
		}
	}

	showUnappliedChangesReply := NewShowDeploymentsDiff(diff, lhs)
	ctx.Renderer.Render(showUnappliedChangesReply)
	return nil
}

type ShowDeploymentsDiff struct {
	diff gojsondiff.Diff
	lhs  []byte
}

func NewShowDeploymentsDiff(diff gojsondiff.Diff, lhs []byte) *ShowDeploymentsDiff {
	return &ShowDeploymentsDiff{diff, lhs}
}

func (ShowDeploymentsDiff) Title() string {
	return "Unapplied changes"
}

func (r *ShowDeploymentsDiff) MarshalBinary() ([]byte, error) {
	if r.diff == nil {
		return nil, nil
	}

	formatter := formatter.NewDeltaFormatter()
	diffString, _ := formatter.Format(r.diff)
	return []byte(diffString), nil
}

func (r *ShowDeploymentsDiff) Headers() []string {
	return []string{"diff"}
}

func (r *ShowDeploymentsDiff) Fields() []map[string]string {
	var diffString string

	if r.diff == nil || len(r.diff.Deltas()) == 0 {
		diffString = "No unapplied changes"
	} else {
		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		var aJson map[string]interface{}
		_ = json.Unmarshal(r.lhs, &aJson)

		formatter := formatter.NewAsciiFormatter(aJson, config)
		diffString, _ = formatter.Format(r.diff)
	}

	fields := map[string]string{
		"diff": diffString,
	}

	resp := []map[string]string{fields}
	return resp
}
