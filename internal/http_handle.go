package internal

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/hidu/goutils/html_util"
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
			"<div><b>Command</b> :&nbsp;[&nbsp;" + item.Command +
			"&nbsp;]&nbsp;<b>Timeout</b> :&nbsp;" + fmt.Sprintf("%.1f", item.getTimeout().Seconds()) + "s</div>"
		if item.Intro != "" {
			tabsBd = tabsBd + "<div><b>Intro</b> :&nbsp;&nbsp;" + item.Intro + "</div>"
		}
		tabsBd = tabsBd + "</div>"
		tabsBd = tabsBd + "<fieldset><ul class='ul-1'>"
		for _, _param := range item.paramsAll {
			if _param.isValParam && _param.Name != "charset" && _param.Name != "format" {
				placeholder := ""
				if _param.DefaultValue != "" {
					placeholder = "placeholder='" + _param.DefaultValue + "'"
				}
				if _param.HTML != "" {
					placeholder += " " + _param.HTML
				}
				tabsBd += "<li>" + _param.Name + ":"

				_paramValues := _param.getValues()

				if len(_paramValues) == 0 {
					tabsBd += "<input class='r-text p_" + _param.Name + "' type='text' name='" + _param.Name + "' " + placeholder + ">"
				} else {
					options := html_util.NewHtml_Options()
					for _, _v := range _paramValues {
						_optionKey := _v
						_optionVal := _v
						_pos := strings.Index(_v, ":")
						if _pos > -1 {
							_optionKey = strings.TrimSpace(_v[:_pos])
							_optionVal = strings.TrimSpace(_v[_pos+1:])
						}
						options.AddOption(_optionKey, _optionVal, _param.DefaultValue == _optionKey)
					}
					tabsBd += html_util.Html_select(_param.Name, options, "class='r-select p_"+_param.Name+"'", placeholder)
					tabsBd += "</select>\n"
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
		if item.CacheLife > 3 && srv.cacheAble() {
			tabsBd += fmt.Sprintf(cacheLiTPL, item.CacheLife)
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
	req := request{writer: w, req: r, cmd2: srv}
	req.Deal()
}
