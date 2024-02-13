package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Data represents a slice of map[string]interface{}, for data from APIs
type Data []map[string]interface{}

// API URLs
const baseURL = "http://localhost:3000/api"
const imgURL = "http://localhost:3000/api/images/"

// Struct for passing data to the HTML template
type ViewData struct {
	APIHost          string
	ModelInfo        Data
	ManufacturerInfo Data
	CategoryInfo     Data
}

type compData struct { // A struct for comparing models, it gets fed into it the maps of two spesific models
	APIHost string
	ModelA  map[string]interface{}
	ModelB  map[string]interface{}
}

type Model struct { // A struct for giving a html template information to use in the reccomender
	ID    string
	Name  string
	Image string
	Count int
}

// Handler for car display table page, show all or search
func HandleTable(w http.ResponseWriter, r *http.Request) {
	//get search info from request
	query := r.URL.Query().Get("q")
	searchField := r.URL.Query().Get("field")

	var modelData Data
	var manufacturerData Data
	var categoryData Data

	// get all data from APIs
	err := fetchData("/models", &modelData)
	if err != nil {
		http.Error(w, "Failed to fetch model data", http.StatusInternalServerError)
		log.Printf("Error fetching model data: %v", err)
		return
	}
	err = fetchData("/manufacturers", &manufacturerData)
	if err != nil {
		http.Error(w, "Failed to fetch manufacturer data", http.StatusInternalServerError)
		log.Printf("Error fetching manufacturer data: %v", err)
		return
	}
	err = fetchData("/categories", &categoryData)
	if err != nil {
		http.Error(w, "Failed to fetch category data", http.StatusInternalServerError)
		log.Printf("Error fetching category data: %v", err)
		return
	}

	// parse template
	tmpl, err := template.ParseFiles("design.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}

	// check for search info, filter results, return to home with blank search
	if query != "" {
		switch searchField {
		case "model":
			modelData = search(query, modelData, modelData, "id")
		case "manufacturer":
			modelData = search(query, modelData, manufacturerData, "manufacturerId")
		case "category":
			modelData = search(query, modelData, categoryData, "categoryId")
		}
	} else if searchField != "" {
		HandleIntro(w, r)
		return
	}

	// build total data struct to pass to HTML template
	urlData := ViewData{
		APIHost:          imgURL,
		ModelInfo:        modelData,
		ManufacturerInfo: manufacturerData,
		CategoryInfo:     categoryData,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, urlData)
}

// Handler for welcome page
func HandleIntro(w http.ResponseWriter, r *http.Request) {
	// Fetch the top 3 car models the user has interacted with
	topCars, err := fetchTopCarsForUser() // This fetches the top three models from the reccomender.csv file and if all is 0 then it gives an empty return
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error fetching top cars for user: %v", err)
		return
	}

	tmpl, err := template.New("intro.html").ParseFiles("intro.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, topCars)
}

func fetchTopCarsForUser() ([]Model, error) {
	// Open the file
	file, err := os.Open("recommender.csv")
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Read the file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse the records into a slice of Models
	var models []Model
	for _, record := range records[1:] { // Skip the header row
		count, err := strconv.Atoi(record[3])
		if err != nil {
			return nil, fmt.Errorf("error converting string to int: %v", err)
		}
		models = append(models, Model{
			ID:    record[0],
			Name:  record[1],
			Image: imgURL + record[2],
			Count: count,
		})
	}

	// Sort the models by count in descending order
	sort.Slice(models, func(i, j int) bool {
		return models[i].Count > models[j].Count
	})

	// If all counts are 0, return two empty maps
	if models[0].Count == 0 {
		return []Model{}, nil
	}

	// Return the top 3 models
	if len(models) > 3 {
		models = models[:3]
	}
	return models, nil
}

// Handler for manufacturer info display page
func HandleManf(w http.ResponseWriter, r *http.Request) {
	var ManufacturerData Data // Creates an empty variable into which the manufacturer data can be read into

	err := fetchData("/manufacturers", &ManufacturerData)
	if err != nil {
		http.Error(w, "Failed to fetch manufacturer data", http.StatusInternalServerError)
		log.Printf("Error fetching manufacturer data: %v", err)
		return
	}

	// Read country codes from CSV file into a map
	countryCodes := make(map[string]string)
	file, err := os.Open("country-codes.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		countryCodes[record[0]] = strings.ToLower(record[1]) // Assuming country name is in the first column and country code is in the second within the country-codes.csv file
	}

	// Add countryCode field to each item in ManufacturerData so that countrycodes can be used to create an image link
	for i, item := range ManufacturerData {
		if code, ok := countryCodes[item["country"].(string)]; ok {
			ManufacturerData[i]["countryCode"] = code
		}
	}

	tmpl, err := template.New("manf.html").ParseFiles("manf.html") // Creates a html template
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, ManufacturerData); err != nil { // If there is no error then the data is read into the template and can be used in the html
		log.Printf("Error executing template: %v", err)
	}
}

