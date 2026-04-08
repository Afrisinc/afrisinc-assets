package fileutil

import (
	"crypto/rand"
	"encoding/base32"
	"regexp"
	"strings"
	"time"
)

var encoding = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(base32.NoPadding)

// NewID generates a time-sortable, URL-safe unique identifier.
// Format: 10 timestamp chars + 16 random chars = 26 chars total.
func NewID() string {
	now := time.Now().UnixNano()
	b := make([]byte, 10)
	for i := 9; i >= 0; i-- {
		b[i] = byte(now & 0xff)
		now >>= 8
	}
	rnd := make([]byte, 10)
	_, _ = rand.Read(rnd)
	return encoding.EncodeToString(b) + encoding.EncodeToString(rnd)
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a filename or title into a URL-safe slug.
// e.g. "Hero Banner (v2).PNG" → "hero-banner-v2"
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "file"
	}
	return s
}

// SafeExtension returns a normalised file extension for the given MIME type,
// falling back to the original extension if the MIME type is unrecognised.
func SafeExtension(mimeType, original string) string {
	known := map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/webp":      ".webp",
		"image/gif":       ".gif",
		"image/svg+xml":   ".svg",
		"video/mp4":       ".mp4",
		"video/webm":      ".webm",
		"application/pdf": ".pdf",
		"font/ttf":        ".ttf",
		"font/otf":        ".otf",
		"font/woff":       ".woff",
		"font/woff2":      ".woff2",
	}
	if ext, ok := known[mimeType]; ok {
		return ext
	}
	// Sanitise the original extension: only allow [a-z0-9]
	ext := strings.ToLower(original)
	if !regexp.MustCompile(`^\.[a-z0-9]{1,10}$`).MatchString(ext) {
		return ".bin"
	}
	return ext
}

// HumanSize converts bytes to a human-readable string.
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return "%d B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB"}
	return strings.TrimRight(
		strings.TrimRight(
			strings.Replace(
				strings.Replace(
					func(f float64, s string) string {
						return strings.TrimRight(strings.TrimRight(
							formatF(f), "0"), ".") + " " + s
					}(float64(bytes)/float64(div), suffixes[exp]),
					".", ".", 1), "", "", 1),
		"0"), ".")
}

func formatF(f float64) string {
	s := strings.TrimRight(strings.TrimRight(
		// simple 2-decimal format without fmt import
		func() string {
			i := int64(f * 100)
			whole := i / 100
			frac := i % 100
			if frac < 0 {
				frac = -frac
			}
			return strings.Join([]string{
				itoa(whole), ".", pad2(frac),
			}, "")
		}(), "0"), ".")
	return s
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func pad2(n int64) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}
