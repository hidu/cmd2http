package serve

import (
	"bytes"
	"encoding/json"
	"github.com/hidu/goutils"
	"io/ioutil"
	"os"
	"strings"
)

func GetVersion() string {
	return strings.TrimSpace(string(LoadRes("res/version")))
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func LoadRes(path string) []byte {
	return []byte(Assest.GetContent(path))
}

func In_array(item string, arr []string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

func GetCacheKey(cmd string, params []string) string {
	return cmd + "|||" + strings.Join(params, "&")
}

func LoadParamValuesFromFile(file_path string) (values []string) {
	if !IsFileExists(file_path) {
		return
	}
	bf, err := utils.File_get_contents(file_path)
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

func loadJsonFile(jsonPath string, val interface{}) error {
	bs, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(bs), "\n")
	var bf bytes.Buffer
	for _, line := range lines {
		lineNew := strings.TrimSpace(line)
		if lineNew[0] == '#' || (len(lineNew) > 1 && lineNew[0:2] == "//") {
			continue
		}
		bf.WriteString(lineNew)
	}
	return json.Unmarshal(bf.Bytes(), &val)
}
