package configs

import (
	"fmt"
	"os"
	"strconv"
)

func RateLimit(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > 100000 {
		return 0, fmt.Errorf("%s must be between 1 and 100000", name)
	}
	return limit, nil
}
