package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

type Localisation struct {
	Context    string `json:"context"`
	Original   string `json:"original"`
	Translated string `json:"translated"`
}

type LocalisationMap map[string]Localisation

func (ls LocalisationMap) UnmarshalYAML(node *yaml.Node) error {
	nodes := node.Content
	var context string
	for i, n := range nodes {
		if i%2 == 0 {
			context = n.Value
		} else {
			ls[context] = Localisation{
				Context:    context,
				Original:   n.Content[0].Value,
				Translated: n.Content[1].Value,
			}
		}
	}
	return nil
}

var y LocalisationMap

func main() {
	//yamlFile, err := ioutil.ReadFile("/Users/bobbymccann/Code/secure/conf/app/localisations/fr_FR.yaml")
	yamlFile, err := ioutil.ReadFile("/Users/bobbymccann/GolandProjects/ellevenn/example.yaml")
	if err != nil {
		panic(err)
	}
	y = LocalisationMap{}
	err = yaml.Unmarshal(yamlFile, &y)
	if err != nil {
		fmt.Println(err)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/localisation/{context}").HandlerFunc(GetLocalisation)
	router.Methods("POST").Path("/localisation").HandlerFunc(PostLocalisation)
	log.Fatal(http.ListenAndServe(":8111", router))
}

func GetLocalisation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	context := vars["context"]
	_ = json.NewEncoder(w).Encode(y[context])
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

	y[l.Context] = l

	// Send it back to the client
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(y[l.Context])
}
