package koyeb

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/tui/input"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/tui/simple_list"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy",
		Args:  cobra.ExactArgs(0),
		RunE:  WithCLIContext(deploy),
	}
	return deployCmd
}

func deploy(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	createArgs := []string{"service", "create"}

	var app koyeb.App
	appInput := input.New()
	appInput.SetPrompt("Choose an app name (leave blank to generate one)")
	appInput.SetSubmitHandler(func(value string) tea.Msg {
		// XXX: remove me, testing purpose to avoid creating apps
		if value == "" || strings.HasPrefix(value, "err") {
			return input.SubmitErrorMsg{Error: &errors.CLIError{
				What: "Not implemented",
				Additional: []string{
					"line 1",
					"line 2",
				},
				Why:      "An unxpected error occured",
				Solution: "Fix the issue and come back later",
			}}
		}
		app = koyeb.App{Name: &value, Id: koyeb.PtrString("00000000-0000-0000-0000-000000000000")}
		return input.SubmitOkMsg{}

		// createParams := koyeb.NewCreateAppWithDefaults()
		// createParams.SetName(value)
		// res, resp, err := ctx.Client.AppsApi.CreateApp(ctx.Context).App(*createParams).Execute()
		// if err != nil {
		// 	return input.SubmitErrorMsg{Error: errors.NewCLIErrorFromAPIError(
		// 		"Error while creating the app",
		// 		err,
		// 		resp,
		// 	)}
		// }
		// app = res.GetApp()
		// return input.SubmitOkMsg{}
	})

	if abort, err := appInput.Execute(); abort || err != nil {
		return err
	}
	log.Infof("Application %s created", app.GetName())
	createArgs = append(createArgs, "--app", app.GetName())

	var appType string
	appTypeInput := simple_list.New([]simple_list.SimpleListItem{
		{Id: "github", Name: "GitHub", Description: "Build and deploy a GitHub repository"},
		{Id: "docker", Name: "Docker", Description: "Deploy an existing Docker image"},
	})
	appTypeInput.SetSubmitHandler(func(selected simple_list.SimpleListItem) tea.Msg {
		appType = selected.Id
		return simple_list.SubmitOkMsg{}
	})
	appTypeInput.SetPrompt("Select your deployment method")
	if abort, err := appTypeInput.Execute(); abort || err != nil {
		return err
	}

	var err error

	switch appType {
	case "docker":
		createArgs, err = deployDocker(ctx, app, createArgs)
		if err != nil {
			return err
		}
	case "git":
		createArgs, err = deployGit(ctx, app, createArgs)
		if err != nil {
			return err
		}
	}

	if abort, err := createService(ctx, app, createArgs); abort || err != nil {
		return err
	}

	return nil
}

func createService(ctx *CLIContext, app koyeb.App, args []string) (bool, error) {
	cmd := &cobra.Command{}

	// Add all the service flags to the dummy command
	addServiceDefinitionFlags(cmd.Flags())

	// Give the arguments provided interactively
	cmd.SetArgs(args)
	cmd.ParseFlags(args)

	// Prepare the create service request
	createService := koyeb.NewCreateServiceWithDefaults()
	createService.SetAppId(app.GetId())

	createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

	err := parseServiceDefinitionFlags(cmd.Flags(), createDefinition)
	if err != nil {
		return false, err
	}
	createService.SetDefinition(*createDefinition)

	body, err := yaml.Marshal(createService)
	if err != nil {
		return false, err
	}

	fmt.Printf("%s\n", body)

	fileInput := input.New()
	fileInput.SetPrompt("Create the service with this definition?")

	if abort, err := fileInput.Execute(); abort || err != nil {
		return abort, err
	}

	return false, nil
}

func deployDocker(ctx *CLIContext, app koyeb.App, args []string) ([]string, error) {
	var image string
	var tag string

	imageInput := input.New()
	imageInput.SetPrompt("Enter the Docker image to deploy (e.g. nginx:latest, or private.registry.com/nginx:latest)")
	imageInput.SetSubmitHandler(func(value string) tea.Msg {
		parts := strings.Split(value, ":")
		if (len(parts) == 1 && parts[0] == "") ||
			(len(parts) == 2 && (parts[0] == "" || parts[1] == "")) {

			return input.SubmitErrorMsg{Error: fmt.Errorf("Invalid image format. Expected [registry/]image[:tag]")}
		}
		switch len(parts) {
		case 1:
			image = value
			tag = "latest"
		case 2:
			image = parts[0]
			tag = parts[1]
		}
		return input.SubmitOkMsg{}
	})

	if abort, err := imageInput.Execute(); abort || err != nil {
		return args, err
	}
	args = append(args, "--docker", fmt.Sprintf("%s:%s", image, tag))
	return args, nil
}

func deployGit(ctx *CLIContext, app koyeb.App, args []string) ([]string, error) {
	return args, nil
}
