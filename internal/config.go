package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsgo/fsconf/confext"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fsgo/fsconf"
)

func init() {
	confext.Init()
}

type Config struct {
	Listen   string // 服务端口，可选，默认 ":8310"
	Title    string
	Intro    string
	Timeout  int      // 超时时间，单位秒，默认 30s
	Charset  string   // 默认字符集，可选
	Charsets []string // 可选字符集，可选
	Commands map[string]*cmdConfig
	LogPath  string
	CacheDir string
	TmpDir   string            // 运行临时数据目录，可选，默认是和配置文件平行的 tmp 目录
	Users    map[string]string // 账号和密码，如 user:psw,可选

	confDir string // 配置文件所在目录
}

func (cfg *Config) getListen() string {
	if cfg.Listen != "" {
		return cfg.Listen
	}
	return ":8310"
}

func (cfg *Config) getTimeout() int {
	if cfg.Timeout > 0 {
		return cfg.Timeout
	}
	return 30
}

func (cfg *Config) getCacheDir() string {
	if cfg.CacheDir == "" {
		return ""
	}
	if filepath.IsAbs(cfg.CacheDir) {
		return cfg.CacheDir
	}
	return filepath.Join(cfg.confDir, cfg.CacheDir)
}

func (cfg *Config) getLogPath() string {
	if cfg.LogPath == "" {
		return ""
	}
	if filepath.IsAbs(cfg.LogPath) {
		return cfg.LogPath
	}
	return filepath.Join(cfg.confDir, cfg.LogPath)
}

func (cfg *Config) getCharsets() []string {
	if len(cfg.Charsets) > 0 {
		return cfg.Charsets
	}
	return []string{"utf-8", "gbk", "gb2312"}
}

func (cfg *Config) getCharset() string {
	if cfg.Charset != "" {
		return cfg.Charset
	}
	return "utf-8"
}

func (cfg *Config) user(name string) (psw string, found bool) {
	for userName, userPsw := range cfg.Users {
		if userName == name {
			return userPsw, true
		}
	}
	return "", false
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
	Command  string               // 完整的命令，如  ../cmds/ls.sh a $a b $b $c $d|你好
	Intro    string               // 介绍，可选
	Timeout  int                  // 参数时间，单位秒，可选
	Charset  string               // 默认输出字符编码，可选
	Charsets []string             // 可选字符集，可选
	Params   map[string]*cmdParam // 参数配置，可选
	Group    string               // 分组名称，可选
	Cache    int64                // 缓存有效期，可选,单位秒

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

func (cc *cmdConfig) getCacheLife() time.Duration {
	if cc.cfg.CacheDir == "" || cc.Cache < 1 {
		return 0
	}
	return time.Duration(cc.Cache) * time.Second
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
	return cc.cfg.getCharsets()
}

func (cc *cmdConfig) getCharset() string {
	if cc.Charset != "" {
		return cc.Charset
	}
	return cc.cfg.getCharset()
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
		param := new(cmdParam)
		param.name = item

		if item[0] == '$' {
			tmp := strings.Split(item+"|", "|")
			name := tmp[0][1:]
			if _itemConf, has := cc.Params[name]; has {
				param = _itemConf
			}
			param.isValParam = true
			param.name = name
			if param.Default == "" {
				param.Default = tmp[1]
			}
		}
		if param.Values == nil {
			param.Values = make([]string, 0)
		}

		if param.Regexp != "" {
			param.reg = regexp.MustCompile(param.Regexp)
		}

		cc.paramsAll = append(cc.paramsAll, param)
	}
	log.Println("register [", cc.name, "], command:", cc.Command)
}

func (cc *cmdConfig) String() string {
	d, _ := json.MarshalIndent(cc, "", "  ")
	return string(d)
}

type cmdParam struct {
	Default    string // 默认值
	Values     []string
	ValuesFile string // 可选
	HTML       string // HTML 属性，可选
	Regexp     string // 参数的正则表达式，可选

	name       string // 参数名，由配置内容解析得到
	isValParam bool
	reg        *regexp.Regexp
}

func (p *cmdParam) ToString() string {
	return fmt.Sprintf("name:%s,default:%s,isValParam:%v", p.name, p.Default, p.isValParam)
}

func (p *cmdParam) getValues(ctx context.Context, cmd *cmdConfig) []string {
	if p.ValuesFile == "" {
		return p.Values
	}
	return LoadParamValuesFromFile(ctx, p.ValuesFile)
}

var cmdFileNameReg = regexp.MustCompile(`^[A-Za-z0-9_]+\.(json|toml|yml)$`)

func loadConfig(confPath string) *Config {
	var cfg *Config
	if err := fsconf.Parse(confPath, &cfg); err != nil {
		log.Fatalln("load config failed:", err)
	}
	pathAbs, err := filepath.Abs(confPath)
	if err != nil {
		log.Fatalf("filepath.Abs(%q): %v\n", confPath, err)
	}
	confDir := filepath.Dir(pathAbs)
	rootDir := filepath.Dir(confDir)

	if err = os.Chdir(rootDir); err != nil {
		log.Fatalf("chdir(%q): %v\n", rootDir, err)
	}

	cfg.confDir = confDir

	if cfg.TmpDir == "" {
		cfg.TmpDir = filepath.Join(rootDir, "tmp")
	}

	if err = checkDir(cfg.TmpDir); err != nil {
		log.Fatalf("check tmpDir %q failed: %v\n", cfg.TmpDir, err)
	}

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
		cmd.confDir = filepath.Dir(cmdFilePath)
		ext := filepath.Ext(fileName)
		cmdName := fileName[:len(fileName)-len(ext)]
		cmd.name = cmdName
		log.Println("load cmd [", cmdName, "] from [", fileName, "],success")
		cfg.Commands[cmdName] = cmd
	}

	for _, item := range cfg.Commands {
		item.cfg = cfg
		if item.confDir == "" {
			item.confDir = cfg.confDir
		}
	}

	cfg.parse()
	return cfg
}
