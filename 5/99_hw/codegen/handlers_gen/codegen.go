package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type FieldInfo struct {
	Name       string
	ParamName  string
	Type       string
	IsRequired bool
	IsEnum     bool
	EnumErr    string
	EnumSrch   string
	IsDefault  bool
	Default    string
	IsMin      bool
	Min        int
	IsMax      bool
	Max        int
}

type StructInfo struct {
	Name   string
	Fields []*FieldInfo
}

type FuncInfo struct {
	Name      string
	ApiInput  *ApiInput
	FieldName string
	RecvName  string
}

type ApiInput struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type tmp struct {
	Name string
}

var (
	apiGen      = "// apigen:api "
	contextType = "&{context Context}"
)

func CreateFunc(f FuncInfo, out *os.File, structs map[string]StructInfo) {
	TplFuncInitTmpl.Execute(out, f)
	if f.ApiInput.Method != "" {
		TplFuncCheckMethod.Execute(out, f)
	}
	if f.ApiInput.Auth {
		TplFuncCheckAuth.Execute(out, nil)
	}

	TplFuncInitVars.Execute(out, f)

	structArg := structs[f.FieldName]

	for _, field := range structArg.Fields {
		TplFuncReadField.Execute(out, field)
	}
	TplFuncCall.Execute(out, f)
	TplFuncSendRes.Execute(out, nil)
	TplFuncEnd.Execute(out, nil)
}

func CreateServeHTTP(funcMap map[string][]FuncInfo, out *os.File) {
	for key, funcs := range funcMap {
		TplSHTTPInit.Execute(out, tmp{key})
		for _, f := range funcs {
			TplSHTTPCase.Execute(out, f)
		}
		TplSHTTPEnd.Execute(out, nil)
	}
}

func ParseStructs(node *ast.File, out *os.File) map[string]StructInfo {
	structsMap := make(map[string]StructInfo)

DECL_LOOP:
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			// fmt.Printf("SKIP %#T is not *ast.GenDecl\n", f)
			continue
		}

		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				// fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				// fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}

			// fmt.Printf("Work with Struct %s\n", currType.Name.Name)

			newStruct := &StructInfo{
				Name:   currType.Name.Name,
				Fields: make([]*FieldInfo, 0, len(currStruct.Fields.List)),
			}
		FIELDS_LOOP:
			for _, field := range currStruct.Fields.List {
				_, ok := field.Type.(*ast.Ident)
				if !ok {
					/* Тип поля структуры не ast.Ident, значит структура эта нам не нужна */
					continue DECL_LOOP
				}

				if field.Tag != nil {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					if tag.Get("apivalidator") == "" {
						/* идем к следующему полю, это пропускаем */
						continue FIELDS_LOOP
					}

					fieldInfo := &FieldInfo{
						Name: field.Names[0].Name,
						Type: field.Type.(*ast.Ident).Name,
					}

					tagVal := tag.Get("apivalidator")
					tagParams := strings.Split(tagVal, ",")

				PARAM_LOOP:
					for _, param := range tagParams {
						if param == "required" {
							/* этот параметр распарсили, у него нет "=" */
							fieldInfo.IsRequired = true
							continue PARAM_LOOP
						}
						markName, markVal := strings.Split(param, "=")[0], strings.Split(param, "=")[1]
						switch markName {
						case "paramname":
							fieldInfo.ParamName = markVal
						case "default":
							fieldInfo.IsDefault = true
							fieldInfo.Default = markVal
						case "min":
							fieldInfo.IsMin = true
							fieldInfo.Min, _ = strconv.Atoi(markVal)
						case "max":
							fieldInfo.IsMax = true
							fieldInfo.Max, _ = strconv.Atoi(markVal)
						case "enum":
							fieldInfo.IsEnum = true
							enum := strings.Split(markVal, "|")
							fieldInfo.EnumErr = strings.Join(enum, ", ")
							srchList := []string{}
							for _, str := range enum {
								srchList = append(srchList, fmt.Sprintf("\"%s\"", str))
							}
							fieldInfo.EnumSrch = strings.Join(srchList, ", ")

						}
					}
					if fieldInfo.ParamName == "" {
						fieldInfo.ParamName = strings.ToLower(fieldInfo.Name)
					}
					newStruct.Fields = append(newStruct.Fields, fieldInfo)
				}

			}

			/* Пишем в мапу только если хоть к каким-то полям применяемся */
			if len(newStruct.Fields) > 0 {
				structsMap[newStruct.Name] = *newStruct
			}
		}
	}

	return structsMap
}

func ParseFuncs(node *ast.File, out *os.File, structs map[string]StructInfo) map[string][]FuncInfo {
	funcsMap := make(map[string][]FuncInfo)

DECL_LOOP:
	for _, d := range node.Decls {
		f, ok := d.(*ast.FuncDecl)
		if !ok {
			// fmt.Printf("SKIP %#T is not *ast.GenDecl\n", f)
			continue
		}
		/* Нет комментария */
		if f.Doc == nil {
			continue DECL_LOOP
		}
		/* Не метод */
		if f.Recv == nil {
			continue DECL_LOOP
		}
		/* Список получателей !=1, т.е. метод не от одного объекта */
		if len(f.Recv.List) != 1 {
			continue DECL_LOOP
		}
		/* Список аргументов только из 2х */
		if len(f.Type.Params.List) != 2 {
			continue DECL_LOOP
		}

		isApply := false
		apiParams := &ApiInput{}

	COMMENT_LOOP:
		for _, comment := range f.Doc.List {
			if !strings.HasPrefix(comment.Text, apiGen) {
				continue COMMENT_LOOP
			}
			isApply = true
			apiParamsStr := strings.Replace(comment.Text, apiGen, "", 1)
			errJson := json.Unmarshal([]byte(apiParamsStr), apiParams)
			if errJson != nil {
				log.Fatalln(errJson)
			}
			break COMMENT_LOOP
		}
		/* Не нашли нужный коммент в функций */
		if !isApply {
			continue DECL_LOOP
		}

		newFunc := &FuncInfo{
			Name:     f.Name.Name,
			ApiInput: apiParams,
		}

		field := f.Type.Params.List[0]
		/* первый параметр должен быть типа contextType */
		if fmt.Sprint(field.Type) != contextType {
			continue DECL_LOOP
		}
		field = f.Type.Params.List[1]
		_, ok = field.Type.(*ast.Ident)
		/* Второй параметр должен быть структурой */
		if !ok {
			continue DECL_LOOP
		}
		newFunc.FieldName = field.Type.(*ast.Ident).Name

		field = f.Recv.List[0]
		typeSearch := field.Type.(*ast.StarExpr).X.(*ast.Ident).Name
		newFunc.RecvName = typeSearch

		funcRecv, isExist := funcsMap[typeSearch]
		if isExist {
			funcRecv = append(funcRecv, *newFunc)
			funcsMap[typeSearch] = funcRecv
		} else {
			funcRecv = []FuncInfo{*newFunc}
			funcsMap[typeSearch] = funcRecv
		}

		CreateFunc(*newFunc, out, structs)

	}
	return funcsMap
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	TplInit.Execute(out, "")

	structsMap := ParseStructs(node, out)
	funcMap := ParseFuncs(node, out, structsMap)
	CreateServeHTTP(funcMap, out)

}
