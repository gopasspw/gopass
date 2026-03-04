package create

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

func fmtfn(d int, n string, t string) string {
	strlen := 40 - d
	// indent - [N] - text (trailing spaces)
	fmtStr := "%" + strconv.Itoa(d) + "s%s %-" + strconv.Itoa(strlen) + "s"
	debug.Log("d: %d, n: %q, t: %q, strlen: %d, fmtStr: %q", d, n, t, strlen, fmtStr)

	return fmt.Sprintf(fmtStr, "", color.GreenString("["+n+"]"), t)
}

// extractHostname tries to extract the hostname from a URL in a filepath-safe
// way for use in the name of a secret.
func extractHostname(in string) string {
	if in == "" {
		return ""
	}
	// help url.Parse by adding a scheme if one is missing. This should still
	// allow for any scheme, but by default we assume http (only for parsing)
	urlStr := in
	if !strings.Contains(urlStr, "://") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(urlStr)
	if err == nil {
		if ch := fsutil.CleanFilename(u.Hostname()); ch != "" {
			return ch
		}
	}

	return fsutil.CleanFilename(in)
}