func HandleCompare(w http.ResponseWriter, r *http.Request) {
	compModel := Data{} // Initialize compModel as an empty slice of map[string]interface{}

	err := r.ParseForm()
	if err != nil {
		// handle error
		fmt.Println("Error parsing form: ", err)
		return
	}
	modelIds := r.Form["modelIds"]

	incrementCount(modelIds[0]) // Increase the first models popularity count in the reccomender.csv
	incrementCount(modelIds[1]) // Increase the second models popularity count in the reccomender.csv

	// Fetch and load the first model
	var modelOne map[string]interface{}
	err = fetchData("/models/"+modelIds[0], &modelOne)
	if err != nil {
		// handle error
		fmt.Println("Error fetching data for model 1: ", err)
		return
	}
	compModel = append(compModel, modelOne) // Append the fetched model to compModel

	// Fetch and load the second model
	var modelTwo map[string]interface{}
	err = fetchData("/models/"+modelIds[1], &modelTwo)
	if err != nil {
		// handle error
		fmt.Println("Error fetching data for model 2: ", err)
		return
	}
	compModel = append(compModel, modelTwo) // Append the fetched model to compModel

	tmpl, err := template.ParseFiles("comp.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}

	compData := compData{
		APIHost: imgURL,
		ModelA:  compModel[0], // Use the first model in compModel
		ModelB:  compModel[1], // Use the second model in compModel
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, compData)
}

func HandleIncrementCount(w http.ResponseWriter, r *http.Request) {

	// Parse the request body
	var data struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// Increment the count for the specified model in the recommender.csv file
	incrementCount(data.ID)
}

func incrementCount(id string) {
	// Open the file
	file, err := os.OpenFile("recommender.csv", os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	// Read the file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Find the row with the specified ID and increment the count
	for i, record := range records {
		if record[0] == id {
			count, err := strconv.Atoi(record[3])
			if err != nil {
				log.Fatalf("Error converting string to int: %v", err)
			}
			count++
			records[i][3] = strconv.Itoa(count)
			break
		}
	}

	// Write the updated data back to the file
	file.Seek(0, 0)
	file.Truncate(0)
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}
	writer.Flush()
	file.Close()
}

// Retrieves data from API
func fetchData(endpoint string, target interface{}) error {
	url := fmt.Sprintf("%s%s", baseURL, endpoint) // Creates a combined string which is the location of the data
	response, err := http.Get(url)                // Checks if the data is there and if it gets a response from the server
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status code: %d", response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(target) // Uses the json package to decode information form the endpoint
	if err != nil {
		return err
	}
	return nil
}

// Search function for filtering results. Handles any field
func search(query string, models Data, list Data, toSearch string) Data {
	var searchResults Data
	var matches []interface{}

	// check list for matches on query, e.g. manufacturers, store in slice
	for _, item := range list {
		if strings.Contains(strings.ToLower(item["name"].(string)), strings.ToLower(query)) {
			matches = append(matches, item["id"])
		}
	}
	// check appropriate field in models for matches in slice, rebuild results
	for _, model := range models {
		for _, item := range matches {
			if item == model[toSearch] {
				searchResults = append(searchResults, model)
			}
		}
	}
	return searchResults
}

// Runs API server, and our page server in a separate channel
func main() {
	cwd, _ := os.Getwd()                  // Gets the working directory of the device so the server can be run on any machine
	apiPath := filepath.Join(cwd, "/api") // Gets the directory of the api inside the working directory

	err := os.Chdir(apiPath) // Changes the directory to the api one in order to write commands to it
	if err != nil {
		fmt.Println("Error changing directory to 'api':", err)
		return
	}
	defer os.Chdir(cwd)

	// Install the necessary packages for the api
	cmd := exec.Command("npm", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error installing packages:", err)
		return
	}

	cmd = exec.Command("make", "run") // Create a command to make the api run
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start() // Writes the command to the api to start
	if err != nil {
		fmt.Println("Did not find the API.")
		return
	}

	time.Sleep(time.Second * 5)

	os.Chdir(cwd) // Returns the working directory to this server

	// Check if the file reccomender.csv file exists and moves on if it does not, otherwise it creates a new one with counts set to 0
	if _, err := os.Stat("recommender.csv"); os.IsNotExist(err) {
		var compData Data
		// Fetch the data
		err := fetchData("/models", &compData)
		if err != nil {
			log.Fatalf("Error fetching data: %v", err)
		}

		// Create the file
		file, err := os.Create("recommender.csv")
		if err != nil {
			log.Fatalf("Error creating file: %v", err)
		}

		// Create a CSV writer
		writer := csv.NewWriter(file)

		// Write the header
		if err := writer.Write([]string{"id", "name", "image", "count"}); err != nil {
			log.Fatalf("Error writing header to csv: %v", err)
		}

		for _, model := range compData {
			id := fmt.Sprintf("%v", model["id"])
			name := fmt.Sprintf("%v", model["name"])
			image := fmt.Sprintf("%v", model["image"])
			count := "0"
			if err := writer.Write([]string{id, name, image, count}); err != nil {
				log.Fatalf("Error writing record to csv: %v", err)
			}
		}
		writer.Flush()
		file.Close()
	}

	address := "localhost:4444" // Creating a mux server on the local 4444 port with multible pages and multible different functions
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/", HandleIntro)                             // The introduction page
	mux.HandleFunc("/list", HandleTable)                         // The main table page
	mux.HandleFunc("/list/compare", HandleCompare)               // the comparrison page
	mux.HandleFunc("/manufacturer", HandleManf)                  // The manufacturers page
	mux.HandleFunc("/list/incrementCount", HandleIncrementCount) // A page used only to feed reccomender counts back into the server

	theServer := &http.Server{
		Addr:           address,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on %s", address)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := theServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-sigCh
	log.Println("\nReceived interrupt signal. Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := theServer.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown error:", err)
	}

	log.Println("Server gracefully stopped")
}
