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

func (srv *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {

	params := ProfileParams{}
	var val string

	val = r.FormValue("login")
	 
	if val == "" {
		sendError(w, "login"+" must me not empty", http.StatusBadRequest)
		return
	}
	
	
	
	params.Login =  val 

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
		sendError(w, "JSON", http.StatusInternalServerError)
		return
	}
	_, errByte = io.Writer(w).Write(resByte)
	if errByte != nil {
		sendError(w, "Write", http.StatusInternalServerError)
		return
	}
}
func (srv *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		sendError(w, "bad method", http.StatusNotAcceptable)
		return
	}

	auth := r.Header.Get("X-Auth")
	if auth != xAuth {
		sendError(w, "unauthorized", http.StatusForbidden)
		return
	}

	params := CreateParams{}
	var val string

	val = r.FormValue("login")
	 
	if val == "" {
		sendError(w, "login"+" must me not empty", http.StatusBadRequest)
		return
	}
	if len(val) < 10 {
		sendError(w, "login"+" len must be >= "+"10", http.StatusBadRequest)
		return
	}
	
	
	params.Login =  val 

	val = r.FormValue("full_name")
	 
	
	
	
	
	params.Name =  val 

	val = r.FormValue("status")
	if val == "" {
		val = "user"
	} 
	
	
	
	
	if !isStringInList(val, []string{ "user", "moderator", "admin" }) {
		sendError(w, "status"+" must be one of "+"[user, moderator, admin]", http.StatusBadRequest)
		return
	}
	params.Status =  val 

	val = r.FormValue("age")
	 vIntAge, vIntErrAge := strconv.Atoi(val)
	if vIntErrAge != nil {
		sendError(w, "age"+" must be int", http.StatusBadRequest)
		return
	}
	
	if vIntAge < 0 {
		sendError(w, "age"+" must be >= "+"0", http.StatusBadRequest)
		return
	}
	if vIntAge > 128 {
		sendError(w, "age"+" must be <= "+"128", http.StatusBadRequest)
		return
	}
	
	params.Age =  vIntAge 

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
		sendError(w, "JSON", http.StatusInternalServerError)
		return
	}
	_, errByte = io.Writer(w).Write(resByte)
	if errByte != nil {
		sendError(w, "Write", http.StatusInternalServerError)
		return
	}
}
func (srv *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		sendError(w, "bad method", http.StatusNotAcceptable)
		return
	}

	auth := r.Header.Get("X-Auth")
	if auth != xAuth {
		sendError(w, "unauthorized", http.StatusForbidden)
		return
	}

	params := OtherCreateParams{}
	var val string

	val = r.FormValue("username")
	 
	if val == "" {
		sendError(w, "username"+" must me not empty", http.StatusBadRequest)
		return
	}
	if len(val) < 3 {
		sendError(w, "username"+" len must be >= "+"3", http.StatusBadRequest)
		return
	}
	
	
	params.Username =  val 

	val = r.FormValue("account_name")
	 
	
	
	
	
	params.Name =  val 

	val = r.FormValue("class")
	if val == "" {
		val = "warrior"
	} 
	
	
	
	
	if !isStringInList(val, []string{ "warrior", "sorcerer", "rouge" }) {
		sendError(w, "class"+" must be one of "+"[warrior, sorcerer, rouge]", http.StatusBadRequest)
		return
	}
	params.Class =  val 

	val = r.FormValue("level")
	 vIntLevel, vIntErrLevel := strconv.Atoi(val)
	if vIntErrLevel != nil {
		sendError(w, "level"+" must be int", http.StatusBadRequest)
		return
	}
	
	if vIntLevel < 1 {
		sendError(w, "level"+" must be >= "+"1", http.StatusBadRequest)
		return
	}
	if vIntLevel > 50 {
		sendError(w, "level"+" must be <= "+"50", http.StatusBadRequest)
		return
	}
	
	params.Level =  vIntLevel 

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
		sendError(w, "JSON", http.StatusInternalServerError)
		return
	}
	_, errByte = io.Writer(w).Write(resByte)
	if errByte != nil {
		sendError(w, "Write", http.StatusInternalServerError)
		return
	}
}
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/profile":
		srv.handlerProfile(w,r)

	case "/user/create":
		srv.handlerCreate(w,r)

	default:
			sendError(w, "unknown method", http.StatusNotFound)
		}
	}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/create":
		srv.handlerCreate(w,r)

	default:
			sendError(w, "unknown method", http.StatusNotFound)
		}
	}
