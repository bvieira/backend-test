package jobs

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// removeAccents remove accents from a string
func removeAccents(value string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, value)
	return result
}

func hash(value string) string {
	h := sha1.New()
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

// createID create an sha1 hash id by converting to lowercase, removing accents and replacing all non alphanumeric with '-'
func createID(values ...string) string {
	var buffer bytes.Buffer
	for i, v := range values {
		buffer.WriteString(removeAccents(v))
		if i < len(values)-1 {
			buffer.WriteString(" ")
		}
	}
	r := regexp.MustCompile("[[:^alnum:]]+")
	return hash(r.ReplaceAllString(strings.ToLower(buffer.String()), "-"))
}
