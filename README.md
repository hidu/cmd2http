cmd2http
=========

make system command as http service
<pre>
echo -n $wd $a $b|defaultValue 

http://localhost/<b>echo?wd=hello&a=world</b> ==&gt;   <b>#echo -n hello world defaultValue</b> 
</pre>
config file
<pre>    
{
   port:8310,
   title:"super"
   intro:"intro msg"
   cmds:{
      pwd:{
          cmd:"pwd",
          intro:"cmd intor",
       },
      echo:{
         cmd:"echo -n $wd|你好 $a $b"
        }
   }
}
</pre>

use /s/ as static root
use /s/index.html as index page

usr /s/my.js to control the help page form

<pre>
// /s/my.js example

function form_echo{
    if(1>2){//some condtion
       jw.msg("param wrong!")
       return false;
      }
}
</pre>