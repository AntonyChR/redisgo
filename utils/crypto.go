package utils 

import (
	b64 "encoding/base64"
	"strings"

	uuid "github.com/google/uuid"
)

func GenerateUUID() string {
	uuid := uuid.New().String()
	cleanUUID := strings.ReplaceAll(uuid, "-", "")
	return cleanUUID
}

func decodeBase64(s string) []byte {
	dec, _ := b64.StdEncoding.DecodeString(s)
	return dec
}
