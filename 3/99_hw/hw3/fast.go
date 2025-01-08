package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	// "log"
	"hw3/model"
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

		user := &model.User{}
		err = user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		browsers := user.Browsers
		// fmt.Println(browsers)

		for _, browser := range browsers {
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
		email := strings.ReplaceAll(user.Email, "@", " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", idx, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
