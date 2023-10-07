package internal

import (
	"context"
	"log"
	"os"
	"os/exec"
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

func checkDir(fp string) error {
	info, err := os.Stat(fp)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		_ = os.RemoveAll(fp)
	}
	return os.MkdirAll(fp, 0750)
}

// LoadParamValuesFromFile load values file as  slice
func LoadParamValuesFromFile(ctx context.Context, file string) (values []string) {
	var content []byte
	var err error
	if strings.HasSuffix(file, ".sh") {
		cmd := exec.CommandContext(ctx, file)
		content, err = cmd.Output()
	} else {
		content, err = os.ReadFile(file)
	}
	if err != nil {
		log.Printf("LoadParamValuesFromFile(%q) failed, error=%v\n", file, err)
		return []string{"Error:" + err.Error()}
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		values = append(values, line)
	}
	return values
}
