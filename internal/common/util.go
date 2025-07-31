package common

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ToStringPointer(t *time.Time) *string {
	if t != nil {
		str := t.Format(time.RFC3339)
		return &str
	}
	return nil
}

func ConvertStringToDate(input string) time.Time {
	if input == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", input)

	if err != nil {
		return time.Time{}
	}
	return t
}

func ConvertStringToTimeOnly(input string) time.Time {
	if input == "" {
		return time.Time{}
	}
	t, err := time.Parse("15:04:05", input)
	if err != nil {
		return time.Time{}
	}
	return t
}
