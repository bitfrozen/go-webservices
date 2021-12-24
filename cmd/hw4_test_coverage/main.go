package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type UserBackend struct {
	Id        int    `xml:"id" json:"Id"`
	Name      string `xml:"-" json:"Name"`
	FirstName string `xml:"first_name" json:"-"`
	LastName  string `xml:"last_name" json:"-"`
	Age       int    `xml:"age" json:"Age"`
	About     string `xml:"about" json:"About"`
	Gender    string `xml:"gender" json:"Gender"`
}

type Users struct {
	Version  string        `xml:"version,attr"`
	UserList []UserBackend `xml:"row"`
}

func parseUserFile(path string) []UserBackend {
	//result := make([]UserBackend, 0, 0)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	u := new(Users)
	err = xml.Unmarshal(data, &u)
	if err != nil {
		panic(err)
	}

	for i := range u.UserList {
		u.UserList[i].Name = u.UserList[i].FirstName + " " + u.UserList[i].LastName
	}

	return u.UserList
}

//func SearchServer(w http.ResponseWriter, r *http.Request) {
//
//}

func main() {
	users := parseUserFile("D:/developer/go-webservices/cmd/hw4_test_coverage/dataset.xml")
	for _, u := range users {
		j, err := json.Marshal(u)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(j))
	}
}
