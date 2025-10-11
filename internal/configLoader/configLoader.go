package configLoader

import (
	"log"
	"os"
	"strings"
)

func MustReadSecret(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read secret %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}