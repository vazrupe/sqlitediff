package sqlitediff

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

func hash1(items ...interface{}) string {
	h := md5.New()
	for _, item := range items {
		switch item := item.(type) {
		case int64:
			io.WriteString(h, fmt.Sprintf("%d", item))
		case float64:
			io.WriteString(h, fmt.Sprintf("%f", item))
		case string:
			io.WriteString(h, item)
		case time.Time:
			io.WriteString(h, item.String())
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
