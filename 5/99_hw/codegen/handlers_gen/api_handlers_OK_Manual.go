package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const xAuth = "100500"

type Res struct {
	Error  string      `json:"error"`
	Result interface{} `json:"response,omitempty"`
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

func (srv *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, `bad method`, http.StatusNotAcceptable)
		return
	}

	auth := r.Header.Get("X-Auth")
	if auth != xAuth {
		sendError(w, `unauthorized`, http.StatusForbidden)
		return
	}

	var v string
	var vInt int
	var errInt error

	params := CreateParams{}
	v = r.FormValue("login")
	if v == "" {
		sendError(w, `login`+` must me not empty`, http.StatusBadRequest)
		return
	}
	if len(v) < 10 {
		sendError(w, `login`+` len must be >= `+`10`, http.StatusBadRequest)
		return
	}
	params.Login = v

	v = r.FormValue("full_name")
	params.Name = v

	v = r.FormValue("status")
	if v == "" {
		v = "user"
	}
	if !isStringInList(v, []string{"user", "moderator", "admin"}) {
		sendError(w, `status`+` must be one of `+`[user, moderator, admin]`, http.StatusBadRequest)
		return
	}

	params.Status = v

	v = r.FormValue("age")
	vInt, errInt = strconv.Atoi(v)
	if errInt != nil {
		sendError(w, `age`+` must be int`, http.StatusBadRequest)
		return
	}
	if vInt < 0 {
		sendError(w, `age`+` must be >= `+`0`, http.StatusBadRequest)
		return
	}
	if vInt > 128 {
		sendError(w, `age`+` must be <= `+`128`, http.StatusBadRequest)
		return
	}
	params.Age = vInt

	var errCall interface{}
	res, errCall := srv.Create(r.Context(), params)
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

	resultStruct := Res{Result: res}
	resByte, errByte := json.Marshal(resultStruct)
	if errByte != nil {
		sendError(w, `JSON`, http.StatusInternalServerError)
		return
	}
	_, errByte = io.Writer(w).Write(resByte)
	if errByte != nil {
		sendError(w, `Write`, http.StatusInternalServerError)
		return
	}
}

func (srv *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	var v string

	params := ProfileParams{}
	v = r.FormValue("login")
	if v == "" {
		sendError(w, `login`+` must me not empty`, http.StatusBadRequest)
		return
	}
	params.Login = v

	var errCall interface{}
	res, errCall := srv.Profile(r.Context(), params)
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

	resultStruct := Res{Result: res}
	resByte, errByte := json.Marshal(resultStruct)
	if errByte != nil {
		sendError(w, `JSON`, http.StatusInternalServerError)
		return
	}
	_, errByte = io.Writer(w).Write(resByte)
	if errByte != nil {
		sendError(w, `Write`, http.StatusInternalServerError)
		return
	}
}

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		srv.handlerProfile(w, r)
	case "/user/create":
		srv.handlerCreate(w, r)
	default:
		sendError(w, "unknown method", http.StatusNotFound)
	}
}
