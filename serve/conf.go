package serve

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	confPath    string              `json:"-"`
}
type cmdItem struct {
	Name        string               `json:"-"`
	CmdRaw      string               `json:"cmd"`
	Cmd         string               `json:"-"`
	Charset     string               `json:"charset"`
	paramsAll   []*cmdParam          `json:"-"`
	Params      map[string]*cmdParam `json:"params"`
	Intro       string               `json:"intro"`
	Timeout     int                  `json:"timeout"`
	Charsetlist []string             `json:"charset_list'`
	Group       string               `json:"group"`
	CacheLife   int64                `json:"cache_life"`
}

type cmdParam struct {
	Name         string
	DefaultValue string `json:"default_value"`
	isValParam   bool
	Values       []string `json:"values"`
	Html         string   `json:"html"`
	ValuesFile   string   `json:"values_file'`
}

func (p *cmdParam) ToString() string {
	return fmt.Sprintf("name:%s,default:%s,isValParam:%x", p.Name, p.DefaultValue, p.isValParam)
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
			cmdConf.Timeout = 30
		}
		cmdConf.CmdRaw = strings.TrimSpace(cmdConf.CmdRaw)

		//cmd eg  echo -n $wd|你好 $a $b
		ps := regexp.MustCompile(`\s+`).Split(cmdConf.CmdRaw, -1)
		//       fmt.Println(ps)
		cmdConf.Cmd = ps[0]
		//@todo
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
			cmdConf.paramsAll = append(cmdConf.paramsAll, _param)
		}
		log.Println("register[", cmdName, "] cmd:", cmdConf.CmdRaw)
	}
}

func (conf *serverConf) groups() []string {
	var groups []string
	for _, cmdItem := range conf.Cmds {
		if !In_array(cmdItem.Group, groups) {
			groups = append(groups, cmdItem.Group)
		}
	}
	return groups
}

func loadConfig(confPath string) (serConf *serverConf) {
	err := loadJsonFile(confPath, &serConf)
	if err != nil {
		log.Fatalln("load config failed:", err)
	}
	serConf.parse()
	pathAbs, _ := filepath.Abs(confPath)
	serConf.confPath = pathAbs
	conf_dir := filepath.Dir(pathAbs)
	os.Chdir(conf_dir)
	log.Println("chdir ", conf_dir)
	return
}
