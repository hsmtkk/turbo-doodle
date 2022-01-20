package env

import (
	"log"
	"os"
	"strconv"
)

func RequiredString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("you must define %s environment variable", key)
	}
	return val
}

func RequiredInt(key string) int {
	s := RequiredString(key)
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("failed to parse %s as int; %s", s, err)
	}
	return n
}
