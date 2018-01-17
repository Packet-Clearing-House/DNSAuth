package shttp

import (
	"testing"
	"net/http"
	"io/ioutil"
)

func youpi() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("YOUPI"))
	}
}

func TestBGPServer(t *testing.T) {
	Handle("/test", youpi())
	Start(HttpServerConfig{
		Addr:     "127.0.0.1:8080",
		Acl:      []string{"127.0.0.1", "::1"},
		Password: "nope",
	})

	client := &http.Client{}

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/test", nil)
	req.SetBasicAuth("", "nope")
	if resp, err := client.Do(req); err != nil {
		t.Fatal(err)
	} else {
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			t.Fatal(err)
		} else {
			if string(body) != "YOUPI" {
				t.Fatal("Expecting YOUPI, got ", string(body), " instead!")
			}
		}
	}

	Reload(HttpServerConfig{
		Addr:     "127.0.0.1:8081",
		Acl:      []string{"127.0.0.1"},
		Password: "anothernope",
	})

	req, _ = http.NewRequest("GET", "http://127.0.0.1:8081/test", nil)
	req.SetBasicAuth("", "anothernope")
	if resp, err := client.Do(req); err != nil {
		t.Fatal(err)
	} else {
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			t.Fatal(err)
		} else {
			if string(body) != "YOUPI" {
				t.Fatal("Expecting YOUPI, got ", string(body), " instead!")
			}
		}
	}
}