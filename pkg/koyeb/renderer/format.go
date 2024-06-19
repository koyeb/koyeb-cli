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
	GetSize() int64
}

type ByteSize int64

func (b ByteSize) GetSize() int64 {
	return int64(b)
}

type KBSize int64

func (k KBSize) GetSize() int64 {
	return int64(k) * 1024
}

type MBSize int64

func (m MBSize) GetSize() int64 {
	return int64(m) * 1024 * 1024
}

type GBSize int64

func (g GBSize) GetSize() int64 {
	return int64(g) * 1024 * 1024 * 1024
}

func FormatSize(sized Size) string {
	size := sized.GetSize()

	switch {
	case size > 1024*1024*1024:
		return fmt.Sprintf("%.2fG", float64(size)/1024/1024/1024)
	case size > 1024*1024:
		return fmt.Sprintf("%.2fM", float64(size)/1024/1024)
	case size > 1024:
		return fmt.Sprintf("%.2fK", float64(size)/1024)
	default:
		return fmt.Sprintf("%d", size)
	}
}
