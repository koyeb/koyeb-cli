package koyeb

import (
	"fmt"
	"os"
	"path"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/replicate/cog/pkg/config"
	"github.com/replicate/cog/pkg/dockerfile"
)

const (
	buildUseCudaBaseImage string = "auto"
	useCogBaseImage       bool   = true
)

type cogAddon struct {
	dockerfileExists bool
	cleanupFuncs     []func() error
	directory        string
}

func (c *cogAddon) PreDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	archive, ok := definition.GetArchiveOk()
	if !ok {
		return errors.New("cog addon only supports archive deployments")
	}

	if _, ok := archive.GetBuildpackOk(); ok {
		return errors.New("cog addon only supports archive with docker deployments")
	}
	c.patchArchive(archive)
	err := c.launchCog(ctx, c.directory)
	if err != nil {
		return err
	}
	return nil
}

func (c *cogAddon) patchArchive(archive *koyeb.ArchiveSource) {
	docker := archive.GetDocker()
	docker.SetDockerfile("Dockerfile.cog")
	archive.SetDocker(docker)
}

func (c *cogAddon) launchCog(ctx *CLIContext, projectDirFlag string) error {
	cfg, projectDir, err := config.GetConfig(projectDirFlag)
	if err != nil {
		return err
	}

	generator, err := dockerfile.NewGenerator(cfg, projectDir)
	if err != nil {
		return fmt.Errorf("Error creating Dockerfile generator: %w", err)
	}
	c.appendCleanupFunc(func() error {
		if err := generator.Cleanup(); err != nil {
			log.Warnf("Error cleaning up after build: %v", err)
		}
		return err
	})

	generator.SetUseCudaBaseImage(buildUseCudaBaseImage)
	generator.SetUseCogBaseImage(useCogBaseImage)
	dockerfile, err := generator.GenerateDockerfileWithoutSeparateWeights()
	if err != nil {
		return err
	}
	log.Debugf("Generated Dockerfile:\n%s", dockerfile)
	err = os.WriteFile(path.Join(c.directory, "Dockerfile.cog"), []byte(dockerfile), 0o644)
	if err != nil {
		return err
	}
	return nil
}

func (c *cogAddon) appendCleanupFunc(f func() error) {
	c.cleanupFuncs = append(c.cleanupFuncs, f)
}

func (c *cogAddon) PostDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	return nil
}

func (c *cogAddon) Setup(ctx *CLIContext, dir string) error {
	c.directory = dir
	path := path.Join(dir, "Dockerfile.cog")
	if _, err := os.Stat(path); err == nil {
		c.dockerfileExists = true
		log.Infof("Dockerfile.cog already exists, skipping")
		return nil
	}
	return nil
}

func (c *cogAddon) Cleanup(ctx *CLIContext) error {
	if !c.dockerfileExists {
		path := path.Join(c.directory, "Dockerfile.cog")
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}

	for _, f := range c.cleanupFuncs {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}
