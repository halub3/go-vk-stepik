package main

import "text/template"

var (
	TplInit = template.Must(template.New("initTpl").Parse(`
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)
const xAuth = "100500"

type Res struct {
	Error  string      ` + "`" + `json:"error"` + "`" + `
	Result interface{} ` + "`" + `json:"response,omitempty"` + "`" + `
}

func sendError(w http.ResponseWriter, text string, code int) {
	w.WriteHeader(code)
	resStruct := Res{Error: text}
	resByte, _ := json.Marshal(resStruct)
	io.Writer(w).Write(resByte)
}

func isStringInList(search interface{}, l []string) bool {
	for _, s := range l {
		if fmt.Sprint(search) == s {
			return true
		}
	}
	return false
}
`))

	TplFuncInitTmpl = template.Must(template.New("funcInitTmpl").Parse(`
func (srv *{{.RecvName}}) handler{{.Name}}(w http.ResponseWriter, r *http.Request) {
`))

	TplFuncCheckMethod = template.Must(template.New("FuncCheckMethod").Parse(`
	if r.Method != "{{.ApiInput.Method}}" {
		sendError(w, "bad method", http.StatusNotAcceptable)
		return
	}
`))

	TplFuncCheckAuth = template.Must(template.New("FuncCheckAuth").Parse(`
	auth := r.Header.Get("X-Auth")
	if auth != xAuth {
		sendError(w, "unauthorized", http.StatusForbidden)
		return
	}
`))

	TplFuncInitVars = template.Must(template.New("TplFuncInitVars").Parse(`
	params := {{.FieldName}}{}
	var val string
`))

	TplFuncReadField = template.Must(template.New("TplFuncReadField").Parse(`
	val = r.FormValue("{{.ParamName}}")
	{{if .IsDefault}}if val == "" {
		val = "{{.Default}}"
	}{{end}} {{if eq .Type "int"}}vInt{{.Name}}, vIntErr{{.Name}} := strconv.Atoi(val)
	if vIntErr{{.Name}} != nil {
		sendError(w, "{{.ParamName}}"+" must be int", http.StatusBadRequest)
		return
	}{{end}}
	{{ if .IsRequired }}if val == "" {
		sendError(w, "{{.ParamName}}"+" must me not empty", http.StatusBadRequest)
		return
	}{{ end }}
	{{ if .IsMin }}{{ if eq .Type "string" }}if len(val) < {{.Min}} {
		sendError(w, "{{.ParamName}}"+" len must be >= "+"{{.Min}}", http.StatusBadRequest)
		return
	}{{ else }}if vInt{{.Name}} < {{.Min}} {
		sendError(w, "{{.ParamName}}"+" must be >= "+"{{.Min}}", http.StatusBadRequest)
		return
	}{{ end }}{{end}}
	{{ if .IsMax }}{{ if eq .Type "string" }}if len(val) > {{.Max}} {
		sendError(w, "{{.ParamName}}"+" len must be <= "+"{{.Max}}", http.StatusBadRequest)
		return
	}{{ else }}if vInt{{.Name}} > {{.Max}} {
		sendError(w, "{{.ParamName}}"+" must be <= "+"{{.Max}}", http.StatusBadRequest)
		return
	}{{ end }}{{end}}
	{{if .IsEnum}}
	if !isStringInList(val, []string{ {{.EnumSrch}} }) {
		sendError(w, "{{.ParamName}}"+" must be one of "+"[{{.EnumErr}}]", http.StatusBadRequest)
		return
	}{{end}}
	params.{{.Name}} = {{ if eq .Type "string" }} val {{ else }} vInt{{.Name}} {{ end }}
`))

	TplFuncCall = template.Must(template.New("TplFuncCall").Parse(`
	var errCall interface{}
	res, errCall := srv.{{.Name}}(r.Context(), params)
	if errCall != nil {
		errFormat, ok := errCall.(ApiError)
		if !ok {
			errFormat := errCall.(error)
			sendError(w, errFormat.Error(), http.StatusInternalServerError)
		} else {
			sendError(w, errFormat.Error(), errFormat.HTTPStatus)
		}
		return
	}
`))

	TplFuncSendRes = template.Must(template.New("TplFuncSendRes").Parse(`
	resultStruct := Res{Result: res}
	resByte, errByte := json.Marshal(resultStruct)
	if errByte != nil {
		sendError(w, "JSON", http.StatusInternalServerError)
		return
	}
	_, errByte = io.Writer(w).Write(resByte)
	if errByte != nil {
		sendError(w, "Write", http.StatusInternalServerError)
		return
	}
`))

	TplFuncEnd = template.Must(template.New("TplFuncReadField").Parse(`}`))

	TplSHTTPInit = template.Must(template.New("TplSHTTPInit").Parse(`
func (srv *{{.Name}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
`))

	TplSHTTPCase = template.Must(template.New("TplSHTTPCase").Parse(`
	case "{{.ApiInput.Url}}":
		srv.handler{{.Name}}(w,r)
`))

	TplSHTTPEnd = template.Must(template.New("TplFuncReadField").Parse(`
	default:
			sendError(w, "unknown method", http.StatusNotFound)
		}
	}
`))
)
