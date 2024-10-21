package env

import (
	"fmt"
	"os"
	"time"
)

func GetEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%s not set", key)
	}
	return value, nil
}

func GetEnvDuration(key string) (time.Duration, error) {
	value, err := GetEnv(key)
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return duration, nil
}

func GetSecret(secretName string) ([]byte, error) {
	secret, err := os.ReadFile(fmt.Sprintf("/run/secrets/%s", secretName))
	if err != nil {
		return nil, fmt.Errorf("load secret %s: %w", secretName, err)
	}
	return secret, nil
}
