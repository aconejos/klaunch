package utils

import (
	"fmt"
	"time"
)

// GenerateLogFilename creates a timestamped log filename
func GenerateLogFilename(prefix string) string {
	datePrefix := time.Now().Format("20060102_150405")
	return fmt.Sprintf("logs/%s_%s.log", datePrefix, prefix)
}

// FormatError formats error messages consistently
func FormatError(operation string, err error) string {
	return fmt.Sprintf("Error %s: %v", operation, err)
}
