cmd2http
=========

make system command as http service

config file
<pre>    
{
   port:8310,
   title:"super"
   cmds:{
      pwd:{
          cmd:"pwd"
       },
      echo:{
         cmd:"echo -n $wd|你好 $a $b"
        }
   }
}
</pre>