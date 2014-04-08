cmd2http
=========
make system command as http service

将cli程序转换为http服务

#build
use build.sh to compile,dest file is in the dest subdir

windows users should by use the <b>cygwin</b>,because i use zip command to embed resource files(js and css).

#useage

##execute
./cmd2http -conf=../example/cmd2http.conf -port=8080

##visit with browser
> there is an index (or help) page that you can use it more easy.
> the url is : <a>http://localhost:8080/</a>

###call a command

> url: http://localhost/<b>echo?wd=hello&a=world</b> 
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

##custon index page
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


