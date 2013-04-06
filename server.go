package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

var keystore = make(map[string]interface{})
var keyIdRegexp = regexp.MustCompile("^/key/([^/]+)$")

func writeError(err error, w http.ResponseWriter) {
	var errOut = make(map[string]string, 0)
	errOut["error"] = err.Error()
	out, err := json.Marshal(&errOut)
	if err != nil {
		log.Printf("error writing response: %s", err.Error())
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(out)
}

func writeMessage(msg string, w http.ResponseWriter) {
	var msgOut = make(map[string]string, 0)
	msgOut["success"] = "true"
	if msg != "" {
		msgOut["message"] = msg
	}

	out, err := json.Marshal(&msgOut)
	if err != nil {
		log.Printf("error writing response: %s", err.Error())
		return
	}
	w.Write(out)
}

func getKey(w http.ResponseWriter, r *http.Request) {
	id := keyIdRegexp.ReplaceAllString(r.URL.Path, "$1")
	if id == "" {
		w.Write([]byte("{}"))
		return
	}
	log.Println("look up", id)
	value := keystore[id]
	result := make(map[string]interface{})
	result[id] = value
	out, err := json.Marshal(&result)
	if err != nil {
		writeError(err, w)
		return
	} else {
		w.Write(out)
	}
}

func handleKey(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		postKey(w, r)
	} else if r.Method == "PUT" {
		putKey(w, r)
	} else {
		getKeys(w, r)
	}
}

func postKey(w http.ResponseWriter, r *http.Request) {
	setKey(w, r, false)
}

func putKey(w http.ResponseWriter, r *http.Request) {
	setKey(w, r, true)
}

func setKey(w http.ResponseWriter, r *http.Request, noUpdate bool) {
	req := r.FormValue("data")
	if req == "" {
		writeError(fmt.Errorf("no key specified"), w)
		return
	}
	kv := make(map[string]interface{}, 0)
	err := json.Unmarshal([]byte(req), &kv)
	if err != nil {
		writeError(err, w)
		return
	}

	result := make(map[string]interface{})
	result["success"] = true

	for k, v := range kv {
		if _, ok := keystore[k]; ok && noUpdate {
			if _, ok := result["not_updated"]; !ok {
				result["not_updated"] = make([]string, 0)
			}
		} else if v != nil {
			keystore[k] = v
			if _, ok := result["set"]; !ok {
				result["set"] = make([]string, 0)
			}
			result["set"] = append(result["set"].([]string), k)
		} else {
			delete(keystore, k)
			if _, ok := result["removed"]; !ok {
				result["set"] = make([]string, 0)
			}
			result["set"] = append(result["removed"].([]string), k)
		}
	}
	out, err := json.Marshal(&result)
	if err != nil {
		writeError(err, w)
	}
	w.Write(out)
}

func getKeys(w http.ResponseWriter, r *http.Request) {
	keys := make([]string, 0)
	for k, _ := range keystore {
		keys = append(keys, k)
	}
	out, err := json.Marshal(keys)
	if err != nil {
		writeError(err, w)
		return
	}
	w.Write(out)
}

func root(w http.ResponseWriter, r *http.Request) {
	index := `drunken_dangerzone!

Simple in-memory key-value store powered by Go. The following endpoints
are supported:

  * GET /key     - retrieve a list of all keys stored in the server
  * PUT /key     - json-encode a list of keys and their values, and they
                   will be stored as long as they aren't already present
                   in the keystore.
  * POST /key    - similar to PUT, but will overwrite any existing values.
  * GET /key/:id - retrieve the value associated with the id.
`
	w.Write([]byte(index))
}

func main() {
	port := flag.String("p", "8080", "port to listen on")

	http.HandleFunc("/key/", getKey)
	http.HandleFunc("/key", handleKey)
	http.HandleFunc("/", root)
	log.Printf("drunken dangerzone is starting on http://127.0.0.1:%s", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}
