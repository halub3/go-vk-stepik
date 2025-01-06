package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	// "log"
)

var (
	ANDROID = "Android"
	MSIE    = "MSIE"
)

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	rd := bufio.NewReader(file)

	seenBrowsers := []string{}
	foundUsers := ""

	// for i, line := range lines {
	for idx := 0; ; idx++ {
		// fmt.Printf("%v %v\n", err, line)
		line, _, err := rd.ReadLine()
		if err == io.EOF {
			break
		}

		user := make(map[string]interface{})
		err = json.Unmarshal(line, &user)
		// user := &u.User{}
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}
			okAndroid, okMSIE := strings.Contains(browser, ANDROID), strings.Contains(browser, MSIE)
			// fmt.Println(isAndroid, isMSIE, browser)
			if okAndroid || okMSIE {
				if okAndroid {
					isAndroid = true
				}
				if okMSIE {
					isMSIE = true
				}
				// fmt.Println(isAndroid, isMSIE, browser)
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.ReplaceAll(user["email"].(string), "@", " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", idx, user["name"], email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
