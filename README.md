cmd2http
=========
将 CLI 程序(系统命令、脚本等)转换为 HTTP 服务  


## Install
```
go install github.com/hidu/cmd2http@master
```

## Run
```
./cmd2http -conf=./example/conf/cmd2http.json
```

访问首页: http://localhost:8310/

**hello world demo:**  
```
           URL : http://localhost:8310/echo?wd=hello&a=world
Command Config : <b>echo -n $wd $a $b|defaultValue </b>  
Command   Exec : <b>echo -n hello world defaultValue</b>  
```

##配置说明

###目录结构

```
├── cmds
│   ├── ls.sh
│   └── sleep.sh
├── conf
│   ├── cmd2http.json
│   ├── cmd
│   │   ├── echo.json
│   │   ├── ls.json
│   │   ├── sleep_1.json
│   │   ├── sleep.json
│   │   ├── wrong_json.json
│   │   └── 错误的文件名.json
│   └── s
│       ├── my.css
│       └── my.js
└── data
    └── ls_d.csv

```

### 主配置文件 
支持 `.json`、`.toml` 后缀的文件。

如 `conf/cmd2http.json`:
```json
{
   "Port":8310,
   "Title":"default title",
   "Intro":"intro info",
   "Timeout":30,
   "CacheDir":"../cache_data/",
   "Charset":"utf-8",
   "Charsets": [
        "gbk",
        "utf-8"
    ],
   "Commands":{
      "pwd":{
          "Command":"pwd",
          "Intro":"cmd intor",
          "Timeout":10
       },
      "echo":{
         "Command":"echo -n $wd|你好 $a $b",
         "Cache":120
        }
   }
}
```
命令配置(Commands)(如上的pwd，echo)也可以配置到单独的文件，位于上述配置文件(cmd2http.json)目录下的 `cmd` 子目录下去。  

所以在配置文件中写的路径都使用以此为基准目录的相对路径即可。 

配置项说明： 
*   Port      : HTTP 服务监听端口
*   Title     : HTTP 页面标题
*   Intro     : 介绍
*   Timeout   : 默认的运行超时时间
*   CacheDir  : 运行结果缓存存放的目录（单项命令中配置了 Cache 项后生效）
*   Charset   : 全局默认的编码，只用于 HTML 页面结果展现
*   Charsets  : 全局默认的编码可选值，只用于 HTML 页面结果展现

### 命令配置文件
支持 `.json`、`.toml` 后缀的文件，配置文件存放于 `conf/cmd/` 目录。

`ls.json` 内容为：
```json
{
    "Command": "../cmds/ls.sh a $a b $b $c $d|你好",
    "Intro": "hello",
    "Timeout": 3,
    "Cache": 1800,
    "Group": "分组1",
    "Charset":"utf-8",
   "Charsets": [
        "gbk",
        "utf-8"
    ],
    "Params": {
        "c": {
            "Values": ["1","2","3" ],
            "HTML": "style='width:200px'"
        },
        "d": {
            "ValuesFile": "../data/ls_d.csv"
        }
    }
}
```
配置项说明：  
*  Command   :  待运行的命令，参数使用`$`前缀，如 `$a`,`$a1`,`$a_1`  
*  Intro : 介绍
*  Timeout : 当前命令的运行超时时间，若没有设置或者为0 则使用全局的 timeout
*  Cache : 运行结果缓存有效期，单位秒，为 0 或者全局的 `CacheDir` 没有设置的时候不使用缓存
*  Params : 参数配置
*  Params.c : 参数 `$c`的配置项
*  Params.c.Values : 参数 `$c`的可选值，用来在 form 表单中展现，只能是字符串,values 有值的情况下使用 select 展现样式，否则为 input=text
*  Params.c.HTML : 参数 `$c`的 form 控件额外的 HTML 代码块
*  Params.c.ValuesFile : 参数 `$c`的可选值录入文件(eg:[可选值示例文件](./example/data/ls_d.csv))
*  Charset   : 默认的编码，只用于 HTML 页面结果展现
*  Charsets  : 默认的编码可选值，只用于 HTML 页面结果展现
*  Group     :  页面展现分组，默认为`default`

### 命令如何读取参数
```
"Command": "../cmds/ls.sh a $a b $b $c $d|你好"
```
如上，命令中定义了很多参数，`ls.sh`一共可以读取到6个参数，其中 字符串`a`,`b` 是固定的，`$a $b $c $d` 这几个则可以从http接口读取到。  
`$d|你好` 表示当http接口读取到的值为空时的默认值。  
通过http接口调用时，等效于 这样调用 `../cmds/ls.sh a "$a的值" b "$b的值" "$c的值" "$d的值"`  

在shell中你可以这样获取参数值:
```
echo "第0个参数:$0"
echo "第1个参数:$1"
echo "第2个参数:$2"
echo "第3个参数:$3"
echo "第4个参数:$4"
```

或者这样(使用环境变量)：
```
echo '$a的值:' $c2h_form_a
echo '$b的值:' $c2h_form_b
echo '$c的值:' $c2h_form_c
echo '$d的值:' $c2h_form_d
```
统一使用了前缀 <font color=blue>c2h_form_</font> 以和其他环境变量区分开！ 


##页面自定义
若需要自定义首页，可以使用`/s/index.html`  
或者使用`/s/my.css` 和 `/s/my.js` 来自己对页面进行控制或者自定义  
`/s/` 是当前配置文件目录下的子目录  


```javascript
// /s/my.js example

// onsubmit
function form_sleep_submit(){
    var input_n=findByName(this,'n');
    if(input_n.val()&lt;10){
       jw.msg("param wrong!")
       return false;
      }
}

// jsonp oncall
function form_sleep_jsonp(data){
    alert(data.data)
}
```

## 作为 HTTP 服务调用
`http://127.0.0.1:8310/ls?format=plain&a=123&b=456&c=789` 
这样调用即可，参数 `format` 是用于控制输出内容格式的，接口调用的时候直接使用 `format=plain` 即可。  
下面运行界面截图中展现的连接地址即为 API 调用地址。  

##界面截图
![界面截图](http://hidu.github.io/cmd2http/screenshot/index.png)


