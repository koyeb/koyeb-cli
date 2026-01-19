package koyeb

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	LargeFileWarnSize = 10 * 1024 * 1024       // 10MB
	MaxUploadSize     = 1 * 1024 * 1024 * 1024 // 1G
)

// FsRead reads a file from the sandbox
func (h *SandboxHandler) FsRead(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	path := args[1]

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	content, err := client.ReadFile(ctx.Context, path)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while reading file from sandbox",
			Why:        fmt.Sprintf("failed to read file: %s", path),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the file exists and is readable",
		}
	}

	// Write binary-safe content to stdout
	os.Stdout.Write(content)
	return nil
}

// FsWrite writes content to a file in the sandbox
func (h *SandboxHandler) FsWrite(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	path := args[1]

	var content []byte

	// Check if content is provided via file flag
	localFile, err := cmd.Flags().GetString("file")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --file flag",
			Orig:     err,
			Solution: "Check the flag syntax",
		}
	}

	if localFile != "" {
		data, err := os.ReadFile(localFile)
		if err != nil {
			return &errors.CLIError{
				What:       "Error while reading local file",
				Why:        fmt.Sprintf("failed to read file: %s", localFile),
				Additional: nil,
				Orig:       err,
				Solution:   "Check that the local file exists and is readable",
			}
		}
		content = data
	} else if len(args) >= 3 {
		content = []byte(args[2])
	} else {
		return &errors.CLIError{
			What:       "Error while writing file to sandbox",
			Why:        "no content provided",
			Additional: nil,
			Solution:   "Provide content as an argument or use -f flag to read from a local file",
		}
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	err = client.WriteFile(ctx.Context, path, content)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while writing file to sandbox",
			Why:        fmt.Sprintf("failed to write file: %s", path),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the path is valid and the directory exists",
		}
	}

	log.Infof("File written: %s (%d bytes)", path, len(content))
	return nil
}

// FsLs lists directory contents in the sandbox
func (h *SandboxHandler) FsLs(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	path := "."
	if len(args) >= 2 {
		path = args[1]
	}

	longFormat, err := cmd.Flags().GetBool("long")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --long flag",
			Orig:     err,
			Solution: "Check the flag syntax",
		}
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	entries, err := client.ListDir(ctx.Context, path)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while listing directory in sandbox",
			Why:        fmt.Sprintf("failed to list directory: %s", path),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the directory exists",
		}
	}

	if longFormat {
		// Long format with details
		for _, entry := range entries {
			typeChar := "-"
			if entry.IsDir {
				typeChar = "d"
			}

			mode := entry.Mode
			if mode == "" {
				mode = "------"
			}

			size := fmt.Sprintf("%8d", entry.Size)
			if entry.IsDir {
				size = "       -"
			}

			modTime := entry.ModTime
			if modTime == "" {
				modTime = "-"
			}

			fmt.Printf("%s%s %s %s %s\n", typeChar, mode, size, modTime, entry.Name)
		}
	} else {
		// Simple format
		for _, entry := range entries {
			name := entry.Name
			if entry.IsDir {
				name = name + "/"
			}
			fmt.Println(name)
		}
	}

	return nil
}

// FsMkdir creates a directory in the sandbox
func (h *SandboxHandler) FsMkdir(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	path := args[1]

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	err = client.MakeDir(ctx.Context, path)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while creating directory in sandbox",
			Why:        fmt.Sprintf("failed to create directory: %s", path),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the path is valid",
		}
	}

	log.Infof("Directory created: %s", path)
	return nil
}

// FsRm removes a file or directory from the sandbox
func (h *SandboxHandler) FsRm(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	path := args[1]

	// Prevent removing root
	cleanPath := filepath.Clean(path)
	if cleanPath == "/" || cleanPath == "" {
		return &errors.CLIError{
			What:     "Refusing to remove path",
			Why:      fmt.Sprintf("'%s' is a protected path", path),
			Solution: "Specify a more specific path to remove",
		}
	}

	recursive, err := cmd.Flags().GetBool("recursive")
	if err != nil {
		return &errors.CLIError{
			What:     "Error parsing flags",
			Why:      "failed to parse --recursive flag",
			Orig:     err,
			Solution: "Check the flag syntax",
		}
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	if recursive {
		err = client.DeleteDir(ctx.Context, path)
	} else {
		err = client.DeleteFile(ctx.Context, path)
	}

	if err != nil {
		return &errors.CLIError{
			What:       "Error while removing file/directory from sandbox",
			Why:        fmt.Sprintf("failed to remove: %s", path),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the path exists. Use -r flag for directories.",
		}
	}

	log.Infof("Removed: %s", path)
	return nil
}

// FsUpload uploads a local file or directory to the sandbox
func (h *SandboxHandler) FsUpload(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	localPath := args[1]
	remotePath := args[2]

	recursive, _ := cmd.Flags().GetBool("recursive")
	force, _ := cmd.Flags().GetBool("force")

	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while uploading to sandbox",
			Why:        fmt.Sprintf("failed to access local path: %s", localPath),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the local path exists and is readable",
		}
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}
	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	if fileInfo.IsDir() {
		if !recursive {
			return &errors.CLIError{
				What:     "Error while uploading to sandbox",
				Why:      fmt.Sprintf("'%s' is a directory", localPath),
				Solution: "Use -r/--recursive flag to upload directories",
			}
		}
		return h.uploadDirectory(ctx.Context, client, localPath, remotePath, force)
	}

	return h.uploadFile(ctx.Context, client, localPath, remotePath, fileInfo)
}

