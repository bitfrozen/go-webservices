package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type User struct {
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
	Job      string   `json:"job"`
	Email    string   `json:"email"`
	Country  string   `json:"country"`
	Company  string   `json:"company"`
	Browsers []string `json:"browsers"`
}

var userPool = sync.Pool{
	New: func() interface{} { return new(User) },
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	//r := regexp.MustCompile("@")
	var seenBrowsers []string
	uniqueBrowsers := 0
	foundUsers := ""
	lineCounter := -1

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		lineCounter++

		user := userPool.Get().(*User)
		err := json.Unmarshal(bytes, user)
		if err != nil {
			panic(err)
		}
		//users = append(users, *user)
		userPool.Put(user)

		isAndroid := false
		isMSIE := false

		browsers := user.Browsers
		for _, browser := range browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
			if strings.Contains(browser, "MSIE") {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		//email := r.ReplaceAllString(user.Email, " [at] ")
		email := strings.Replace(user.Email, "@", " [at] ", -1)
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", lineCounter, user.Name, email)
	}

	_, err = fmt.Fprintln(out, "found users:\n"+foundUsers)
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
	if err != nil {
		return
	}
}

func main() {
	FastSearch(os.Stdout)
}
