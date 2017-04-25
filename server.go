package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
	port := flag.Int("p", 80, "port to run service on")
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Cache-Control")

		var body []byte
		var err error
		body, err = proxy(client, r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "%s", err)
		} else {
			fmt.Fprintf(w, "%s", body)
		}
	})

	log.Printf("Serving at localhost:%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

const ua = "Mozilla/5.0 (X11; Linux x86_64; rv:5.0) Gecko/20100101 Firefox/5.0)"

func proxy(client *http.Client, r *http.Request) ([]byte, error) {
	query := r.URL.Query()
	if len(query["url"]) == 0 {
		return nil, errors.New(`{"error":"Missing url param."}`)
	}

	var method string
	if len(query["method"]) > 0 {
		method = strings.ToUpper(query["method"][0])
	} else {
		method = "GET"
	}

	var req *http.Request
	var err error
	req, err = http.NewRequest(method, query["url"][0], r.Body)
	req.Header.Add("user-agent", ua)
	for _, v := range query["header"] {
		kv := strings.Split(v, "|")
		if len(kv) < 2 {
			return nil, errors.New(fmt.Sprintf(`{"error":"%s: malformed header, headers must be seperated by the string \"|\""}`, v))
		}
		if strings.ToLower(kv[0]) == "user-agent" {
			req.Header.Del("user-agent")
		}
		req.Header.Add(kv[0], kv[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(`{"error":%q}`, err))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf(`{"error":%q}`, err))
	}
	return body, nil
}
