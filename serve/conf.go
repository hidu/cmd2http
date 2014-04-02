package serve

import(
      "fmt"
      "regexp"
      "strings"
      "log"
)
type param struct{
  name string
  defaultValue string
  isValParam bool
  values []string
  html string
  values_file string
}

func (p *param)ToString() string{
    return fmt.Sprintf("name:%s,default:%s,isValParam:%x",p.name,p.defaultValue,p.isValParam);
}

type Conf struct{
   name string
   cmdStr string
   cmd string
   charset string
   params []*param
   intro string
   timeout int
   charset_list []string
   group string
   cache_life int64
}

var confMap_groups []string

var charset_list []string

var charset_default string


func (cmd2 *Cmd2HttpServe)ParseConfig(){
    config:=cmd2.Config
    
    if cmd2.Port==0 {
       cmd2.Port=config.Int("port",8310)
     }
    cmd2.logPath=config.String("log_path","./cmd2http.log")
        
    charset_list=config.StringList("charset_list",[]string{})
    
    charset_default=config.String("charset","utf-8");
    cmd2.cacheDirPath=config.String("cache_dir","")
    
    if(!In_array(charset_default,charset_list)){
       charset_list=append(charset_list,charset_default);   
    }
    cmd2.Charset_list=charset_list
    cmd2.Charset=charset_default
    
    timeout:=config.Int("timeout",30)
    if(timeout<1){
      timeout=1
    }
    
    cmd2.CmdConfs=make(map[string]*Conf)
    confMap_groups=make([]string,0,10)
    
    cmds:=config.Object("cmds",make(map[string]interface{}))
    
    for k,_:=range cmds{
       conf_path_pre:="cmds."+k+"."
       
       conf:=new(Conf)
       conf.name=k
       conf.timeout=timeout
       conf.group=config.String(conf_path_pre+"group","default")
       
       if (!In_array(conf.group,confMap_groups)){
           confMap_groups=append(confMap_groups,conf.group)
       }
       
       conf.charset=config.String(conf_path_pre+"charset",charset_default)
      conf.intro=config.String(conf_path_pre+"intro","")
       
       conf.charset_list=config.StringList(conf_path_pre+"charset_list",charset_list)
       
        if(!In_array(conf.charset,conf.charset_list)){
               conf.charset_list=append(conf.charset_list,conf.charset);   
        }
       
       conf.timeout=config.Int(conf_path_pre+"timeout",timeout)

       conf.cmdStr=config.String(conf_path_pre+"cmd","")
       
       conf.cmdStr=strings.TrimSpace(conf.cmdStr)
       conf.params=make([]*param,0,10)
       
       conf.cache_life=int64(config.Int(conf_path_pre+"cache",0))
//       fmt.Println("conf.cache_life",conf.cache_life)
       ps:=regexp.MustCompile(`\s+`).Split(conf.cmdStr,-1)
//       fmt.Println(ps)
       conf.cmd=ps[0]
       
       for i:=1;i<len(ps);i++ {
           item:=ps[i]
//           fmt.Println("i:",i,item)
            _param:=new(param)
            _param.name=item
           
           if(item[0]=='$'){
            _param.isValParam=true;
            tmp:=strings.Split(item+"|","|")
            _param.name=tmp[0][1:]
            _param.defaultValue=tmp[1]
           _param.html=config.String(conf_path_pre+"params."+_param.name+".html","")
           _param.values=config.StringList(conf_path_pre+"params."+_param.name+".values",[]string{})
           _param.values_file=config.String(conf_path_pre+"params."+_param.name+".values_file","")
            }
           conf.params=append(conf.params,_param)
//           fmt.Println(_param.name,_param.defaultValue)
        }
       log.Println("register[",k,"] cmd:",conf.cmdStr)
       cmd2.CmdConfs[k]=conf
    }
}