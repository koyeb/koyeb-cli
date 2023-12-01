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
	isCopyToInstance := len(manager.Src.InstanceID) == 0

	if err := manager.Src.Validate(ctx, isCopyToInstance); err != nil {
		return err
	}

	if err := manager.Dst.Validate(ctx, isCopyToInstance); err != nil {
		return err
	}

	if isCopyToInstance {
		return manager.copyToInstance(ctx)
	}
	return manager.copyFromInstance(ctx)
}

func (manager *CopyManager) copyToInstance(ctx *CLIContext) error {
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

func (manager *CopyManager) copyFromInstance(ctx *CLIContext) error {
	reader, writer := io.Pipe()

	go func(dstPath string, reader io.ReadCloser) {
		defer reader.Close()

		err := Untar(dstPath, reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to tar file: %s\n", err)
			os.Exit(1)
		}
	}(manager.Dst.FilePath, reader)

	pathDir := path.Dir(manager.Src.FilePath)
	pathBase := path.Base(manager.Src.FilePath)
	retCode, err := ctx.ExecClient.ExecWithStreams(
		ctx.Context,
		&StdStreams{
			Stdin:  bytes.NewReader([]byte{}),
			Stdout: writer,
			Stderr: os.Stderr,
		},
		ExecId{
			Id:   manager.Src.InstanceID,
			Type: koyeb.EXECCOMMANDREQUESTIDTYPE_INSTANCE_ID,
		},
		[]string{"tar", "-C", pathDir, "-czf", "-", pathBase},
	)
	if err != nil {
		return fmt.Errorf("failed to copy path %s: %w", manager.Src.FilePath, err)
	}

	if retCode != 0 {
		os.Exit(retCode)
	}

	return nil
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

func (spec *FileSpec) validateRemotePath(ctx *CLIContext, isDir bool) error {
	// if isDir is true, we check if the path is a directory
	// otherwise we allow path to be a file or a directory
	testFlag := "-r"
	if isDir {
		testFlag = "-d"
	}

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
		[]string{"test", testFlag, spec.FilePath},
	)
	if err != nil {
		return fmt.Errorf("failed to check if path %s exist %v", spec.FilePath, err)
	}

	if retCode != 0 {
		fmt.Fprintf(os.Stdout, "Remote path %s doesn't exist", spec.FilePath)
		os.Exit(retCode)
	}

	return err
}

// isDir is used to determine if we should check existence of a remote directory or a path in general
func (spec *FileSpec) Validate(ctx *CLIContext, isDir bool) error {
	if len(spec.InstanceID) == 0 {
		return spec.validateLocalPath()
	}
	return spec.validateRemotePath(ctx, isDir)
}
