package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"

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

func (ls LocalisationMap) MarshalYAML() (interface{}, error) {
	contexts := make([]string, 0, len(ls))
	for k := range ls {
		contexts = append(contexts, k)
	}
	sort.Strings(contexts)
	node := yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}
	for _, key := range contexts {
		l := ls[key]
		node.Content = append(node.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: l.Context,
		})
		node.Content = append(node.Content, &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{
					Kind:  yaml.ScalarNode,
					Value: l.Original,
				},
				{
					Kind:  yaml.ScalarNode,
					Value: l.Translated,
				},
			},
		})
	}
	return node, nil
}

var y LocalisationMap
var yamlPath string

func main() {
	yamlPath = os.Args[1]
	yamlFile, err := ioutil.ReadFile(yamlPath)
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
	l := y[context]
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(l)
}

func PostLocalisation(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 10000))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	var l Localisation
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	if err := json.Unmarshal(body, &l); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	y[l.Context] = l

	// Send it back to the client
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(y[l.Context])

	// Write the file to disk
	bbuf := new(bytes.Buffer)
	encoder := yaml.NewEncoder(bbuf)
	encoder.SetIndent(2)
	err = encoder.Encode(y)
	if err != nil {
		panic(err)
	}
	_ = ioutil.WriteFile(yamlPath, bbuf.Bytes(), 0644)
}
