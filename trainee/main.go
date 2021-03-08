package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {

	go func() {
		handle := http.NewServeMux()
		handle.HandleFunc("/", handleFunc)
		srv := http.Server{
			Addr:    "localhost:80",
			Handler: handle,
		}
		srv.ListenAndServe()
	}()

	handleTest := http.NewServeMux()
	handleTest.HandleFunc("/", handleTestFunc)
	srv2 := http.Server{
		Addr:    "localhost:8090",
		Handler: handleTest,
	}
	srv2.ListenAndServe()
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Hello \n")
		fmt.Fprintf(w, "URI %q \n", r.RequestURI)
		fmt.Fprintf(w, "URL %q \n", r.URL)
		fmt.Fprintf(w, "Path %q \n", r.URL.Path)
	}

}
func handleTestFunc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Test \n")
		resp, err := http.Get("http://localhost")
		if err != nil {
			fmt.Println(err)
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprintf(w, "%s", string(respBody))
	}
}
