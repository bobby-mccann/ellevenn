package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Localisation struct {
	Context    string `json:"context"`
	Original   string `json:"original"`
	Translated string `json:"translated"`
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/localisation/{context}", GetLocalisation)
	log.Fatal(http.ListenAndServe(":8111", router))
}

func GetLocalisation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	context := vars["context"]
	json.NewEncoder(w).Encode(Localisation{
		Context:    context,
		Original:   "this is the original text",
		Translated: "this is the translated text",
	})
}

func PostLocalisation(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 10000))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var l Localisation
	if err := json.Unmarshal(body, &l); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	// TODO: write our localisation

	// Send it back to the client
	json.NewEncoder(w).Encode(l)
}
