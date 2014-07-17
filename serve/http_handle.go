package serve
import (
  "fmt"
  "net/http"
  "regexp"
  "text/template"
  "github.com/hidu/goutils"
  "time"
  "mime"
  "path/filepath"
  "strings"
)
var htmls map[string]string=make(map[string]string)

func (cmd2 *Cmd2HttpServe)helpPageCreate(){
//        if _,has:=htmls["body"];has{
//         return
//        }
         tabs_bd:="<div class='bd'>";
        groups:=make(map[string][]string)
       
       for name,_conf:=range cmd2.CmdConfs{
           if _,_has:=groups[_conf.group];!_has{
            groups[_conf.group]=[]string{}
               }
           groups[_conf.group]=append(groups[_conf.group],name)
            
           tabs_bd+="\n\n<div class='cmd_div' id='div_"+name+"' style='display:none'>\n"
           _form_str:=` <form action='/%s' methor='get' onsubmit='return form_check(this,"%s")' id='form_%s'>`;
           tabs_bd+=fmt.Sprintf(_form_str,name,name,name)
           tabs_bd+="<div class='note note-g'><div><b>uri</b> :&nbsp;/"+name+"</div>"+
           "<div><b>command</b> :&nbsp;[&nbsp;"+_conf.cmdStr+
           "&nbsp;]&nbsp;<b>timeout</b> :&nbsp;"+fmt.Sprintf("%d",_conf.timeout)+"s</div>"
           if(_conf.intro!=""){
             tabs_bd=tabs_bd+"<div><b>intro</b> :&nbsp;&nbsp;"+_conf.intro+"</div>"
              }
           tabs_bd=tabs_bd+"</div>";
           tabs_bd=tabs_bd+"<fieldset><ul class='ul-1'>"
              for _,_param:=range _conf.params{
                if(_param.isValParam && _param.name!="charset" && _param.name!="format"){
                   if(_param.values_file!=""){
                     _param.values=LoadParamValuesFromFile(_param.values_file)
                         }
                   placeholder:=""
                   if(_param.defaultValue!=""){
                      placeholder="placeholder='"+_param.defaultValue+"'"
                         }
                   if(_param.html!=""){
                      placeholder+=" "+_param.html
                          }
                   tabs_bd+="<li>"+_param.name+":"
                   if(len(_param.values)==0){
                      tabs_bd+="<input class='r-text p_"+_param.name+"' type='text' name='"+_param.name+"' "+placeholder+">";
                   }else{
                      options:=utils.NewHtml_Options()
                       for _,_v:=range _param.values{
                         _option_key:=_v
                         _option_val:=_v
                         _pos:=strings.Index(_v,":")
                         if(_pos>-1){
                         _option_key=strings.TrimSpace(_v[:_pos])
                         _option_val=strings.TrimSpace(_v[_pos+1:])
                                 }
                         options.AddOption(_option_key,_option_val,_param.defaultValue==_option_key)
                              }
                      tabs_bd+=utils.Html_select(_param.name,options,"class='r-select p_"+_param.name+"'",placeholder)
                      tabs_bd+="</select>\n";
                         }
                   tabs_bd+="</li>\n"
                     }
                   }
           tabs_bd+=`<li>format:<select name='format'>
			           <option value=''>default</option>
			           <option value='html'>html</option>
			           <option value='plain'>plain</option>
			           <option value='jsonp'>jsonp</option>
			           </select></li>`;
           if(len(_conf.charset_list)>1 && _conf.charset!="null"){
               tabs_bd+="<li>charset:<select name='charset'>"
               for _,_charset:=range _conf.charset_list{
                       _selected:="";
                       if(_charset==_conf.charset){
                               _selected="selected=selected";
                              }
                   tabs_bd+="<option value='"+_charset+"' "+_selected+">"+_charset+"</option>"
                  }
               tabs_bd+="</select></li>\n"
           }
          if(_conf.cache_life>3 && cmd2.cacheAble){
                  _cache_li_str:=`
                  <li>cache:
                  <select name='cache'>
	                  <option value='yes'>yes(%ds)</option>
	                  <option value='no'>no</option>
                  </select>
                  </li>`
                  tabs_bd+=fmt.Sprintf(_cache_li_str,_conf.cache_life)
              }
           
           tabs_bd+=`</ul><div class='c'></div>
           <center>
                <input type='submit' class='btn'>
                <span style='margin-right:50px'>&nbsp;</span>
                <input type='reset' class='btn' onclick='form_reset(this.form)' title='reset the form and abort the request'>
            </center>
           </fieldset><br/>
            <div class='div_url'></div>
            <iframe id='ifr_`+_conf.name+`' src='about:_blank' style='border:none;width:99%;height:20px' onload='ifr_load(this)'></iframe>
            <div class='result'></div>
            </form>
            </div>`;
          }
        
      tabs_str:=tabs_bd+"\n</div>";
        
      content_menu:="<dl id='main_menu'>"
      for _,groupName :=range confMap_groups{
          content_menu+="<dt>"+groupName+"</dt>"
          sub_names:=groups[groupName]
         for _,name:=range sub_names{
           content_menu+="<dd><a href='#"+name+"' onclick=\"show_cmd('"+name+"')\">"+name+"</a></dd>\n";
            }
        }
      content_menu+="</dl>"
      htmls["body"]=tabs_str
      htmls["menu"]=content_menu
}

func (cmd2 *Cmd2HttpServe)myHandler_help(w http.ResponseWriter, r *http.Request){
        title:=cmd2.Config.String("title","")
       cmd2.helpPageCreate()
       tabs_str:=""
       if(IsFileExists("./s/my.css")){
        tabs_str+="<link  type='text/css' rel='stylesheet' href='/s/my.css'>";
        }
      
      if(IsFileExists("./s/my.js")){
        tabs_str+="<script src='/s/my.js'></script>";
        }
       tabs_str+=htmls["body"]
       reg:=regexp.MustCompile(`\s+`)
       tabs_str=reg.ReplaceAllString(tabs_str," ")
       str:=string(LoadRes("res/tpl/help.html"));
       str=reg.ReplaceAllString(str," ")
       
       tpl,_:=template.New("page").Parse(str)
       values :=make(map[string]string)
       values["version"]=version
       values["title"]=title
       values["content_body"]=tabs_str
       values["content_menu"]=htmls["menu"]
       values["intro"]=cmd2.Config.String("intro","")
       
       
       w.Header().Add("c2h",version)
       tpl.Execute(w,values)
}


func (cmd2 *Cmd2HttpServe)myHandler_root(w http.ResponseWriter, r *http.Request){
     req:=Request{writer:w,req:r,cmd2:cmd2}
     req.Deal()
}


func response_res(w http.ResponseWriter,r *http.Request,path string){
    res,err:=GetRes(path)
    if(err!=nil){
        w.WriteHeader(404)
        return;
     }
    finfo,_:=res.Stat()
    modtime:=finfo.ModTime()
    if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
            h := w.Header()
             delete(h, "Content-Type")
           delete(h, "Content-Length")
           w.WriteHeader(http.StatusNotModified)
           return
           }
   mimeType:= mime.TypeByExtension(filepath.Ext(path))
   if(mimeType!=""){
       w.Header().Set("Content-Type",mimeType)
     }
    w.Header().Set("Last-Modified",modtime.UTC().Format(http.TimeFormat))
    w.Write(LoadRes(path))
}

func myHandler_res(w http.ResponseWriter, r *http.Request){
    response_res(w,r,r.URL.Path)
}

