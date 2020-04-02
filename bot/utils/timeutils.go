package utils

import (
	"fmt"
	"strconv"
)

func FormatTime(millis int64) string {
	minutes := (millis / (1000 * 60)) % 60
	hours := (millis / (1000 * 60 * 60)) % 24

	minutesStr := strconv.Itoa(int(minutes))
	if len(minutesStr) == 1 {
		minutesStr = "0" + minutesStr
	}

	return fmt.Sprintf("%d:%s", hours, minutesStr)
}
