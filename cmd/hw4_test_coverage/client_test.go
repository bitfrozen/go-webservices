package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const Token = "a9ab16fc-4d04-4863-820a-4519a30c1ed4"

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

func parseUserFile(path string) ([]UserBackend, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	u := new(Users)
	err = xml.Unmarshal(data, &u)
	if err != nil {
		return nil, err
	}

	for i := range u.UserList {
		u.UserList[i].Name = u.UserList[i].FirstName + " " + u.UserList[i].LastName
	}

	return u.UserList, nil
}

func userListToJSON(users []UserBackend) ([]byte, error) {
	size := len(users)
	result := make([]UserBackend, 0, size)

	for _, u := range users {
		result = append(result, u)
	}
	resultJson, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJson, nil
}

func searchUserFile(params *SearchRequest) ([]UserBackend, error) {
	users, err := parseUserFile("dataset.xml")
	if err != nil {
		return nil, err
	}
	switch params.OrderField {
	case "Id":
		switch params.OrderBy {
		case -1:
			sort.SliceStable(users, func(i, j int) bool {
				return users[i].Id > users[j].Id
			})
		case 1:
			sort.SliceStable(users, func(i, j int) bool {
				return users[i].Id < users[j].Id
			})
		}
	case "Age":
		switch params.OrderBy {
		case -1:
			sort.SliceStable(users, func(i, j int) bool {
				return users[i].Age > users[j].Age
			})
		case 1:
			sort.SliceStable(users, func(i, j int) bool {
				return users[i].Age < users[j].Age
			})
		}
	case "Name":
		switch params.OrderBy {
		case -1:
			sort.SliceStable(users, func(i, j int) bool {
				return users[i].Name > users[j].Name
			})
		case 1:
			sort.SliceStable(users, func(i, j int) bool {
				return users[i].Name < users[j].Name
			})
		}
	}

	var result []UserBackend

	if params.Query == "" {
		size := min(len(users[params.Offset:]), params.Limit)
		result = make([]UserBackend, size)
		copy(result, users[params.Offset:])
	} else {
		for _, u := range users[params.Offset:] {
			if strings.Contains(u.Name, params.Query) || strings.Contains(u.About, params.Query) {
				result = append(result, u)
				if len(result) >= params.Limit {
					break
				}
			}
		}
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func serverBadRequest(w http.ResponseWriter, info string) {
	w.WriteHeader(http.StatusBadRequest)
	e := SearchErrorResponse{Error: info}
	jr, err := json.Marshal(e)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(jr)
	if err != nil {
		log.Fatal(err)
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// check token
	if r.Header.Get("AccessToken") != Token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// get limit
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		serverBadRequest(w, err.Error())
		return
	}

	// get offset
	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		serverBadRequest(w, err.Error())
		return
	}

	// get query
	query := r.FormValue("query")

	// get order field
	orderField := r.FormValue("order_field")
	if orderField != "" && orderField != "Id" && orderField != "Age" && orderField != "Name" {
		serverBadRequest(w, "ErrorBadOrderField")
		return
	}
	if orderField == "" {
		orderField = "Name"
	}

	// get order by
	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		serverBadRequest(w, err.Error())
		return
	}

	params := &SearchRequest{
		Limit:      limit,
		Offset:     offset,
		Query:      query,
		OrderField: orderField,
		OrderBy:    orderBy,
	}

	result, err := searchUserFile(params)
	if err != nil {
		log.Fatal(err)
	}

	resultJson, err := userListToJSON(result)
	if err != nil {
		log.Fatal(err)
	}

	_, err = w.Write(resultJson)
	if err != nil {
		log.Fatal(err)
	}
}

func TestFindUsersBadAuthTokenResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{AccessToken: "Bad token", URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || err.Error() != "Bad AccessToken" {
		t.Error("Failed check to bad token response")
	}
}

