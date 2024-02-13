package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Input struct {
	userInput string
	encoding  bool
}

var temp *template.Template

func HandleFunc(w http.ResponseWriter, r *http.Request) {
	temp = template.Must(template.ParseFiles("design.html"))
	w.WriteHeader(http.StatusOK)
	temp.Execute(w, nil)
}

func Decoder(w http.ResponseWriter, r *http.Request) {
	log.Println("POST Input Data...")
	var cmd *exec.Cmd

	inputStructure := Input{
		userInput: r.FormValue("inputText"),
		encoding:  r.Form.Has("Encode"),
	}
	cwd, err := os.Getwd()

	if inputStructure.encoding { // This checks if the user wants to encode or decode, the lines are different because the command function does workj with an empty variable.
		cmd = exec.Command("go", "run", filepath.Join(cwd, "coder/main.go"), "-encode", inputStructure.userInput)
	} else {
		cmd = exec.Command("go", "run", filepath.Join(cwd, "coder/main.go"), inputStructure.userInput)
	}
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	art := string(output) // Output is written in bytes so it has to be converted into strings.
	if err != nil || strings.Contains(art, "Error") {
		w.WriteHeader(http.StatusBadRequest)
		temp.Execute(w, err)
	} else {
		log.Println("Get data for printing...")
		w.WriteHeader(http.StatusAccepted)
		temp.Execute(w, art)
	}
}

func main() {
	address := "localhost:4444"
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(".")))) //Made a static path so that the server could read the css file from this folder and not from a link.
	mux.HandleFunc("/", HandleFunc)
	mux.HandleFunc("/decoder", Decoder)

	theServer := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	log.Printf("Starting server on %s", address)
	log.Fatal(theServer.ListenAndServe())
}
