package internal

import (
	"bytes"
	"encoding/json"
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
	if !IsFileExists(filePath) {
		return
	}
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

func loadJSONFile(jsonPath string, val any) error {
	bs, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bs), "\n")
	var bf bytes.Buffer
	for _, line := range lines {
		lineNew := strings.TrimSpace(line)
		if (len(lineNew) > 0 && lineNew[0] == '#') || (len(lineNew) > 1 && lineNew[0:2] == "//") {
			continue
		}
		bf.WriteString(lineNew)
	}
	return json.Unmarshal(bf.Bytes(), &val)
}
