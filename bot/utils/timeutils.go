package utils

import "fmt"

func FormatTime(millis int64) string {
	minutes := (millis / (1000 * 60)) % 60
	hours := (millis / (1000 * 60 * 60)) % 24
	return fmt.Sprintf("%d:%d", hours, minutes)
}
