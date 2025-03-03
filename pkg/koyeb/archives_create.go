package koyeb

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/archive"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Create makes a .tar.gz out of the given path, queries the endpoint
// /v1/archives to get a signed URL to upload the archive to, and uploads the
// archive.
func (h *ArchiveHandler) CreateArchive(ctx *CLIContext, path string) (*koyeb.CreateArchiveReply, error) {
	tarball, err := archive.Archive(path, h.ignoreDirectories)
	if err != nil {
		return nil, &errors.CLIError{
			What:       "Unable to create archive",
			Why:        fmt.Sprintf("we encountered an error while archiving the directory `%s`", path),
			Additional: nil,
			Orig:       err,
			Solution:   errors.SolutionFixRequest,
		}
	}
	defer tarball.Close()

	stat, err := tarball.File.Stat()
	if err != nil {
		return nil, &errors.CLIError{
			What:       "Unable to create archive",
			Why:        fmt.Sprintf("error while getting the file info of `%s` (archive of %s)", tarball.File.Name(), path),
			Additional: nil,
			Orig:       err,
			Solution:   errors.SolutionUpdateOrIssue,
		}
	}

	// Request the signed upload URL
	c := koyeb.NewCreateArchiveWithDefaults()
	// The cast to string is necessary because the API expects a string. This is
	// because the underlying type to store the size is uint64, which is not
	// representable in JSON.
	c.SetSize(fmt.Sprintf("%d", stat.Size()))

	res, resp, err := ctx.Client.ArchivesApi.CreateArchive(ctx.Context).Archive(*c).Execute()
	if err != nil {
		return nil, errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while requesting an upload URL to upload the archive '%s' (%d bytes)", path, stat.Size()),
			err,
			resp,
		)
	}
	// Upload the archive to the upload URL
	if err := h.uploadArchive(tarball.File, stat.Size(), *res.GetArchive().UploadUrl); err != nil {
		return nil, err
	}
	return res, nil
}

func (h *ArchiveHandler) Create(ctx *CLIContext, cmd *cobra.Command, path string) error {
	res, err := h.CreateArchive(ctx, path)
	if err != nil {
		return err
	}

	createArchiveReply := NewCreateArchiveReply(res)
	ctx.Renderer.Render(createArchiveReply)
	return nil
}

func (h *ArchiveHandler) uploadArchive(file *os.File, size int64, url string) error {
	log.Debugf("Start uploading archive %s to %s (%d bytes)", file.Name(), url, size)

	client := http.Client{}

	// Reset the file pointer to the beginning, otherwise the file will be empty
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Debugf("Unable to seek to the beginning of the file: %v", err)
	}

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		log.Debugf("Error while creating the request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-goog-content-length-range", fmt.Sprintf("%d,%d", size, size))

	resp, err := client.Do(req)
	if err != nil {
		return &errors.CLIError{
			What: "Error while uploading archive",
			Why:  fmt.Sprintf("Failed to upload the archive %s", file.Name()),
			Additional: []string{
				"An error occurred while uploading the archive",
			},
			Orig:     nil,
			Solution: "Make sure that your network connection is stable. If the problem persists, try to update the CLI to the latest version.",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readAllErr := io.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("Unable to read response body: %v", readAllErr)
		}

		if len(body) > 0 {
			log.Debugf("Response body:\n%s", body)
		}

		return &errors.CLIError{
			What: "Error while uploading archive",
			Why:  fmt.Sprintf("Failed to upload the archive %s", file.Name()),
			Additional: []string{
				fmt.Sprintf("We tried to upload the archive to our archive storage but the server returned a HTTP/%d response", resp.StatusCode),
			},
			Orig:     nil,
			Solution: errors.SolutionTryAgainOrUpdateOrIssue,
		}
	}

	log.Debugf("Finished uploading archive %s to %s", file.Name(), url)
	return nil
}

type CreateArchiveReply struct {
	value *koyeb.CreateArchiveReply
}

func NewCreateArchiveReply(value *koyeb.CreateArchiveReply) *CreateArchiveReply {
	return &CreateArchiveReply{value}
}

func (CreateArchiveReply) Title() string {
	return "Archive"
}

func (r *CreateArchiveReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *CreateArchiveReply) Headers() []string {
	return []string{"id", "size"}
}

func (r *CreateArchiveReply) Fields() []map[string]string {
	item := r.value.GetArchive()
	fields := map[string]string{
		// While most of the time we use the FormatID function to format the ID
		// to display a short version of it (unless the --full flag is used), we
		// don't do it here because we always want to display the full ID of the
		// archive.
		//
		// This is because there is currently no API to list archives, and
		// consequently no way to resolve a short ID to a full ID. As a result,
		// we need to provide the long ID to allow users to reference the
		// archive in services.
		"id":   item.GetId(),
		"size": fmt.Sprintf("%s bytes", item.GetSize()),
	}

	resp := []map[string]string{fields}
	return resp
}
