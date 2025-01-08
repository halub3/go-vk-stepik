package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	ID        int    `xml:"id"`
	GUID      string `xml:"guid"`
	IsActive  bool   `xml:"isActive"`
	Balance   string `xml:"balance"`
	Pictute   string `xml:"picture"`
	Age       int    `xml:"age"`
	EyeColor  string `xml:"eyeColor"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	Company   string `xml:"company"`
	Email     string `xml:"email"`
	About     string `xml:"about"`
}

type Rows struct {
	List []Row `xml:"row"`
}

const (
	mainToken = "qwe123asd"
)

var fileName = "dataset.xml"

func readXML() (*Rows, error) {
	rows := new(Rows)
	xmlData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(xmlData, &rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func findUsers(req SearchRequest) ([]User, error) {
	rows, err := readXML()
	if err != nil {
		return []User{}, err
	}
	users := []User{}

	for _, row := range rows.List {
		name := row.FirstName + " " + row.LastName
		about := row.About

		isContain := strings.Contains(name, req.Query) || strings.Contains(about, req.Query)
		if isContain {
			user := User{Id: row.ID, Name: name, Age: row.Age, About: about, Gender: row.Gender}
			users = append(users, user)
		}
	}

	switch req.OrderBy {
	case OrderByAsc:
		sort.Slice(users, func(i, j int) bool {
			switch req.OrderField {
			case "Id":
				return users[i].Id > users[j].Id
			case "Age":
				return users[i].Age > users[j].Age
			case "Name":
				fallthrough
			default:
				return users[i].Name > users[j].Name
			}
		})
	case OrderByDesc:
		sort.Slice(users, func(i, j int) bool {
			switch req.OrderField {
			case "Id":
				return users[i].Id < users[j].Id
			case "Age":
				return users[i].Age < users[j].Age
			case "Name":
				fallthrough
			default:
				return users[i].Name < users[j].Name
			}
		})
	}

	if req.Offset > len(users) { // offset больше всего кол-ва, т.е. вышли за пределы
		return []User{}, nil
	} else if len(users) > req.Offset+req.Limit { // длина больше запрашиваемой
		return users[req.Offset : req.Limit+req.Offset], nil
	} else { // выводим все
		return users, nil
	}
}

func sendError(w http.ResponseWriter, text string, statusCode int) {
	w.WriteHeader(statusCode)
	textErr, errTextErr := json.Marshal(SearchErrorResponse{text})
	if errTextErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, errWrite := io.Writer(w).Write(textErr)
	if errWrite != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token != mainToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var errLimit, errOffset, errOrderBy error
	req := SearchRequest{}
	req.Limit, errLimit = strconv.Atoi(r.FormValue("limit"))
	req.Offset, errOffset = strconv.Atoi(r.FormValue("offset"))
	req.Query = r.FormValue("query")
	req.OrderField = r.FormValue("order_field")
	req.OrderBy, errOrderBy = strconv.Atoi(r.FormValue("order_by"))

	/*
		Ошибки преобразования к Int нужных параметров
	*/
	if errLimit != nil || errOffset != nil || errOrderBy != nil {
		sendError(w, "ErrorBadIntValues", http.StatusBadRequest)
		return
	}

	/*
	 order_field - по какому полю сортировать. Работает по полям Id, Age, Name, если пустой - то сортируем по Name,
	 				если что-то другое - SearchServer ругается ошибкой.
	*/
	if req.OrderField != "Id" && req.OrderField != "Age" && req.OrderField != "Name" && req.OrderField != "" {
		sendError(w, "ErrorBadOrderField", http.StatusBadRequest)
		return
	}

	/*
		order_by - направление сортировки (как есть, по убыванию, по возрастанию), в client.go есть соответствующие константы
	*/
	if req.OrderBy != OrderByAsc && req.OrderBy != OrderByDesc && req.OrderBy != OrderByAsIs {
		sendError(w, "ErrorBadOrderBy", http.StatusBadRequest)
		return
	}

	users, errFindUsers := findUsers(req)
	if errFindUsers != nil {
		sendError(w, "Cannot Find Users", http.StatusInternalServerError)
		return
	}
	fmt.Println("DONE")

	resJSON, errJSON := json.Marshal(users)
	if errJSON != nil {
		sendError(w, "Cannot Marshal JSON", http.StatusInternalServerError)
		return
	}

	_, errWrite := io.Writer(w).Write(resJSON)
	if errWrite != nil {
		sendError(w, "Cannot Write Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// func main() {
// 	http.HandleFunc("/", SearchServer)

// 	fmt.Println("starting server at :8080")
// 	http.ListenAndServe(":8080", nil)
// }
