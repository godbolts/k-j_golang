package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"forum/structs"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Open the log file for writing, create if it doesn't exist, append to it if it does
	logfile, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open logfile: %v", err)
	}
	defer logfile.Close()

	// Set the log output to go to both standard output and the logfile
	log.SetOutput(io.MultiWriter(os.Stdout, logfile))

	// Print a message to both the console and the logfile
	log.Println("Server started")

	// Your existing code...
	http.HandleFunc("/users", getUsers)
	http.HandleFunc("/register", registerUser)
	http.HandleFunc("/login", authentication)
	http.HandleFunc("/posts", postList)
	http.HandleFunc("/create-post", makePost)
	http.HandleFunc("/create-comment", makeComment)
	http.HandleFunc("/edit-profile", editProfile)
	http.HandleFunc("/user", getUser)
	http.HandleFunc("/get-post", getPost)
	http.HandleFunc("/profile", getProfile)

	http.HandleFunc("/postreaction", apiReactPost)

	http.HandleFunc("/commentreaction", apiReactComment)

	server := &http.Server{Addr: ":8080"}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server started on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-signalCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract registration data from the form
	email := r.Form.Get("email")
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check if email or username already exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR username = ?", email, username).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Registration failed: email or username already in use", http.StatusConflict)
		return
	}

	// Hash and salt the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insert new user into the database
	result, err := db.Exec("INSERT INTO users (username, email, password_hash, date_created) VALUES (?, ?, ?, ?)",
		username, email, hashedPassword, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly inserted user
	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create blank profile
	_, err = db.Exec("INSERT INTO profiles (user_id, about_me) VALUES (?,?)", userID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query users from the database
	rows, err := db.Query("SELECT user_id, username, date_created, email FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare a slice to hold the results
	var users []structs.User

	// Iterate over the query results and append them to the slice
	for rows.Next() {
		var user structs.User
		if err := rows.Scan(&user.ID, &user.Username, &user.DateCreated, &user.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Convert the slice to JSON
	jsonData, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func authentication(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract login data from the form
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check password and hash
	var user structs.User
	err = db.QueryRow("SELECT user_id, username, password_hash FROM users WHERE  username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
	}

	// Generate a UUID and create a row in sessions table
	token := uuid.New().String()
	now := time.Now()
	_, err = db.Exec("INSERT INTO sessions (session_id, user_id, start_time, expiry_time) VALUES (?, ?, ?, ?)",
		token, user.ID, now.Format("2006-01-02 15:04:05"), now.Add(time.Hour).Format("2006-01-02 15:04:05"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return token to client
	response := map[string]string{"token": token}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonResponse)
}

func postList(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Retrieve search query and category filter from form data
	search := r.Form.Get("search")
	category := r.Form.Get("category")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var query string
	var args []interface{}

	if category != "" && search == "" {
		query = "SELECT * FROM posts WHERE post_category = ?"
		args = append(args, category)
	} else if category != "" && search != "" {
		query = "SELECT * FROM posts WHERE post_category = ? AND (post_title LIKE ? OR post_content LIKE ?)"
		args = append(args, category, "%"+search+"%", "%"+search+"%")
	} else if category == "" && search != "" {
		query = "SELECT * FROM posts WHERE post_title LIKE ? OR post_content LIKE ?"
		args = append(args, "%"+search+"%", "%"+search+"%")
	} else {
		query = "SELECT * FROM posts"
	}

	// Query users from the database
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare a slice to hold the results
	var posts []structs.Post

	// Iterate over the query results and append them to the slice
	for rows.Next() {
		var post structs.Post
		if err := rows.Scan(&post.ID, &post.Author, &post.Category, &post.Title, &post.Content, &post.Created, &post.Likes, &post.Dislikes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Check if no posts were found
	if len(posts) == 0 {
		http.Error(w, "No posts found", http.StatusNoContent)
		return
	}

	// Convert the slice to JSON
	jsonData, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func makePost(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session token from the cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		if err != http.ErrNoCookie {
			// Other error occurred, handle accordingly
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// Make sure to check if the cookie is not nil before accessing its value
	var token string
	if cookie != nil {
		token = cookie.Value
	}

	user := userDBreq(w, token)

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract registration data from the form
	title := r.Form.Get("title")
	category := r.Form.Get("category")
	content := r.Form.Get("content")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert new post into the database
	result, err := db.Exec("INSERT INTO posts (user_id, post_category, post_title, post_content, post_time, post_likes, post_dislikes) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.Username, category, title, content, time.Now().Format("2006-01-02 15:04:05"), 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly inserted post
	postID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the postID back in the response
	responseData := struct {
		PostID int64 `json:"postID"`
	}{
		PostID: postID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseData)
}

func makeComment(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session token from the cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		if err != http.ErrNoCookie {
			// Other error occurred, handle accordingly
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// Make sure to check if the cookie is not nil before accessing its value
	var token string
	if cookie != nil {
		token = cookie.Value
	}

	user := userDBreq(w, token)

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract data from the form
	content := r.Form.Get("content")
	postID := r.Form.Get("postID")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert new post into the database
	_, err = db.Exec("INSERT INTO comments (post_id, username, comment_content, comment_time, comment_likes, comment_dislikes) VALUES (?, ?, ?, ?, ?, ?)",
		postID, user.Username, content, time.Now().Format("2006-01-02 15:04:05"), 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session token from the cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		if err != http.ErrNoCookie {
			// Other error occurred, handle accordingly
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Token cookie not found, return unauthorized status
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Get the token value for further processing or authentication
	token := cookie.Value

	user := userDBreq(w, token)

	// Convert the user to JSON
	jsonData, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func getPost(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from URL parameters
	values := r.URL.Query()
	postID := values.Get("id")
	if postID == "" {
		http.Error(w, "Post ID not provided", http.StatusBadRequest)
		return
	}

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the database for the post with the given ID
	var post structs.Post
	err = db.QueryRow("SELECT * FROM posts WHERE post_id = ?", postID).Scan(&post.ID, &post.Author, &post.Category, &post.Title, &post.Content, &post.Created, &post.Likes, &post.Dislikes)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Query the database for the comments with the given post ID
	var comments []structs.Comment
	rows, err := db.Query("SELECT * FROM comments WHERE post_id = ?", postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	// Iterate over the query results and append them to the slice
	for rows.Next() {
		var comment structs.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Content, &comment.Created, &comment.Likes, &comment.Dislikes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	// Create a struct to hold both post and comments data
	type PostWithComments struct {
		Post     structs.Post      `json:"post"`
		Comments []structs.Comment `json:"comments"`
	}

	// Marshal the struct to JSON
	data := PostWithComments{
		Post:     post,
		Comments: comments,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	// Get the username from URL parameters
	values := r.URL.Query()
	username := values.Get("username")
	if username == "" {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the database for the user with the given username
	var user structs.User
	err = db.QueryRow("SELECT user_id, username, date_created, email FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.DateCreated, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Query the database for the profile with the given ID
	var about_me string
	err = db.QueryRow("SELECT about_me FROM profiles WHERE user_id = ?", user.ID).Scan(&about_me)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Query database for my posts
	rows, err := db.Query("SELECT * FROM posts where user_id = ?", user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var myPosts []structs.Post
	// Iterate over the query results and append them to the slice
	for rows.Next() {
		var post structs.Post
		if err := rows.Scan(&post.ID, &post.Author, &post.Category, &post.Title, &post.Content, &post.Created, &post.Likes, &post.Dislikes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		myPosts = append(myPosts, post)
	}

	// Prepare statement, query database for my liked posts
	stmt, err := db.Prepare(`
	SELECT posts.*
	FROM posts
	JOIN post_likes ON posts.post_id = post_likes.post_id
	WHERE post_likes.user_id = ?
	AND post_likes.is_like = 1
`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	rows, err = stmt.Query(user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var myLikedPosts []structs.Post
	for rows.Next() {
		var post structs.Post
		if err := rows.Scan(&post.ID, &post.Author, &post.Category, &post.Title, &post.Content, &post.Created, &post.Likes, &post.Dislikes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		myLikedPosts = append(myLikedPosts, post)
	}

	// Create a struct to hold all profile data
	type Profile struct {
		User         structs.User   `json:"post"`
		About_Me     string         `json:"about_me"`
		MyPosts      []structs.Post `json:"my_posts"`
		MyLikedPosts []structs.Post `json:"my_liked_posts"`
	}

	// Marshal the struct to JSON
	data := Profile{
		User:         user,
		About_Me:     about_me,
		MyPosts:      myPosts,
		MyLikedPosts: myLikedPosts,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func editProfile(w http.ResponseWriter, r *http.Request) {
	// Get user profile from query params
	queryParams := r.URL.Query()
	username := queryParams.Get("username")

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract profile data from the form
	about_me := r.Form.Get("about_me")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Find user_id from username
	var user structs.User
	err = db.QueryRow("SELECT user_id FROM users WHERE username = ?", username).Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert new info into the database
	_, err = db.Exec(`UPDATE profiles SET about_me = ? WHERE user_id = ?;`, about_me, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func userDBreq(w http.ResponseWriter, token string) structs.User {
	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return structs.User{}
	}
	defer db.Close()

	// Prepare the SQL statement
	stmt, err := db.Prepare(`
	SELECT users.*
	FROM sessions
	JOIN users ON sessions.user_id = users.user_id
	WHERE sessions.session_id = ?
`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return structs.User{}
	}
	defer stmt.Close()

	// Execute the query with the session token
	row := stmt.QueryRow(token)

	// Scan the result into user struct
	var user structs.User
	err = row.Scan(&user.ID, &user.Username, &user.DateCreated, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			// No active session with the given token, return unauthorized status
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return structs.User{}
		}
		// Other error occurred, handle accordingly
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return structs.User{}
	}
	return user
}

func apiReactPost(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form data: %v", err)
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Retrieve parameters from the request
	userID := r.Form.Get("username")
	postID := r.Form.Get("post")
	reactionType := r.Form.Get("type")

	var like int
	var dislike int

	if reactionType == "like" {
		like = 1
		dislike = 0
	} else {
		like = 0
		dislike = 1
	}

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check if a row exists for the given user_id and post_id
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&count)
	if err != nil {
		log.Printf("Failed to query database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			log.Printf("Rolling back transaction due to error: %v", err)
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}()

	if count > 0 {
		// If a row exists, update the existing row
		_, err = tx.Exec("UPDATE post_likes SET is_like = ? , is_dislike = ? WHERE user_id = ? AND post_id = ?", like, dislike, userID, postID)
		if err != nil {
			log.Printf("Failed to update post reaction: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {

		// If no row exists, insert a new row
		_, err = tx.Exec("INSERT INTO post_likes (user_id, post_id, is_like, is_dislike) VALUES (?, ?, ?, ?)", userID, postID, like, dislike)
		if err != nil {
			log.Printf("Failed to insert post reaction: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = tx.Commit()
	tx, err = db.Begin()

	// Update the posts table with the updated counts of likes and dislikes
	_, err = tx.Exec("UPDATE posts SET post_likes = (SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND is_like = 1), post_dislikes = (SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND is_dislike = 1) WHERE post_id = ?", postID, postID, postID)
	if err != nil {
		log.Printf("Failed to update posts table: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reaction updated successfully"))
}

func apiReactComment(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form data: %v", err)
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Retrieve parameters from the request
	userID := r.Form.Get("username")
	postID := r.Form.Get("comment")
	reactionType := r.Form.Get("type")

	log.Printf("Received parameters: userID=%s, commentID=%s, reactionType=%s", userID, postID, reactionType)

	var like int
	var dislike int

	if reactionType == "like" {
		like = 1
		dislike = 0
	} else {
		like = 0
		dislike = 1
	}

	log.Printf("Determined like=%d, dislike=%d", like, dislike)

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	log.Println("Database connection opened")

	// Check if a row exists for the given user_id and post_id
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, postID).Scan(&count)
	if err != nil {
		log.Printf("Failed to query database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Existing count for user_id=%s, comment_id=%s: %d", userID, postID, count)

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			log.Printf("Rolling back transaction due to error: %v", err)
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}()

	if count > 0 {
		// If a row exists, update the existing row
		_, err = tx.Exec("UPDATE comment_likes SET is_like = ? , is_dislike = ? WHERE user_id = ? AND comment_id = ?", like, dislike, userID, postID)
		if err != nil {
			log.Printf("Failed to update post reaction: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Updated comment_likes for user_id=%s, comment_id=%s", userID, postID)
	} else {
		// If no row exists, insert a new row
		_, err = tx.Exec("INSERT INTO comment_likes (user_id, comment_id, is_like, is_dislike) VALUES (?, ?, ?, ?)", userID, postID, like, dislike)
		if err != nil {
			log.Printf("Failed to insert post reaction: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Inserted new row into comment_likes for user_id=%s, comment_id=%s", userID, postID)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Transaction committed successfully")

	// Start a new transaction
	tx, err = db.Begin()
	if err != nil {
		log.Printf("Failed to start new transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the comments table with the updated counts of likes and dislikes
	_, err = tx.Exec("UPDATE comments SET comment_likes = (SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = 1), comment_dislikes = (SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_dislike = 1) WHERE comment_id = ?", postID, postID, postID)
	if err != nil {
		log.Printf("Failed to update comments table: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Updated comments table for comment_id=%s", postID)

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Transaction committed successfully")

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reaction updated successfully"))
}
