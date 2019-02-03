package serve

import (
	"fmt"
	"github.com/hidu/goutils/html_util"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var htmls = make(map[string]string)

func (cmd2 *Cmd2HttpServe) helpPageCreate() {
	//	if _, has := htmls["body"]; has {
	//		return
	//	}
	tabsBd := "<div class='bd'>"
	groups := make(map[string][]string)

	for name, _conf := range cmd2.config.Cmds {
		if _, _has := groups[_conf.Group]; !_has {
			groups[_conf.Group] = []string{}
		}
		groups[_conf.Group] = append(groups[_conf.Group], name)

		tabsBd += "\n\n<div class='cmd_div' id='div_" + name + "' style='display:none'>\n"
		_formStr := ` <form action='/%s' methor='get' onsubmit='return form_check(this,"%s")' id='form_%s'>`
		tabsBd += fmt.Sprintf(_formStr, name, name, name)
		tabsBd += "<div class='note note-g'><div><b>uri</b> :&nbsp;/" + name + "</div>" +
			"<div><b>command</b> :&nbsp;[&nbsp;" + _conf.CmdRaw +
			"&nbsp;]&nbsp;<b>timeout</b> :&nbsp;" + fmt.Sprintf("%d", _conf.Timeout) + "s</div>"
		if _conf.Intro != "" {
			tabsBd = tabsBd + "<div><b>intro</b> :&nbsp;&nbsp;" + _conf.Intro + "</div>"
		}
		tabsBd = tabsBd + "</div>"
		tabsBd = tabsBd + "<fieldset><ul class='ul-1'>"
		for _, _param := range _conf.paramsAll {
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
		tabsBd += `<li>format:<select name='format'>
			           <option value=''>default</option>
			           <option value='html'>html</option>
			           <option value='plain'>plain</option>
			           <option value='jsonp'>jsonp</option>
			           </select></li>`
		if len(_conf.Charsetlist) > 1 && _conf.Charset != "null" {
			tabsBd += "<li>charset:<select name='charset'>"
			for _, _charset := range _conf.Charsetlist {
				_selected := ""
				if _charset == _conf.Charset {
					_selected = "selected=selected"
				}
				tabsBd += "<option value='" + _charset + "' " + _selected + ">" + _charset + "</option>"
			}
			tabsBd += "</select></li>\n"
		}
		if _conf.CacheLife > 3 && cmd2.cacheAble {
			_cacheLiStr := `
                  <li>cache:
                  <select name='cache'>
	                  <option value='yes'>yes(%ds)</option>
	                  <option value='no'>no</option>
                  </select>
                  </li>`
			tabsBd += fmt.Sprintf(_cacheLiStr, _conf.CacheLife)
		}

		tabsBd += `</ul><div class='c'></div>
           <center>
                <input type='submit' class='btn'>
                <span style='margin-right:50px'>&nbsp;</span>
                <input type='reset' class='btn' onclick='form_reset(this.form)' title='reset the form and abort the request'>
            </center>
           </fieldset><br/>
            <div class='div_url'></div>
            <iframe id='ifr_` + _conf.Name + `' src='about:_blank' style='border:none;width:99%;height:20px' onload='ifr_load(this)'></iframe>
            <div class='result'></div>
            </form>
            </div>`
	}

	tabsStr := tabsBd + "\n</div>"

	contentMenu := "<dl id='main_menu'>"
	groupNames := cmd2.config.groups()
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

func (cmd2 *Cmd2HttpServe) myHandlerHelp(w http.ResponseWriter, r *http.Request) {
	title := cmd2.config.Title
	cmd2.helpPageCreate()
	tabsStr := ""
	if IsFileExists("./s/my.css") {
		tabsStr += "<link  type='text/css' rel='stylesheet' href='/s/my.css'>"
	}

	if IsFileExists("./s/my.js") {
		tabsStr += "<script src='/s/my.js'></script>"
	}
	tabsStr += htmls["body"]
	reg := regexp.MustCompile(`\s+`)
	tabsStr = reg.ReplaceAllString(tabsStr, " ")
	str := Assest.GetContent("res/tpl/help.html")
	str = reg.ReplaceAllString(str, " ")

	tpl, _ := template.New("page").Parse(str)
	values := make(map[string]string)
	values["version"] = version
	values["title"] = title
	values["content_body"] = tabsStr
	values["content_menu"] = htmls["menu"]
	values["intro"] = cmd2.config.Intro

	w.Header().Add("c2h", version)
	tpl.Execute(w, values)
}

func (cmd2 *Cmd2HttpServe) myHandlerRoot(w http.ResponseWriter, r *http.Request) {
	req := request{writer: w, req: r, cmd2: cmd2}
	req.Deal()
}
