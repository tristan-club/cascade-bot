package util

import (
	"github.com/google/uuid"
	"strings"
)

func GenerateUuid(format bool) string {
	uuidValue := uuid.New().String()
	if format {
		uuidValue = strings.Replace(uuidValue, "-", "", -1)
	}
	return uuidValue
}
