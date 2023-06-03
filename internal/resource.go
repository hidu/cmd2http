// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/2

package internal

import (
	"embed"
)

//go:embed resource/static/*
var resourceWeb embed.FS

//go:embed resource/static/css/favicon.ico
var favicon embed.FS

//go:embed resource/tpl/help.html
var helpTPL string