// uploadFile uploads a single file to the sandbox
func (h *SandboxHandler) uploadFile(ctx context.Context, client *SandboxClient, localPath, remotePath string, fileInfo fs.FileInfo) error {
	if fileInfo.Size() > MaxUploadSize {
		return &errors.CLIError{
			What:     "Error while uploading file to sandbox",
			Why:      fmt.Sprintf("file size %d exceeds 1G limit", fileInfo.Size()),
			Solution: "Reduce the file size to 1G or less",
		}
	}

	if fileInfo.Size() > LargeFileWarnSize {
		log.Warnf("Uploading large file (%d bytes) - this may take a while", fileInfo.Size())
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while uploading file to sandbox",
			Why:        fmt.Sprintf("failed to read local file: %s", localPath),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the local file exists and is readable",
		}
	}

	err = client.WriteFile(ctx, remotePath, data)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while uploading file to sandbox",
			Why:        fmt.Sprintf("failed to write remote file: %s", remotePath),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the remote path is valid",
		}
	}

	log.Infof("Uploaded %s to %s (%d bytes)", localPath, remotePath, len(data))
	return nil
}

// uploadDirectory uploads a directory recursively to the sandbox
func (h *SandboxHandler) uploadDirectory(ctx context.Context, client *SandboxClient, localPath, remotePath string, force bool) error {
	_, err := client.StatFile(ctx, remotePath)
	if err == nil {
		// Remote path exists
		if !force {
			return &errors.CLIError{
				What:     "Error while uploading directory to sandbox",
				Why:      fmt.Sprintf("remote path '%s' already exists", remotePath),
				Solution: "Use -f/--force flag to overwrite existing directory",
			}
		}
		log.Warnf("Remote path '%s' exists, overwriting...", remotePath)
	}

	var fileCount, dirCount int

	err = filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", path, err)
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		var targetPath string
		if relPath == "." {
			targetPath = remotePath
		} else {
			targetPath = remotePath + "/" + filepath.ToSlash(relPath)
		}

		if d.Type()&os.ModeSymlink != 0 {
			log.Warnf("Skipping symlink: %s", path)
			return nil
		}

		if d.IsDir() {
			if err := client.MakeDir(ctx, targetPath); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			dirCount++
			log.Debugf("Created directory: %s", targetPath)
		} else {
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info for %s: %w", path, err)
			}

			if info.Size() > MaxUploadSize {
				return &errors.CLIError{
					What:     "Error while uploading directory to sandbox",
					Why:      fmt.Sprintf("file %s exceeds 1G limit (%d bytes)", path, info.Size()),
					Solution: "Reduce the file size to 1G or less",
				}
			}

			if info.Size() > LargeFileWarnSize {
				log.Warnf("Uploading large file %s (%d bytes)", path, info.Size())
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			if err := client.WriteFile(ctx, targetPath, data); err != nil {
				return fmt.Errorf("failed to upload file %s: %w", targetPath, err)
			}
			fileCount++
			log.Debugf("Uploaded file: %s -> %s", path, targetPath)
		}

		return nil
	})

	if err != nil {
		if cliErr, ok := err.(*errors.CLIError); ok {
			return cliErr
		}
		log.Infof("Uploaded %d files and %d directories to %s before failure", fileCount, dirCount, remotePath)
		return &errors.CLIError{
			What:       "Error while uploading directory to sandbox",
			Why:        "failed during directory traversal",
			Additional: nil,
			Orig:       err,
			Solution:   "Check that all files are readable and remote paths are valid",
		}
	}

	log.Infof("Uploaded %d files and %d directories to %s", fileCount, dirCount, remotePath)
	return nil
}

// FsDownload downloads a file from the sandbox
func (h *SandboxHandler) FsDownload(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	sandboxName := args[0]
	remotePath := args[1]
	localPath := args[2]

	// Check if local path already exists
	if _, err := os.Stat(localPath); err == nil {
		// File exists - check if it's a directory
		fileInfo, _ := os.Stat(localPath)
		if fileInfo.IsDir() {
			// If it's a directory, append the remote filename
			localPath = filepath.Join(localPath, filepath.Base(remotePath))
		} else {
			log.Warnf("Overwriting existing file: %s", localPath)
		}
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(localPath)
	if parentDir != "." && parentDir != "/" {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return &errors.CLIError{
				What:       "Error while downloading file from sandbox",
				Why:        fmt.Sprintf("failed to create local directory: %s", parentDir),
				Additional: nil,
				Orig:       err,
				Solution:   "Check that you have write permissions to the parent directory",
			}
		}
	}

	info, err := h.GetSandboxInfo(ctx, sandboxName)
	if err != nil {
		return err
	}

	client := NewSandboxClient(info.Domain, info.SandboxSecret)

	content, err := client.ReadFile(ctx.Context, remotePath)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while downloading file from sandbox",
			Why:        fmt.Sprintf("failed to read remote file: %s", remotePath),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the remote file exists",
		}
	}

	err = os.WriteFile(localPath, content, 0644)
	if err != nil {
		return &errors.CLIError{
			What:       "Error while downloading file from sandbox",
			Why:        fmt.Sprintf("failed to write local file: %s", localPath),
			Additional: nil,
			Orig:       err,
			Solution:   "Check that the local path is writable",
		}
	}

	log.Infof("Downloaded %s to %s (%d bytes)", remotePath, localPath, len(content))
	return nil
}
