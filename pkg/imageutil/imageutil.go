package imageutil

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"

	_ "golang.org/x/image/webp"
)

// IsRasterImage reports whether the MIME type is a decodable raster format.
func IsRasterImage(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

// Dimensions decodes the image header from r and returns width × height.
// It does NOT read the full image body — image.DecodeConfig is fast.
func Dimensions(r io.Reader) (width, height int, err error) {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, fmt.Errorf("imageutil: decode config: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

// IsSVG is a quick heuristic check (MIME type already validated upstream).
func IsSVG(mimeType string) bool {
	return strings.Contains(mimeType, "svg")
}