func TestFindUsersInternalServerErrorResponse(t *testing.T) {
	fk := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	ts := httptest.NewServer(http.HandlerFunc(fk))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || err.Error() != "SearchServer fatal error" {
		t.Error("Failed check for internal server error response")
	}
}

func TestFindUsersBadRequestWithBrokenErrorJsonResponse(t *testing.T) {
	fk := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"Error": 400`)
		if err != nil {
			log.Fatal(err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(fk))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || !strings.Contains(err.Error(), "cant unpack error json:") {
		t.Error("Failed check for broken json in bad request error")
	}
}

func TestFindUsersBadRequestWithOrderFieldResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{OrderField: "About"})
	if sr != nil || err.Error() != "OrderFeld About invalid" {
		t.Error("Failed check for bad order field response")
	}
}

func TestFindUsersDefaultBadRequestResponse(t *testing.T) {
	fk := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"Error": "400"}`)
		if err != nil {
			log.Fatal(err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(fk))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || !strings.Contains(err.Error(), "unknown bad request error:") {
		t.Error("Failed check for default bad request response")
	}
}

func TestFindUsersNegativeLimit(t *testing.T) {
	sc := &SearchClient{}
	sr, err := sc.FindUsers(SearchRequest{Limit: -1})
	if sr != nil || err.Error() != "limit must be > 0" {
		t.Error("Failed check for negative limit")
	}
}

func TestFindUsersTooBigLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{Limit: 30})
	if len(sr.Users) > 25 || err != nil {
		t.Error("Failed check for too high limit")
	}
}

func TestFindUsersNegativeOffset(t *testing.T) {
	sc := &SearchClient{}
	sr, err := sc.FindUsers(SearchRequest{Offset: -1})
	if sr != nil || err.Error() != "offset must be > 0" {
		t.Error("Failed check for negative offset")
	}
}

func TestFindUsersAllRequestFieldsExist(t *testing.T) {
	fk := func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Error(err)
		}
		if _, ok := r.Form["limit"]; !ok {
			t.Error("Request parameter `limit` missing")
		}
		if _, ok := r.Form["offset"]; !ok {
			t.Error("Request parameter `offset` missing")
		}
		if _, ok := r.Form["query"]; !ok {
			t.Error("Request parameter `query` missing")
		}
		if _, ok := r.Form["order_field"]; !ok {
			t.Error("Request parameter `order_field` missing")
		}
		if _, ok := r.Form["order_by"]; !ok {
			t.Error("Request parameter `order_by` missing")
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(fk))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	_, _ = sc.FindUsers(SearchRequest{})
}

func TestFindUsersTimeoutErrorResponse(t *testing.T) {
	fk := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
		w.WriteHeader(http.StatusInternalServerError)
	}
	ts := httptest.NewServer(http.HandlerFunc(fk))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || !strings.Contains(err.Error(), "timeout for") {
		t.Error("Failed check for timeout response")
	}
}

func TestFindUsersUnknownErrorResponse(t *testing.T) {
	sc := &SearchClient{AccessToken: Token, URL: ""}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || !strings.Contains(err.Error(), "unknown error") {
		t.Error("Failed check for unknown error response")
	}
}

func TestFindUserBadJsonEncoding(t *testing.T) {
	fk := func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"Error": 400}`)
		if err != nil {
			log.Fatal(err)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(fk))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{})
	if sr != nil || !strings.Contains(err.Error(), "cant unpack result json:") {
		t.Error("Failed check for bad json encoding")
	}
}

func TestFindUserOnePageResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{Query: "Rebekah Sutton", Limit: 5})
	if err != nil || sr.NextPage {
		t.Error("Failed single-page result check")
	}
}

func TestFindUserMultiPageResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{AccessToken: Token, URL: ts.URL}
	sr, err := sc.FindUsers(SearchRequest{Limit: 5})
	if err != nil || !sr.NextPage {
		t.Error("Failed multi-page result check")
	}
}
