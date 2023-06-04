package internal

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/fsgo/fsgo/fshtml"
)

var htmls = make(map[string]string)

const formatLi = `
<li>format:<select name='format'>
	<option value=''>default</option>
	<option value='html'>html</option>
	<option value='plain'>plain</option>
	<option value='jsonp'>jsonp</option>
	</select>
</li>`

const cacheLiTPL = `
<li>cache:
<select name='cache'>
  <option value='yes'>yes(%ds)</option>
  <option value='no'>no</option>
</select>
</li>`

func (srv *Server) helpPageCreate() {
	tabsBd := "<div class='bd'>"
	groups := make(map[string][]string, len(srv.config.Commands))

	for name, item := range srv.config.Commands {
		groupName := item.getGroup()
		groups[groupName] = append(groups[groupName], name)

		tabsBd += "\n\n<div class='cmd_div' id='div_" + name + "' style='display:none'>\n"
		_formStr := ` <form action='/%s' methor='get' onsubmit='return form_check(this,"%s")' id='form_%s'>`
		tabsBd += fmt.Sprintf(_formStr, name, name, name)
		tabsBd += "<div class='note note-g'><div><b>URI</b> :&nbsp;/" + name + "</div>" +
			"<div><b>Command</b> :&nbsp;[&nbsp;" + item.Command + "&nbsp;]&nbsp;" +
			"<b>Timeout</b> :&nbsp;" + fmt.Sprintf("%.1f", item.getTimeout().Seconds()) + "s&nbsp;" +
			"<b>Cache</b> :&nbsp;" + fmt.Sprintf("%.1f", item.getCacheLife().Seconds()) + "s&nbsp;" +
			"</div>"
		if item.Intro != "" {
			tabsBd = tabsBd + "<div><b>Intro</b> :&nbsp;&nbsp;" + item.Intro + "</div>"
		}
		tabsBd = tabsBd + "</div>"
		tabsBd = tabsBd + "<fieldset><ul class='ul-1'>"
		for _, param := range item.paramsAll {
			if param.isValParam && param.name != "charset" && param.name != "format" {
				var placeholder string
				if param.Default != "" {
					placeholder = "placeholder='" + param.Default + "'"
				}
				if param.HTML != "" {
					placeholder += " " + param.HTML
				}
				tabsBd += "<li>" + param.name + ":"

				_paramValues := param.getValues()

				if len(_paramValues) == 0 {
					tabsBd += "<input class='r-text p_" + param.name + "' type='text' name='" + param.name + "' " + placeholder + ">"
				} else {
					var opts []fshtml.Element
					for _, _v := range _paramValues {
						_optionKey := _v
						_optionVal := _v
						_pos := strings.Index(_v, ":")
						if _pos > -1 {
							_optionKey = strings.TrimSpace(_v[:_pos])
							_optionVal = strings.TrimSpace(_v[_pos+1:])
						}

						opt := fshtml.NewAny("option")
						fshtml.SetValue(opt, _optionKey)
						opt.Body = fshtml.ToElements(fshtml.String(_optionVal))
						fshtml.SetSelected(opt, param.Default == _optionKey)
						opts = append(opts, opt)
					}
					se := fshtml.NewSelect(opts...)
					fshtml.SetName(se, param.name)
					fshtml.SetClass(se, "r-select", "p_"+param.name)
					fshtml.SetAttrNoValue(se, placeholder)
					bf, _ := se.HTML()
					tabsBd += string(bf)
				}
				tabsBd += "</li>\n"
			}
		}
		tabsBd += formatLi
		chs := item.getCharsets()
		if len(chs) > 1 && item.Charset != "null" {
			tabsBd += "<li>charset:<select name='charset'>"
			for _, charset := range chs {
				var selected string
				if charset == item.Charset {
					selected = "selected=selected"
				}
				tabsBd += "<option value='" + charset + "' " + selected + ">" + charset + "</option>"
			}
			tabsBd += "</select></li>\n"
		}
		if item.getCacheLife() > 0 && srv.cacheAble() {
			tabsBd += fmt.Sprintf(cacheLiTPL, item.Cache)
		}

		tabsBd += `</ul><div class='c'></div>
           <center>
                <input type='submit' class='btn'>
                <span style='margin-right:50px'>&nbsp;</span>
                <input type='reset' class='btn' onclick='form_reset(this.form)' title='reset the form and abort the request'>
            </center>
           </fieldset><br/>
            <div class='div_url'></div>
            <iframe id='ifr_` + item.name + `' src='about:_blank' style='border:none;width:99%;height:20px' onload='ifr_load(this)'></iframe>
            <div class='result'></div>
            </form>
            </div>`
	}

	tabsStr := tabsBd + "\n</div>"

	contentMenu := "<dl id='main_menu'>"
	groupNames := srv.config.groups()
	for _, groupName := range groupNames {
		if subNames, has := groups[groupName]; has {
			contentMenu += "<dt>" + groupName + "</dt>"
			sort.Strings(subNames)
			for _, name := range subNames {
				contentMenu += "<dd><a href='#" + name + "' onclick=\"show_cmd('" + name + "')\">" + name + "</a></dd>\n"
			}
		}
	}
	contentMenu += "</dl>"
	htmls["body"] = tabsStr
	htmls["menu"] = contentMenu
}

var helpOnce sync.Once

func (srv *Server) handlerHelp(w http.ResponseWriter, r *http.Request) {
	helpOnce.Do(srv.helpPageCreate)

	var tabsStr string
	if IsFileExists("./s/my.css") {
		tabsStr += "<link  type='text/css' rel='stylesheet' href='/s/my.css'>"
	}

	if IsFileExists("./s/my.js") {
		tabsStr += "<script src='/s/my.js'></script>"
	}
	tabsStr += htmls["body"]

	tpl, _ := template.New("page").Parse(helpTPL)
	values := make(map[string]string)
	values["version"] = version
	values["title"] = srv.config.Title
	values["content_body"] = tabsStr
	values["content_menu"] = htmls["menu"]
	values["intro"] = srv.config.Intro

	w.Header().Add("c2h", version)
	tpl.Execute(w, values)
}

func (srv *Server) index(w http.ResponseWriter, r *http.Request) {
	if !srv.checkAuth(w, r) {
		return
	}

	req := request{writer: w, req: r, srv: srv}
	req.Deal()
}

func (srv *Server) checkAuth(w http.ResponseWriter, r *http.Request) (ret bool) {
	if srv.config.BasicAuth == "" {
		return true
	}
	doLogin := func() {
		w.Header().Set("WWW-authenticate", `Basic realm="need login"`)
		w.Header().Set("Content-Type", "text/html;charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("auth required"))
	}
	user, psw, ok := r.BasicAuth()

	defer func() {
		if ret {
			return
		}
		log.Printf("login failed with user=%q psw=%q, RemoteAddr=%s\n", user, psw, r.RemoteAddr)
	}()

	if !ok {
		doLogin()
		return false
	}
	list := strings.Split(srv.config.BasicAuth, ";")
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		arr := strings.SplitN(item, ":", 2)
		if arr[0] == user && arr[1] == psw {
			return true
		}
	}
	doLogin()
	return false
}
