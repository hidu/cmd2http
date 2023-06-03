package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type serverConf struct {
	Port        int                 `json:"port"`
	Title       string              `json:"title"`
	Intro       string              `json:"intro"`
	Timeout     int                 `json:"timeout"`
	CharsetList []string            `json:"charset_list"`
	Cmds        map[string]*cmdItem `json:"cmds"`
	LogPath     string              `json:"log_path"`
	CacheDir    string              `json:"cache_dir"`
	confPath    string
}

type cmdItem struct {
	Name      string `json:"-"`
	CmdRaw    string `json:"cmd"`
	Cmd       string `json:"-"`
	Charset   string `json:"charset"`
	paramsAll []*cmdParam
	Params    map[string]*cmdParam `json:"params"`
	Intro     string               `json:"intro"`
	Timeout   int                  `json:"timeout"`
	Charsets  []string             `json:"charset_list"`
	Group     string               `json:"group"`
	CacheLife int64                `json:"cache_life"`
}

func (c *cmdItem) getTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

func (c *cmdItem) String() string {
	d, _ := json.MarshalIndent(c, "", "  ")
	return string(d)
}

func (conf *serverConf) String() string {
	d, _ := json.MarshalIndent(conf, "", "  ")
	return string(d)
}

type cmdParam struct {
	Name         string `json:"-"`
	DefaultValue string `json:"default_value"`
	isValParam   bool
	Values       []string `json:"values"`
	HTML         string   `json:"html"`
	ValuesFile   string   `json:"values_file"`
}

func (p *cmdParam) ToString() string {
	return fmt.Sprintf("name:%s,default:%s,isValParam:%v", p.Name, p.DefaultValue, p.isValParam)
}

func (p *cmdParam) getValues() []string {
	if p.ValuesFile == "" {
		return p.Values
	}
	return LoadParamValuesFromFile(p.ValuesFile)
}

func (conf *serverConf) parse() {
	if conf.Port <= 0 {
		conf.Port = 8310
	}
	if conf.Timeout < 1 {
		conf.Timeout = 30
	}
	for cmdName, cmdConf := range conf.Cmds {
		if cmdConf.Group == "" {
			cmdConf.Group = "default"
		}

		if cmdConf.Timeout < 1 {
			cmdConf.Timeout = conf.Timeout
		}
		if cmdConf.Charsets == nil {
			cmdConf.Charsets = make([]string, 0)
		}
		if cmdConf.Params == nil {
			cmdConf.Params = make(map[string]*cmdParam)
		}
		cmdConf.CmdRaw = strings.TrimSpace(cmdConf.CmdRaw)

		// cmd eg  echo -n $wd|你好 $a $b
		ps := regexp.MustCompile(`\s+`).Split(cmdConf.CmdRaw, -1)
		//       fmt.Println(ps)
		cmdConf.Cmd = ps[0]
		// @todo
		for i := 1; i < len(ps); i++ {
			item := ps[i]
			//           fmt.Println("i:",i,item)
			_param := new(cmdParam)
			_param.Name = item

			if item[0] == '$' {
				tmp := strings.Split(item+"|", "|")
				name := tmp[0][1:]
				if _itemConf, has := cmdConf.Params[name]; has {
					_param = _itemConf
				}
				_param.isValParam = true
				_param.Name = name
				if _param.DefaultValue == "" {
					_param.DefaultValue = tmp[1]
				}
			}
			if _param.Values == nil {
				_param.Values = make([]string, 0)
			}
			cmdConf.paramsAll = append(cmdConf.paramsAll, _param)
		}
		log.Println("register[", cmdName, "] cmd:", cmdConf.CmdRaw)
	}
}

func (conf *serverConf) groups() []string {
	var groups []string
	for _, cmdItem := range conf.Cmds {
		if !InArray(cmdItem.Group, groups) {
			groups = append(groups, cmdItem.Group)
		}
	}
	sort.Strings(groups)
	return groups
}

func loadConfig(confPath string) (serConf *serverConf) {
	err := loadJSONFile(confPath, &serConf)
	if err != nil {
		log.Fatalln("load config failed:", err)
	}
	pathAbs, _ := filepath.Abs(confPath)
	serConf.confPath = pathAbs
	confDir := filepath.Dir(pathAbs)
	os.Chdir(confDir)
	log.Println("chdir ", confDir)
	fileNames, err := filepath.Glob(fmt.Sprintf("%s%scmd%s*.json", confDir, string(filepath.Separator), string(filepath.Separator)))
	if err != nil {
		log.Println("scan cmd config files in sub dir [cmd] failed,skip,err:", err)
	}
	cmdFileNameReg := regexp.MustCompile(`^[A-Za-z0-9_]+\.json$`)
	for _, cmdFilePath := range fileNames {
		_, cmdFileName := filepath.Split(cmdFilePath)
		if !cmdFileNameReg.MatchString(cmdFileName) {
			log.Println(`sub cmd config file ignore [`, cmdFileName, `],not match reg:^[A-Za-z0-9_]+\.json$`)
			continue
		}
		var cmd *cmdItem
		err := loadJSONFile(cmdFilePath, &cmd)
		if err != nil {
			log.Println("load cmd from [", cmdFileName, "] failed,err:", err)
			continue
		}
		if cmd.Charsets == nil {
			cmd.Charsets = serConf.CharsetList
		}
		cmdName := cmdFileName[:len(cmdFileName)-5]
		log.Println("load cmd [", cmdName, "] from [", cmdFileName, "],success")
		serConf.Cmds[cmdName] = cmd
	}

	serConf.parse()
	return
}
