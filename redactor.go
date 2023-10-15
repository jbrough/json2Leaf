package json2Leaf

import "regexp"

var (
	PersonNameRe = regexp.MustCompile(`[A-Z]([a-z]+|\.)(?:\s+[A-Z]([a-z]+|\.))*(?:\s+[a-z][a-z\-]+){0,2}\s+[A-Z]([a-z]+|\.)`)
)

// this is blunt and ineffective but intended to flag fields that have not been
// explicitly excluded.
func Redact(str string) (r string, ok bool) {
	r = str
	if ok = PersonNameRe.MatchString(str); ok {
		r = PersonNameRe.ReplaceAllString(str, "[PII redacted]")
	}

	return
}
