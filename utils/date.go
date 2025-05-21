package utils


import (
	"time"
)

const dateLayout = "02.01.2006"

// parseDate парсит строку с датой в формате DD.MM.YYYY и возвращает time.Time
// Если строка пустая или nil, возвращает nil
func ParseDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil || *dateStr == "" {
		return nil, nil
	}

	parsedDate, err := time.Parse(dateLayout, *dateStr)
	if err != nil {
		return nil, err
	}

	return &parsedDate, nil
}

// parseRequiredDate парсит обязательную строку с датой в формате DD.MM.YYYY
func ParseRequiredDate(dateStr string) (time.Time, error) {
	return time.Parse(dateLayout, dateStr)
}