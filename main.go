package main

import (
	"encoding/json"
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
