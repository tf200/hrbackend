package bucket

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	filenameBaseSanitizer = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)
	underscoreCollapse    = regexp.MustCompile(`_+`)
	filenameExtSanitizer  = regexp.MustCompile(`[^a-zA-Z0-9.]+`)
)

func ValidateFile(header *multipart.FileHeader, maxFileSize int64) error {
	if header.Size == 0 {
		return fmt.Errorf("file cannot be empty")
	}

	if header.Size > maxFileSize {
		return fmt.Errorf("file size exceeds maximum limit of 10MB")
	}

	return nil
}

func GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	nameWithoutExt := strings.TrimSuffix(originalFilename, ext)
	timestamp := time.Now().Format("20060102150405")
	id := uuid.New().String()[:8]

	return fmt.Sprintf("%s_%s_%s%s",
		sanitizeFilename(nameWithoutExt, 0),
		timestamp,
		id,
		ext,
	)
}

func sanitizeFilename(filename string, maxLen int) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	base = filenameBaseSanitizer.ReplaceAllString(base, "_")
	base = underscoreCollapse.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._- _")
	if base == "" {
		base = "file"
	}

	ext = filenameExtSanitizer.ReplaceAllString(ext, "")
	if ext == "." {
		ext = ""
	}

	out := base + ext
	if maxLen > 0 && len(out) > maxLen {
		keepExt := ext
		if len(keepExt) >= maxLen {
			keepExt = ""
		}

		maxBase := maxLen - len(keepExt)
		if maxBase < 1 {
			maxBase = maxLen
			keepExt = ""
		}

		baseTrunc := base
		if len(baseTrunc) > maxBase {
			baseTrunc = baseTrunc[:maxBase]
		}

		out = baseTrunc + keepExt
	}

	return out
}
