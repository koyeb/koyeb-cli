package koyeb

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type CopyManager struct {
	Src *FileSpec
	Dst *FileSpec
}

func NewCopyManager(src, dst *FileSpec) (*CopyManager, error) {
	if err := ValidateTargets(src, dst); err != nil {
		return nil, err
	}

	return &CopyManager{
		Src: src,
		Dst: dst,
	}, nil
}

func (manager *CopyManager) Copy(ctx *CLIContext) error {
	if len(manager.Src.InstanceID) == 0 {
		return manager.copyToInstance(ctx)
	}
	return manager.copyFromInstance()
}

func (manager *CopyManager) copyToInstance(ctx *CLIContext) error {
	if err := manager.Src.Validate(ctx); err != nil {
		return err
	}

	if err := manager.Dst.Validate(ctx); err != nil {
		return err
	}

	reader, writer := io.Pipe()

	go func(srcPath string, writer io.WriteCloser) {
		defer writer.Close()

		err := Tar(srcPath, writer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to tar file: %s\n", err)
			os.Exit(1)
		}
	}(manager.Src.FilePath, writer)

	retCode, err := ctx.ExecClient.ExecWithStreams(
		ctx.Context,
		&StdStreams{
			Stdin:  reader,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
		ExecId{
			Id:   manager.Dst.InstanceID,
			Type: koyeb.EXECCOMMANDREQUESTIDTYPE_INSTANCE_ID,
		},
		[]string{"tar", "-C", manager.Dst.FilePath, "-xf", "-"},
	)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if retCode != 0 {
		os.Exit(retCode)
	}

	return nil
}

func (manager *CopyManager) copyFromInstance() error {
	return errors.New("not implemented")
}

func ValidateTargets(srcSpec, dstSpec *FileSpec) error {
	if srcSpec.InstanceID != "" && dstSpec.InstanceID != "" {
		return fmt.Errorf("one of src or dest must be a local file specification")
	}

	if srcSpec.InstanceID == "" && dstSpec.InstanceID == "" {
		return fmt.Errorf("one of src or dest must be an instance name")
	}

	if len(srcSpec.FilePath) == 0 || len(dstSpec.FilePath) == 0 {
		return errors.New("filepath cannot be empty")
	}

	return nil
}

var (
	errFileSpecDoesntMatchFormat = errors.New("filespec must match the canonical format: [instance:]file/path")
)

type FileSpec struct {
	InstanceID string
	FilePath   string
}

func (spec *FileSpec) validateLocalPath() error {
	if _, err := os.Stat(spec.FilePath); err != nil {
		return fmt.Errorf("filepath %s doesn't exist", spec.FilePath)
	}
	return nil
}

func (spec *FileSpec) validateRemotePath(ctx *CLIContext) error {
	retCode, err := ctx.ExecClient.ExecWithStreams(
		ctx.Context,
		&StdStreams{
			Stdin:  bytes.NewReader([]byte{}),
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
		ExecId{
			Id:   spec.InstanceID,
			Type: koyeb.EXECCOMMANDREQUESTIDTYPE_INSTANCE_ID,
		},
		[]string{"test", "-d", spec.FilePath},
	)
	if err != nil {
		return fmt.Errorf("remote filepath %s doesn't exist %v", spec.FilePath, err)
	}

	if retCode != 0 {
		os.Exit(retCode)
	}

	return err
}

func (spec *FileSpec) Validate(ctx *CLIContext) error {
	if len(spec.InstanceID) == 0 {
		return spec.validateLocalPath()
	}
	return spec.validateRemotePath(ctx)
}
