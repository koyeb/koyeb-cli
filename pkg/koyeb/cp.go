package koyeb

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
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

	var tarErr error
	go func(srcPath string, writer io.WriteCloser) {
		defer writer.Close()
		tarErr = Tar(srcPath, writer)
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
	if err != nil || retCode != 0 {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        fmt.Sprintf("Failed copying path %v", manager.Src.FilePath),
			Additional: []string{"This might indicate issues with connectivity to the instance"},
			Orig:       fmt.Errorf("err: %v, return code: %v, tar err: %v", err, retCode, tarErr),
			Solution:   "Make sure that your network connection is stable. If the problem persists, try to update the CLI to the latest version.",
		}
	}

	return nil
}

func (manager *CopyManager) copyFromInstance(ctx *CLIContext) error {
	reader, writer := io.Pipe()

	var untarErr error
	go func(dstPath string, reader io.ReadCloser) {
		defer reader.Close()

		untarErr = Untar(dstPath, reader)
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
	if err != nil || retCode != 0 {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        fmt.Sprintf("Failed copying path %v", manager.Src.FilePath),
			Additional: []string{"This might indicate issues with connectivity to the instance"},
			Orig:       fmt.Errorf("err: %v, return code: %v, untar err: %v", err, retCode, untarErr),
			Solution:   "Make sure that your network connection is stable. If the problem persists, try to update the CLI to the latest version.",
		}
	}

	return nil
}

func ValidateTargets(srcSpec, dstSpec *FileSpec) error {
	if srcSpec.InstanceID != "" && dstSpec.InstanceID != "" {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        "One of source or destination path must be a local path",
			Additional: []string{"This might indicate that you passed two remote paths"},
			Orig:       nil,
			Solution:   "If the problem persists, try to update the CLI to the latest version.",
		}
	}

	if srcSpec.InstanceID == "" && dstSpec.InstanceID == "" {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        "One of source or destination must be a remote path",
			Additional: []string{"This might indicate that you passed two local paths"},
			Orig:       nil,
			Solution:   "If the problem persists, try to update the CLI to the latest version.",
		}
	}

	if len(srcSpec.FilePath) == 0 || len(dstSpec.FilePath) == 0 {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        "Paths cannot be empty",
			Additional: []string{"This might indicate that you passed empty path"},
			Orig:       nil,
			Solution:   "If the problem persists, try to update the CLI to the latest version.",
		}
	}

	return nil
}

type FileSpec struct {
	InstanceID string
	FilePath   string
}

func (spec *FileSpec) validateLocalPath() error {
	if _, err := os.Stat(spec.FilePath); err != nil {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        fmt.Sprintf("The local path %s doesn't exist", spec.FilePath),
			Additional: nil,
			Orig:       err,
			Solution:   "Make sure that the local path exists",
		}
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
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        fmt.Sprintf("Failed to verify if remote path %v exists", spec.FilePath),
			Additional: []string{"This might indicate issues with connectivity to the instance"},
			Orig:       err,
			Solution:   "Make sure that your network connection is stable. If the problem persists, try to update the CLI to the latest version.",
		}
	}

	if retCode != 0 {
		return &errors.CLIError{
			What:       "Error while copying",
			Why:        fmt.Sprintf("The remote path %s doesn't exist", spec.FilePath),
			Additional: []string{"You might also double check if you have permissions to read this path"},
			Orig:       err,
			Solution:   "Make sure that the remote path exists",
		}
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
