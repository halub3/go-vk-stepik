package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

/*
Структура с тестами, которые будем заполнять
*/
type TestCase struct {
	Search SearchRequest
	Result SearchResponse
	Token  string
	Error  error
}

/*
Ошибки с фиксированным текстом
*/
var (
	errLimitMustBeMoreZero    = errors.New("limit must be > 0")
	errOffsetMustBeMoreZero   = errors.New("offset must be > 0")
	errBadAccessToken         = errors.New("Bad AccessToken")
	errSearchServerFatalError = errors.New("SearchServer fatal error")
)

/*
	Блок создания ошибок с кастомным текстом
*/
/* START ------------------------------------------------------------------------------------- */
func errorBadOrderField(orderField string) error {
	return errors.New("OrderFeld " + orderField + " invalid")
}

func errorCannotUnpackErrJSON(text string) error {
	return errors.New("cant unpack error json: " + text)
}

func errorTimeout(text string) error {
	return errors.New("timeout for " + text)
}

func errUnknownError(text string) error {
	return errors.New("unknown error " + text)
}

func errUnknowBadRequestError(text string) error {
	return errors.New("unknown bad request error: " + text)
}

func errCannotUnpackResultJSON(text string) error {
	return errors.New("cant unpack result json: " + text)
}

/* ------------------------------------------------------------------------------------- END */

/*
Тесты с ошибками по флоу работы сервера (которые он может генерить сам)
*/
func TestSearchNegative(t *testing.T) {
	cases := []TestCase{
		{ // limit must be > 1
			Search: SearchRequest{
				Limit: -100,
			},
			Token: "qwe123asd",
			Error: errLimitMustBeMoreZero,
		},
		{ // limit must be > 1
			Search: SearchRequest{
				Offset: -100,
			},
			Token: "qwe123asd",
			Error: errOffsetMustBeMoreZero,
		},
		{ // bad AccessToken
			Search: SearchRequest{},
			Token:  "1234",
			Error:  errBadAccessToken,
		},
		{ // bad OrderField
			Search: SearchRequest{OrderField: "BADFIELD"},
			Token:  "qwe123asd",
			Error:  errorBadOrderField("BADFIELD"),
		},
		{ // bad OrderBy
			Search: SearchRequest{OrderBy: 100},
			Token:  "qwe123asd",
			Error:  errUnknowBadRequestError("ErrorBadOrderBy"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, item := range cases {
		sc := SearchClient{AccessToken: item.Token, URL: ts.URL}
		_, errSearch := sc.FindUsers(item.Search)
		if !reflect.DeepEqual(errSearch, item.Error) {
			t.Errorf("[%d] Wrong error, got `%v`, but need `%v`\n", caseNum, errSearch, item.Error)
		}

	}
}

/*
Отдельный тест на ошибку чтения файла, подменяем файл в начале перед запуском
*/
func TestBadReadFileFatalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	item := TestCase{
		Token: "qwe123asd",
		Error: errSearchServerFatalError,
	}
	fileName = "badfile.xml"
	sc := SearchClient{AccessToken: item.Token, URL: ts.URL}
	_, errSearch := sc.FindUsers(item.Search)
	if !reflect.DeepEqual(errSearch, item.Error) {
		t.Errorf("[%d] Wrong error, got `%v`, but need `%v`\n", 0, errSearch, item.Error)
	}

	fileName = "dataset.xml"
}

/*
Тесты с корректными значениями (можно еще добавить)
*/
func TestSearchPositive(t *testing.T) {
	cases := []TestCase{
		{
			Search: SearchRequest{
				Limit:      2,
				Offset:     1,
				Query:      "se",
				OrderField: "Name",
				OrderBy:    1,
			},
			Result: SearchResponse{
				Users: []User{
					{
						Id:     16,
						Name:   "Annie Osborn",
						Age:    35,
						About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
						Gender: "female",
					},
					{Id: 19, Name: "Bell Bauer", Age: 26, About: "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n", Gender: "male"},
				},
				NextPage: true,
			},
			Token: "qwe123asd",
		},
		{
			Search: SearchRequest{
				Limit:      100,
				Offset:     1,
				Query:      "Alex",
				OrderField: "Name",
				OrderBy:    1,
			},
			Result: SearchResponse{
				Users: []User{},
			},
			Token: "qwe123asd",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	for caseNum, item := range cases {
		sc := SearchClient{AccessToken: item.Token, URL: ts.URL}
		result, errSearch := sc.FindUsers(item.Search)
		if errSearch != nil {
			t.Errorf("[%d] Unexpected error `%v`\n", caseNum, errSearch)
		}
		if !reflect.DeepEqual(*result, item.Result) {
			t.Errorf("[%d] Wrong result\n got____ >>> `%v`\nexpected >>> `%v`\n", caseNum, *result, item.Result)
		}

	}
}

/*
Тесты с хэндлерами, которые в ответ подкладывают какие-то ошибки
*/
func TestBadHandlers(t *testing.T) {
	cases := []struct {
		Handler http.HandlerFunc
		Test    TestCase
	}{
		{
			Handler: HandlerGenErrUnpackErrorJSON,
			Test: TestCase{
				Token: "qwe123asd",
				Error: errorCannotUnpackErrJSON("unexpected end of JSON input"),
			},
		},
		{
			Handler: HandlerTimeOut,
			Test: TestCase{
				Token: "qwe123asd",
				Error: errorTimeout("limit=1&offset=0&order_by=0&order_field=&query="),
			},
		},
		{
			Handler: HandlerGenErrUnpackResJSON,
			Test: TestCase{
				Token: "qwe123asd",
				Error: errCannotUnpackResultJSON("json: cannot unmarshal object into Go value of type []main.User"),
			},
		},
	}

	for caseNum, item := range cases {
		ts := httptest.NewServer(http.HandlerFunc(item.Handler))
		sc := SearchClient{AccessToken: item.Test.Token, URL: ts.URL}
		_, errSearch := sc.FindUsers(item.Test.Search)
		if !reflect.DeepEqual(errSearch, item.Test.Error) {
			t.Errorf("[%d] Wrong error, got `%v`, but need `%v`\n", caseNum, errSearch, item.Test.Error)
		}

	}
}

/*
	Блок хэндлеров, которые будем подкладывать вместо SearchServer, чтобы они генерили какие-то нетипичные ошибки
*/
/* START ------------------------------------------------------------------------------------- */
func HandlerGenErrUnpackErrorJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	_, errWrite := io.WriteString(w, `{"err":123`)
	if errWrite != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func HandlerTimeOut(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
}

func HandlerGenErrUnpackResJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, errWrite := io.WriteString(w, `{"err":123}`)
	if errWrite != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestBadLink(t *testing.T) {
	item := TestCase{
		Error: errUnknownError(`Get "qwe?limit=1&offset=0&order_by=0&order_field=&query=": unsupported protocol scheme ""`),
	}
	srv := SearchClient{item.Token, "qwe"}
	_, errSearch := srv.FindUsers(item.Search)
	if !reflect.DeepEqual(errSearch, item.Error) {
		t.Errorf("[%d] Wrong error, got `%v`, but need `%v`\n", 0, errSearch, item.Error)
	}
}

/* ------------------------------------------------------------------------------------- END */
