package serve

import (
	"fmt"
	"github.com/hidu/goutils"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

var htmls map[string]string = make(map[string]string)

func (cmd2 *Cmd2HttpServe) helpPageCreate() {
	if _, has := htmls["body"]; has {
		return
	}
	tabs_bd := "<div class='bd'>"
	groups := make(map[string][]string)

	for name, _conf := range cmd2.config.Cmds {
		if _, _has := groups[_conf.Group]; !_has {
			groups[_conf.Group] = []string{}
		}
		groups[_conf.Group] = append(groups[_conf.Group], name)

		tabs_bd += "\n\n<div class='cmd_div' id='div_" + name + "' style='display:none'>\n"
		_form_str := ` <form action='/%s' methor='get' onsubmit='return form_check(this,"%s")' id='form_%s'>`
		tabs_bd += fmt.Sprintf(_form_str, name, name, name)
		tabs_bd += "<div class='note note-g'><div><b>uri</b> :&nbsp;/" + name + "</div>" +
			"<div><b>command</b> :&nbsp;[&nbsp;" + _conf.CmdRaw +
			"&nbsp;]&nbsp;<b>timeout</b> :&nbsp;" + fmt.Sprintf("%d", _conf.Timeout) + "s</div>"
		if _conf.Intro != "" {
			tabs_bd = tabs_bd + "<div><b>intro</b> :&nbsp;&nbsp;" + _conf.Intro + "</div>"
		}
		tabs_bd = tabs_bd + "</div>"
		tabs_bd = tabs_bd + "<fieldset><ul class='ul-1'>"
		for _, _param := range _conf.paramsAll {
			if _param.isValParam && _param.Name != "charset" && _param.Name != "format" {
				if _param.ValuesFile != "" {
					_param.Values = LoadParamValuesFromFile(_param.ValuesFile)
				}
				placeholder := ""
				if _param.DefaultValue != "" {
					placeholder = "placeholder='" + _param.DefaultValue + "'"
				}
				if _param.Html != "" {
					placeholder += " " + _param.Html
				}
				tabs_bd += "<li>" + _param.Name + ":"
				if len(_param.Values) == 0 {
					tabs_bd += "<input class='r-text p_" + _param.Name + "' type='text' name='" + _param.Name + "' " + placeholder + ">"
				} else {
					options := utils.NewHtml_Options()
					for _, _v := range _param.Values {
						_option_key := _v
						_option_val := _v
						_pos := strings.Index(_v, ":")
						if _pos > -1 {
							_option_key = strings.TrimSpace(_v[:_pos])
							_option_val = strings.TrimSpace(_v[_pos+1:])
						}
						options.AddOption(_option_key, _option_val, _param.DefaultValue == _option_key)
					}
					tabs_bd += utils.Html_select(_param.Name, options, "class='r-select p_"+_param.Name+"'", placeholder)
					tabs_bd += "</select>\n"
				}
				tabs_bd += "</li>\n"
			}
		}
		tabs_bd += `<li>format:<select name='format'>
			           <option value=''>default</option>
			           <option value='html'>html</option>
			           <option value='plain'>plain</option>
			           <option value='jsonp'>jsonp</option>
			           </select></li>`
		if len(_conf.Charsetlist) > 1 && _conf.Charset != "null" {
			tabs_bd += "<li>charset:<select name='charset'>"
			for _, _charset := range _conf.Charsetlist {
				_selected := ""
				if _charset == _conf.Charset {
					_selected = "selected=selected"
				}
				tabs_bd += "<option value='" + _charset + "' " + _selected + ">" + _charset + "</option>"
			}
			tabs_bd += "</select></li>\n"
		}
		if _conf.CacheLife > 3 && cmd2.cacheAble {
			_cache_li_str := `
                  <li>cache:
                  <select name='cache'>
	                  <option value='yes'>yes(%ds)</option>
	                  <option value='no'>no</option>
                  </select>
                  </li>`
			tabs_bd += fmt.Sprintf(_cache_li_str, _conf.CacheLife)
		}

		tabs_bd += `</ul><div class='c'></div>
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

	tabs_str := tabs_bd + "\n</div>"

	content_menu := "<dl id='main_menu'>"
	for groupName, sub_names := range groups {
		content_menu += "<dt>" + groupName + "</dt>"
		for _, name := range sub_names {
			content_menu += "<dd><a href='#" + name + "' onclick=\"show_cmd('" + name + "')\">" + name + "</a></dd>\n"
		}
	}
	content_menu += "</dl>"
	htmls["body"] = tabs_str
	htmls["menu"] = content_menu
}

func (cmd2 *Cmd2HttpServe) myHandler_help(w http.ResponseWriter, r *http.Request) {
	title := cmd2.config.Title
	cmd2.helpPageCreate()
	tabs_str := ""
	if IsFileExists("./s/my.css") {
		tabs_str += "<link  type='text/css' rel='stylesheet' href='/s/my.css'>"
	}

	if IsFileExists("./s/my.js") {
		tabs_str += "<script src='/s/my.js'></script>"
	}
	tabs_str += htmls["body"]
	reg := regexp.MustCompile(`\s+`)
	tabs_str = reg.ReplaceAllString(tabs_str, " ")
	str := string(LoadRes("res/tpl/help.html"))
	str = reg.ReplaceAllString(str, " ")

	tpl, _ := template.New("page").Parse(str)
	values := make(map[string]string)
	values["version"] = version
	values["title"] = title
	values["content_body"] = tabs_str
	values["content_menu"] = htmls["menu"]
	values["intro"] = cmd2.config.Intro

	w.Header().Add("c2h", version)
	tpl.Execute(w, values)
}

func (cmd2 *Cmd2HttpServe) myHandler_root(w http.ResponseWriter, r *http.Request) {
	req := Request{writer: w, req: r, cmd2: cmd2}
	req.Deal()
}
