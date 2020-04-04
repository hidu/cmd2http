cmd2http
=========
convert system command as http service  
将cli程序(系统命令、脚本等)转换为HTTP服务  


## Install
```
go get -u github.com/hidu/cmd2http
```

## Run
```
./cmd2http -conf=./example/conf/cmd2http.json
```

访问首页: <http://localhost:8310/>  

**hello world demo:**  
```
          url : http://localhost:8310/echo?wd=hello&a=world
command exec : <b>echo -n hello world defaultValue</b>  
       config : <b>echo -n $wd $a $b|defaultValue </b>  
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

###主配置文件(cmd2http.json)
```json
{
   "port":8310,
   "title":"default title",
   "intro":"intro info",
   "timeout":30,
   "cache_dir":"../cache_data/",
   "charset":"utf-8",
   "charset_list": [
        "gbk",
        "utf-8"
    ],
   "cmds":{
      "pwd":{
          "cmd":"pwd",
          "intro":"cmd intor",
          "timeout":10
       },
      "echo":{
         "cmd":"echo -n $wd|你好 $a $b",
         "cache":120
        }
   }
}
```
命令配置(cmds)(如上的pwd，echo)也可以配置到单独的文件，位于上述配置文件(cmd2http.json)目录下的cmd目录下去。  
程序在运行的时候，会自动`Chdir`到配置文件 `cmd2http.json`的目录下去。  
所以在配置文件中写的路径都使用以此为基准目录的相对路径即可。 

配置项说明： 
*   port      : http 服务监听端口
*   title     : http 页面标题
*   intro     : 介绍
*   timeout   : 默认的运行超时时间
*   cache_dir : 运行cache存放的目录（单项命令中配置了cache项后生效）
*   charset    :   全局默认的编码，只用于html页面结果展现
*   charset_list : 全局默认的编码可选值，只用于html页面结果展现

###命令配置文件(json)
ls.json内容为：
```json
{
    "cmd": "../cmds/ls.sh a $a b $b $c $d|你好",
    "intro": "hello",
    "timeout": 3,
    "cache": 1800,
    "group": "分组1",
    "charset":"utf-8",
   "charset_list": [
        "gbk",
        "utf-8"
    ],
    "params": {
        "c": {
            "values": ["1","2","3" ],
            "html": "style='width:200px'"
        },
        "d": {
            "values_file": "../data/ls_d.csv"
        }
    }
}
```
配置项说明：  
*  cmd   :  待运行的命令，参数使用`$`前缀，如 `$a`,`$a1`,`$a_1`  
*  intro : 介绍
*  timeout : 当前命令的运行超时时间，若没有设置或者为0 则使用全局的 timeout
*  cache : 运行结果cache有效期，单位秒，为0 或者全局的cache_dir没有设置的时候不使用cache
*  params : 参数配置
*  params.c : 参数 `$c`的配置项
*  params.c.values : 参数 `$c`的可选值，用来在form表单中展现，只能是字符串,values有值的情况下使用select展现样式，否则为input=text
*  params.c.html : 参数 `$c`的form控件额外的 html代码块
*  params.c.values_file : 参数 `$c`的可选值录入文件(eg:[可选值示例文件](./example/data/ls_d.csv))
*  charset    :   默认的编码，只用于html页面结果展现
*  charset_list : 默认的编码可选值，只用于html页面结果展现
*  group     :  页面展现分组，默认为`default`

###命令如何读取参数
```
"cmd": "../cmds/ls.sh a $a b $b $c $d|你好"
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

##作为http服务调用
http://127.0.0.1:8310/ls?format=plain&a=123&b=456&c=789  
这样调用即可，参数format是用于控制输出内容格式的，接口调用的时候直接使用`format=plain`即可。  
下面运行界面截图中展现的连接地址即为api调用地址。  

##界面截图
![界面截图](http://hidu.github.io/cmd2http/screenshot/index.png)


