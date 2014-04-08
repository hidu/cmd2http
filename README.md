cmd2http
=========
make system command as http service

#build
use build.sh to compile,dest file is in the dest subdir
windows users should by use the <b>cygwin</b>,because i use zip command to embed resource files(js and css).

#useage

##execute
./cmd2http -conf=../example/cmd2http.conf -port=8080

##visit with browser
there is an index (or help) page that you can use it more easy.
the url is : http://localhost:8080/

###call a command
<pre>
http://localhost/<b>echo?wd=hello&a=world</b> 

it eq the command:
<b>#echo -n hello world defaultValue</b> 

config is very simple:
<b>echo -n $wd $a $b|defaultValue </b>
</pre>

##config demo
<pre>    
{
   port:8310,
   title:"default title"
   intro:"intro info"
   timeout:30
   cmds:{
      pwd:{
          cmd:"pwd",
          intro:"cmd intor",
          timeout:10
       },
      echo:{
         cmd:"echo -n $wd|你好 $a $b"
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


