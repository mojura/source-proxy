package proxy

import (
	"strconv"
	"time"
)

func updateFilename(filename string) string {
	unix := time.Now().UnixNano()
	unixStr := strconv.FormatInt(unix, 10)
	return p.match.ReplaceAllString(filename, unixStr)
}
