package internal

import (
	"os"
	"strings"
)

// IsFileExists check file exists
func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// InArray lick php func in_array
func InArray(item string, arr []string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

// GetCacheKey get cache key
func GetCacheKey(cmd string, params []string) string {
	return cmd + "|||" + strings.Join(params, "&")
}

// LoadParamValuesFromFile load values file as  slice
func LoadParamValuesFromFile(filePath string) (values []string) {
	bf, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	lines := strings.Split(string(bf), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		values = append(values, line)
	}
	return
}

type cacheData struct {
	Data []byte
}
