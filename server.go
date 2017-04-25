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

const PORT = 3000
const USER_AGENT = "Mozilla/5.0 (X11; Linux x86_64; rv:5.0) Gecko/20100101 Firefox/5.0)"

func main() {
	port := flag.Int("p", PORT, "port to run service on")
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

func proxy(client *http.Client, r *http.Request) ([]byte, error) {
	query := r.URL.Query()
	var url = query.Get("url")
	if len(url) == 0 {
		return nil, errors.New(`{"error":"Missing url param."}`)
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	var method = strings.ToUpper(query.Get("method"))
	if len(method) == 0 {
		method = "GET"
	}

	var req *http.Request
	var err error
	req, err = http.NewRequest(method, url, r.Body)
	req.Header.Add("user-agent", USER_AGENT)
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
