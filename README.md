cmd2http
=========
convert system command as http service  
将cli程序(系统命令、脚本等)转换为http服务  


##install
安装了golang的用户:  
> go get -u github.com/hidu/cmd2http

或者下载编译二进制： <http://pan.baidu.com/s/1bnkyWLD#path=%252Fcmd2http>  

##运行
```
./cmd2http -conf=./example/conf/cmd2http.json
```

访问首页: <http://localhost:8080/>  

**hello world demo:**  
```
          url : http://localhost:8080/echo?wd=hello&a=world
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
```javascript
{
   "port":8310,
   "title":"default title",
   "intro":"intro info",
   "timeout":30,
   "cache_dir":"../cache_data/",
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


###命令配置文件(json)
ls.json内容为：
```javascript
{
    "cmd": "../cmds/ls.sh a $a b $b $c $d|你好",
    "intro": "hello",
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
*  cmd :  待运行的命令，参数使用`$`前缀，如 `$a`,`$a1`,`$a_1`  
*  intro : 介绍
*  params : 参数配置
*  params.c : 参数 `$c`的配置项
*  params.c.values : 参数 `$c`的可选值，用来在form表单中展现，只能是字符串,values有值的情况下使用select展现样式，否则为input=text
*  params.c.html : 参数 `$c`的form控件额外的 html代码块
*  params.c.values_file : 参数 `$c`的可选值录入文件(eg:[可选值示例文件](./example/data/ls_d.csv))


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


