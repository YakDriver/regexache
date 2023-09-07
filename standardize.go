package regexache

import (
	"regexp"
	"strings"
)

var (
	numFront *regexp.Regexp
	capFront *regexp.Regexp
	lowFront *regexp.Regexp
	undFront *regexp.Regexp
	word     *regexp.Regexp
)

func init() {
	undFront = regexp.MustCompile(`(\[)([^\]]*)(_)([^\]]*)(\])`)
	lowFront = regexp.MustCompile(`(\[)([^\]]*)(a-[b-z])([^\]]*)(\])`)
	capFront = regexp.MustCompile(`(\[)([^\]]*)(A-[B-Z])([^\]]*)(\])`)
	numFront = regexp.MustCompile(`(\[)([^\]]*)(0-9)([^\]]*)(\])`)
	word = regexp.MustCompile(`(\[)([^\]]*)(0-9A-Za-z_)([^\]]*)(\])`)
}

func standardize(expr string) string {
	expr = strings.ReplaceAll(expr, `\_`, "_")
	expr = undFront.ReplaceAllString(expr, "$1$3$2$4$5")
	expr = lowFront.ReplaceAllString(expr, "$1$3$2$4$5")
	expr = capFront.ReplaceAllString(expr, "$1$3$2$4$5")
	expr = numFront.ReplaceAllString(expr, "$1$3$2$4$5")
	expr = word.ReplaceAllString(expr, `$1\w$2$4$5`)
	return expr
}
