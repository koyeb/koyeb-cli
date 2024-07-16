package renderer

import (
	"fmt"
	"time"

	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC822)
}

func FormatAppName(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.App().GetName(id)
		if err == nil {
			return sid
		}
	}
	return id
}

func FormatServiceSlug(mapper *idmapper.Mapper, id string, full bool) string {
	if !full {
		sid, err := mapper.Service().GetSlug(id)
		if err == nil {
			return sid
		}
	}
	return id
}

// FormatID formats the ID to be displayed in the CLI. If full is false, only the first 8 characters are displayed.
func FormatID(fullId string, full bool) string {
	if full {
		return fullId
	}
	if fullId == "" {
		return ""
	}
	return fullId[:8]
}

type Size interface {
	GetSize() (int64, int64, string)
}

type ByteSize int64

func (b ByteSize) GetSize() (int64, int64, string) {
	return int64(b), 1, "B"
}

type KiBSize int64

func (k KiBSize) GetSize() (int64, int64, string) {
	return int64(k) * 1024, 1024, "KiB"
}

type MiBSize int64

func (m MiBSize) GetSize() (int64, int64, string) {
	return int64(m) * 1024 * 1024, 1024 * 1024, "MiB"
}

type GiBSize int64

func (g GiBSize) GetSize() (int64, int64, string) {
	return int64(g) * 1024 * 1024 * 1024, 1024 * 1024 * 1024, "GiB"
}

type KBSize int64

func (k KBSize) GetSize() (int64, int64, string) {
	return int64(k) * 1000, 1000, "KB"
}

type MBSize int64

func (m MBSize) GetSize() (int64, int64, string) {
	return int64(m) * 1000 * 1000, 1000 * 1000, "MB"
}

type GBSize int64

func (g GBSize) GetSize() (int64, int64, string) {
	return int64(g) * 1000 * 1000 * 1000, 1000 * 1000 * 1000, "GB"
}

func FormatSize(sized Size) string {
	size, mul, unit := sized.GetSize()

	return fmt.Sprintf("%.2f%s", float64(size)/float64(mul), unit)
}
