package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fsgo/fsconf"
)

type Config struct {
	Port     int
	Title    string
	Intro    string
	Timeout  int
	Charsets []string
	Commands map[string]*cmdConfig
	LogPath  string
	CacheDir string
	confDir  string // 配置文件所在目录
}

func (cfg *Config) getPort() int {
	if cfg.Port > 0 {
		return cfg.Port
	}
	return 8310
}

func (cfg *Config) getTimeout() int {
	if cfg.Timeout > 0 {
		return cfg.Timeout
	}
	return 30
}

func (cfg *Config) String() string {
	d, _ := json.MarshalIndent(cfg, "", "  ")
	return string(d)
}

func (cfg *Config) parse() {
	for _, cmdConf := range cfg.Commands {
		cmdConf.parser()
	}
}

func (cfg *Config) groups() []string {
	var groups []string
	for _, item := range cfg.Commands {
		name := item.getGroup()
		if !InArray(name, groups) {
			groups = append(groups, name)
		}
	}
	sort.Strings(groups)
	return groups
}

type cmdConfig struct {
	Command   string               // 完整的命令，如  ../cmds/ls.sh a $a b $b $c $d|你好
	Intro     string               // 介绍，可选
	Timeout   int                  // 参数时间，单位秒，可选
	Charset   string               // 默认输出字符编码，可选
	Charsets  []string             // 可选字符集，可选
	Params    map[string]*cmdParam // 参数配置，可选
	Group     string               // 分组名称，可选
	CacheLife int64                // 缓存有效期，可选

	// 以下参数不需要配置

	cmdName   string      // 命令的名称，不包含参数，如 ls,由 Command 解析得到
	name      string      // 配置文件的名称，不包含后缀，如 ls
	paramsAll []*cmdParam // 解析后的参数值，
	cfg       *Config     // 主配置文件，用于读取默认值
	confDir   string      // 当前配置文件所在目录
}

func (cc *cmdConfig) getTimeout() time.Duration {
	if cc.Timeout > 0 {
		return time.Duration(cc.Timeout) * time.Second
	}
	return time.Duration(cc.cfg.getTimeout()) * time.Second
}

func (cc *cmdConfig) getGroup() string {
	if cc.Group != "" {
		return cc.Group
	}
	return "default"
}

func (cc *cmdConfig) getCharsets() []string {
	if len(cc.Charsets) > 0 {
		return cc.Charsets
	}
	return cc.cfg.Charsets
}

func (cc *cmdConfig) parser() {
	if cc.Params == nil {
		cc.Params = make(map[string]*cmdParam)
	}
	cc.Command = strings.TrimSpace(cc.Command)

	// cmd eg:  echo -n $wd|你好 $a $b
	ps := strings.Fields(cc.Command)
	cc.cmdName = ps[0]
	// @todo
	for i := 1; i < len(ps); i++ {
		item := ps[i]
		_param := new(cmdParam)
		_param.Name = item

		if item[0] == '$' {
			tmp := strings.Split(item+"|", "|")
			name := tmp[0][1:]
			if _itemConf, has := cc.Params[name]; has {
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
		cc.paramsAll = append(cc.paramsAll, _param)
	}
	log.Println("register[", cc.name, "] cmd:", cc.Command)
}

func (cc *cmdConfig) String() string {
	d, _ := json.MarshalIndent(cc, "", "  ")
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

var cmdFileNameReg = regexp.MustCompile(`^[A-Za-z0-9_]+\.(json|toml|yml)$`)

func loadConfig(confPath string) *Config {
	var cfg *Config
	if err := fsconf.Parse(confPath, &cfg); err != nil {
		log.Fatalln("load config failed:", err)
	}
	pathAbs, err := filepath.Abs(confPath)
	if err != nil {
		log.Fatalf("filepath.Abs(%q): %v", confPath, err)
	}
	confDir := filepath.Dir(pathAbs)
	cfg.confDir = confDir

	fileNames, err := filepath.Glob(filepath.Join(confDir, "cmd", "*"))
	if err != nil {
		log.Println("scan cmd config files in sub dir [cmd] failed,skip,err:", err)
	}

	if cfg.Commands == nil {
		cfg.Commands = make(map[string]*cmdConfig)
	}

	for _, cmdFilePath := range fileNames {
		fileName := filepath.Base(cmdFilePath)
		if !cmdFileNameReg.MatchString(fileName) {
			log.Println(`sub cmd config file ignore [`, fileName, `],not match reg:`, cmdFileNameReg.String())
			continue
		}
		var cmd *cmdConfig
		err = fsconf.Parse(cmdFilePath, &cmd)
		if err != nil {
			log.Println("load cmd from [", fileName, "] failed,err:", err)
			continue
		}
		cmd.cfg = cfg
		cmd.confDir = filepath.Dir(cmdFilePath)
		ext := filepath.Ext(fileName)
		cmdName := fileName[:len(fileName)-len(ext)]
		cmd.name = cmdName
		log.Println("load cmd [", cmdName, "] from [", fileName, "],success")
		cfg.Commands[cmdName] = cmd
	}

	cfg.parse()
	return cfg
}
