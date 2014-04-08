cmd2http
=========
convert system command as http service  
将cli程序(系统命令、脚本等)转换为http服务  


##build
use build.sh to compile,dest file is in the dest subdir  
windows users should by use the <b>cygwin</b>,because i use zip command to embed resource files(js and css).  

or you can download the binary here <http://pan.baidu.com/s/1ntuhNSH>  

##execute
<code>
./cmd2http -conf=../example/cmd2http.conf -port=8080
</code>

index page: <http://localhost:8080/>  

*execue the hello world  
> url: <http://localhost/<b>echo?wd=hello&a=world></b>  
> command exec: #<b>echo -n hello world defaultValue</b>  
> config: <b>echo -n $wd $a $b|defaultValue </b>  


##config demo
<pre>    
{
   port:8310,
   title:"default title"
   intro:"intro info"
   timeout:30
   cache_dir:"./cache_data/"
   cmds:{
      pwd:{
          cmd:"pwd",
          intro:"cmd intor",
          timeout:10
       },
      echo:{
         cmd:"echo -n $wd|你好 $a $b"
         cache:120
        }
   }
}
</pre>

##custon style page
use /s/ as static root  
use /s/index.html as index page  
you can use /s/my.css and /s/my.js to control the help page form  


<pre>
// /s/my.js example

function form_echo(){
    var input_n=findByName(this,'n');
    if(input_n.val()&lt;10){
       jw.msg("param wrong!")
       return false;
      }
}
</pre>


