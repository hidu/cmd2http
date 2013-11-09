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
