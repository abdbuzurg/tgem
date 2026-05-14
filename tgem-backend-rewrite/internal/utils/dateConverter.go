package utils

import (
	"fmt"
	"time"
)

func DateConverter(date time.Time) string {
	day := date.Day()
	month := date.Month()
	year := date.Year()

	monthStr := ""
	switch month {
	case 1:
		monthStr = "января"
	case 2:
		monthStr = "февраля"
	case 3:
		monthStr = "марта"
	case 4:
		monthStr = "апреля"
	case 5:
		monthStr = "мая"
	case 6:
		monthStr = "июня"
	case 7:
		monthStr = "июля"
	case 8:
		monthStr = "августа"
	case 9:
		monthStr = "сентября"
	case 10:
		monthStr = "октября"
	case 11:
		monthStr = "ноября"
	case 12:
		monthStr = "декабря"
	}

  return fmt.Sprintf("%d %s %d", day, monthStr, year)
}
