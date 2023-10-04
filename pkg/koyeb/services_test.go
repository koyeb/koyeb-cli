package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetGitSourceBuilder(t *testing.T) {
	tests := map[string]struct {
		args     []string
		source   *koyeb.GitSource
		expected *koyeb.GitSource
	}{
		"invalid_git_builder": {
			args:     []string{"--git-builder", "xxx"},
			source:   &koyeb.GitSource{},
			expected: nil,
		},
		"source_buildpack_no_arg": {
			args: []string{},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_docker_no_arg": {
			args: []string{},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_buildpack_set_builder_to_buildpack": {
			args: []string{"--git-builder", "buildpack"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_buildpack_set_builder_to_docker": {
			args: []string{"--git-builder", "docker"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_docker_set_builder_to_buildpack": {
			args: []string{"--git-builder", "buildpack"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_docker_set_builder_to_docker": {
			args: []string{"--git-builder", "docker"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_buildpack_set_docker_args": {
			args: []string{"--git-docker-dockerfile", "Dockerfile.dev"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: nil,
		},
		"source_buildpack_set_buildpack_args": {
			args: []string{"--git-buildpack-run-command", "run", "--git-buildpack-build-command", "build", "--privileged"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{
					Privileged: koyeb.PtrBool(false),
				},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{
					RunCommand:   koyeb.PtrString("run"),
					BuildCommand: koyeb.PtrString("build"),
					Privileged:   koyeb.PtrBool(true),
				},
			},
		},
		"source_docker_set_buildpack_args": {
			args: []string{"--git-buildpack-run-command", "run", "--git-buildpack-build-command", "build", "--privileged"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: nil,
		},
		"source_docker_set_docker_args": {
			args: []string{
				"--git-docker-command", "cmd",
				"--git-docker-args", "arg1", "--git-docker-args", "arg2",
				"--git-docker-entrypoint", "entrypoint.sh", "--git-docker-entrypoint", "entrypoint-arg",
				"--git-docker-dockerfile", "Dockerfile.dev",
				"--git-docker-target", "dev",
				"--privileged",
			},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{
					Command:    koyeb.PtrString("cmd"),
					Args:       []string{"arg1", "arg2"},
					Entrypoint: []string{"entrypoint.sh", "entrypoint-arg"},
					Dockerfile: koyeb.PtrString("Dockerfile.dev"),
					Target:     koyeb.PtrString("dev"),
					Privileged: koyeb.PtrBool(true),
				},
			},
		},
		"source_buildpack_override_privileged_without_other_args": {
			args: []string{"--privileged=false"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{
					Privileged: koyeb.PtrBool(true),
				},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{
					Privileged: koyeb.PtrBool(false),
				},
			},
		},
		"source_docker_override_privileged_without_other_args": {
			args: []string{"--privileged"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{
					Privileged: koyeb.PtrBool(false),
				},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{
					Privileged: koyeb.PtrBool(true),
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}

			addServiceDefinitionFlags(cmd.Flags())

			cmd.SetArgs(tc.args)
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal()
			}

			ret, err := setGitSourceBuilder(cmd.Flags(), tc.source)

			if tc.expected != nil {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, ret)
			} else {
				assert.NotNil(t, err)
				assert.Nil(t, ret)
			}
		})
	}
}
