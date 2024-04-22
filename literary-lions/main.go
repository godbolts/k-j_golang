package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"forum/structs"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Define a struct to hold the template data
type TemplateData struct {
	User       structs.User      // User data
	Posts      []structs.Post    // Posts data (if needed)
	Post       structs.Post      // Single post read
	Message    string            // Message to display
	Categories [4]string         // Categories for posts and filtering
	CurrentURL string            // URL tracking for redirect
	Comments   []structs.Comment // Comments data
}

// Our posting categories for sorting
var categories = [4]string{"Review", "Suggestion", "Author Chat", "Life, the Universe and Everything"}

// Our server address
const address = "0.0.0.0:5555"
const redirectAddress = "localhost:5555"

func main() {
	// Run data.go
	cmdData := exec.Command("go", "run", "data/data.go")
	if err := cmdData.Run(); err != nil {
		log.Fatalf("Failed to run data.go: %v", err)
	}

	// Run api.go
	cmdAPI := exec.Command("go", "run", "api/api.go")
	if err := cmdAPI.Start(); err != nil {
		log.Fatalf("Failed to run api.go: %v", err)
	}

	// Wait for processes to start
	time.Sleep(time.Second * 5)

	mux := http.NewServeMux()

	// Serve static files (CSS, JS, etc.) from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/create-post", postHandler)
	mux.HandleFunc("/create-comment", commentHandler)
	mux.HandleFunc("/read", readpostHandler)
	mux.HandleFunc("/profile", profileHandler)
	mux.HandleFunc("/edit-profile", editprofileHandler)
	mux.HandleFunc("/postreaction", postReaction)
	mux.HandleFunc("/commentreaction", commentReaction)
	mux.HandleFunc("/404", notFoundHandler)
	// Handle paths with trailing slashes
	mux.Handle("/{any}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path[:len(r.URL.Path)-1], http.StatusMovedPermanently)
	}))

	// Handle all other routes
	mux.HandleFunc("/{any}", notFoundHandler)

	server := &http.Server{
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
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-sigCh
	log.Println("\nReceived interrupt signal. Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown error:", err)
	}

	log.Println("Server gracefully stopped")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	http.ServeFile(w, r, "static/404.html")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	message := queryParams.Get("message")

	// Strip query parameters from the request URL
	r.URL.RawQuery = ""

	// check for logged in user
	user := checkUser(w, r)

	// get all posts
	resp, err := apiRequester(w, r, "posts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var posts []structs.Post

	if resp.StatusCode != http.StatusNoContent {
		// Read the JSON data from the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response body", http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(body, &posts)
		if err != nil {
			http.Error(w, "Failed to decode posts data", http.StatusInternalServerError)
			return
		}
	} else {
		message = "No posts found with those criteria"
	}

	// Parse HTML template
	tmpl, err := template.ParseFiles("static/home.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare the data structure to pass to the template
	data := TemplateData{
		User:       user,
		Posts:      posts,
		Message:    message, // Pass the message to the template
		Categories: categories,
	}

	// Execute template with posts data and write to response
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send the request to the API
	resp, err := apiRequester(w, r, "register")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode != http.StatusConflict {
			http.Error(w, fmt.Sprintf("Failed to register user - Status Code: %d", resp.StatusCode), resp.StatusCode)
			return
		}
		message := "Failed to register - email or username taken"
		// Encode the message in URL parameters
		query := url.Values{}
		query.Add("message", message)
		// Redirect the user to the desired URL with the message in parameters
		http.Redirect(w, r, "/?"+query.Encode(), http.StatusFound)
		return
	}

	message := "User registered successfully!"

	// Encode the message in URL parameters
	query := url.Values{}
	query.Add("message", message)

	// Redirect the user to the desired URL with the message in parameters
	http.Redirect(w, r, "/?"+query.Encode(), http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send the request to the API
	resp, err := apiRequester(w, r, "login")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusAccepted {
		message := "Failed to login - incorrect username or password"
		// Encode the message in URL parameters
		query := url.Values{}
		query.Add("message", message)
		// Redirect the user to the desired URL with the message in parameters
		http.Redirect(w, r, "/?"+query.Encode(), http.StatusFound)
		return
	}

	// Parse the response body
	var responseData map[string]string
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		http.Error(w, "Failed to parse response body", http.StatusInternalServerError)
		return
	}

	token, ok := responseData["token"] // Retrieve the token from the map
	if !ok {
		http.Error(w, "Token not found in response", http.StatusInternalServerError)
		return
	}

	// Set cookie with the token
	expiration := time.Now().Add(1 * time.Hour) // Set cookie expiration time
	cookie := http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expiration,
		Path:    "/",
	}

	http.SetCookie(w, &cookie)

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Redirect back to same page
	redirect := r.Form.Get("redirect_url")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the token cookie by setting its expiration to a past time
	expiration := time.Now().Add(-1 * time.Hour) // Set expiration time to the past
	cookie := http.Cookie{
		Name:    "token",
		Value:   "", // Set cookie value to empty
		Expires: expiration,
		Path:    "/",
	}

	http.SetCookie(w, &cookie)

	// Redirect the user to the homepage or a specific URL after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send the request to the API
	resp, err := apiRequester(w, r, "create-post")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusCreated {
		http.Error(w, fmt.Sprintf("Failed to post - Status Code: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// Decode the response body to extract the postID
	var responseData struct {
		PostID int64 `json:"postID"`
	}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		http.Error(w, "Failed to decode response body", http.StatusInternalServerError)
		return
	}

	// Access the postID
	postID := responseData.PostID

	http.Redirect(w, r, "/read?id="+strconv.Itoa(int(postID)), http.StatusSeeOther)
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send the request to the API
	resp, err := apiRequester(w, r, "create-comment")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusCreated {
		http.Error(w, fmt.Sprintf("Failed to comment - Status Code: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Redirect back to same page
	redirect := r.Form.Get("redirect_url")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

func readpostHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	message := queryParams.Get("message")
	postID := queryParams.Get("id")

	// check for logged in user
	user := checkUser(w, r)

	// Fetch the post data based on the post ID
	// Send the request to the API
	resp, err := apiRequester(w, r, "/get-post")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Failed to fetch post - Status Code: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// Create a struct to hold both post and comments data
	type PostWithComments struct {
		Post     structs.Post      `json:"post"`
		Comments []structs.Comment `json:"comments"`
	}
	var postData PostWithComments

	// Read the JSON data from the response body
	err = json.NewDecoder(resp.Body).Decode(&postData)
	if err != nil {
		http.Error(w, "Failed to decode post data", http.StatusInternalServerError)
		return
	}

	// Prepare the data structure to pass to the template
	data := TemplateData{
		User:       user,
		Post:       postData.Post,
		Message:    message, // Pass the message to the template
		Categories: categories,
		CurrentURL: "http://" + redirectAddress + "/read?id=" + postID,
		Comments:   postData.Comments,
	}

	// Parse HTML template
	tmpl, err := template.ParseFiles("static/readpost.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute template with post data and write to response
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// check for logged in user to view profiles
	user := checkUser(w, r)
	if user.Username == "" {
		message := "You must be logged in to view user profiles"
		// Encode the message in URL parameters
		query := url.Values{}
		query.Add("message", message)
		// Redirect the user to the desired URL with the message in parameters
		http.Redirect(w, r, "/?"+query.Encode(), http.StatusSeeOther)
	}

	// Fetch the profile data based on the username
	resp, err := apiRequester(w, r, "/profile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Failed to fetch user - Status Code: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// Create a struct to hold all profile data
	type Profile struct {
		User         structs.User   `json:"post"`
		About_Me     string         `json:"about_me"`
		MyPosts      []structs.Post `json:"my_posts"`
		MyLikedPosts []structs.Post `json:"my_liked_posts"`
	}
	var profile Profile

	// Read the JSON data from the response body
	err = json.NewDecoder(resp.Body).Decode(&profile)
	if err != nil {
		http.Error(w, "Failed to decode post data", http.StatusInternalServerError)
		return
	}

	// Check if user is viewing their own profile for editability
	var itsaMe bool
	if user.ID == profile.User.ID {
		itsaMe = true
	}

	// Create a struct to pass out all profile data
	type ProfileData struct {
		User         structs.User
		About_Me     string
		ItsaMe       bool
		CurrentURL   string
		MyPosts      []structs.Post
		MyLikedPosts []structs.Post
	}
	profileData := ProfileData{
		User:         profile.User,
		About_Me:     profile.About_Me,
		ItsaMe:       itsaMe,
		CurrentURL:   "http://" + redirectAddress + "/profile?username=" + profile.User.Username,
		MyPosts:      profile.MyPosts,
		MyLikedPosts: profile.MyLikedPosts,
	}

	// Parse HTML template
	tmpl, err := template.ParseFiles("static/profile.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute template with post data and write to response
	if err := tmpl.Execute(w, profileData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func editprofileHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send the request to the API
	resp, err := apiRequester(w, r, "edit-profile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusCreated {
		http.Error(w, fmt.Sprintf("Failed to edit profile - Status Code: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Redirect back to same page
	redirect := r.Form.Get("redirect_url")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func checkUser(w http.ResponseWriter, r *http.Request) structs.User {
	userResp, err := apiRequester(w, r, "user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return structs.User{}
	}
	defer userResp.Body.Close()

	var user structs.User
	if userResp.StatusCode != http.StatusUnauthorized {
		// Read the JSON data from the response body
		err := json.NewDecoder(userResp.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to decode user data", http.StatusInternalServerError)
			return structs.User{}
		}
	}
	return user
}

func postReaction(w http.ResponseWriter, r *http.Request) {
	apiRequester(w, r, "postreaction")
}

func commentReaction(w http.ResponseWriter, r *http.Request) {
	apiRequester(w, r, "commentreaction")
}

func apiRequester(w http.ResponseWriter, r *http.Request, address string) (*http.Response, error) {
	// Get the query parameters from the client request
	queryParams := r.URL.Query()

	// Retrieve the session token from the cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		if err != http.ErrNoCookie {
			// Other error occurred, handle accordingly
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}
	}
	// Make sure to check if the cookie is not nil before accessing its value
	var token string
	if cookie != nil {
		token = cookie.Value
	}

	// Parse form data from the request
	err = r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse form data: %v", err)
	}

	// Encode form data
	requestBody := bytes.NewBufferString(r.Form.Encode())

	// Create a new request with the encoded form data
	apiURL := "http://localhost:8080/" + address
	// Append the query parameters to the API URL
	if queryParams != nil {
		apiURL += "?"
		for key, values := range queryParams {
			for _, value := range values {
				apiURL += key + "=" + value + "&"
			}
		}
		// Remove the trailing "&" if it exists
		apiURL = strings.TrimSuffix(apiURL, "&")
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to database API: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Attach the session token as a cookie to the request
	if token != "" {
		req.AddCookie(cookie)
	}

	// Send the request to the API using a shared HTTP client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to database API: %v", err)
	}

	return resp, nil
}
