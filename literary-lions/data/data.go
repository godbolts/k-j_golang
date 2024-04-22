package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	filePath := "./data/database.db"

    // Check if the file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        // The file does not exist, create it
        f, err := os.Create(filePath)
        if err != nil {
            log.Fatal(err)
        }
        defer f.Close()
    } 

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create user table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			username TEXT,
			password_hash TEXT,
			date_created TIMESTAMP,
			email TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Drop sessions table if exists
	_, err = db.Exec(`DROP TABLE IF EXISTS sessions`)
	if err != nil {
		log.Fatal(err)
	}

	// Create session table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			user_id INTEGER,
			start_time TIMESTAMP,
			expiry_time TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create post table if none exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			post_id INTEGER PRIMARY KEY,
			user_id TEXT,
			post_category TEXT,
			post_title TEXT,
			post_content TEXT,
			post_time TIMESTAMP,
			post_likes INTEGER,
			post_dislikes INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create comment table if none exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			comment_id INTEGER PRIMARY KEY,
			post_id INTEGER,
			username TEXT,
			comment_content TEXT,
			comment_time TIMESTAMP,
			comment_likes INTEGER,
			comment_dislikes INTEGER,
			FOREIGN KEY (post_id) REFERENCES posts(post_id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create likes/dislikes tables if none exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS post_likes (
			user_id INTEGER,
			post_id INTEGER,
			is_like BOOLEAN,
			is_dislike BOOLEAN,
			PRIMARY KEY (user_id, post_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			FOREIGN KEY (post_id) REFERENCES posts(post_id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create likes/dislikes tables if none exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comment_likes (
			user_id INTEGER,
			comment_id INTEGER,
			is_like BOOLEAN,
			is_dislike BOOLEAN,
			PRIMARY KEY (user_id, comment_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			FOREIGN KEY (comment_id) REFERENCES comments(comment_id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Create profiles table if none exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS profiles (
			profile_id INTEGER, 
			user_id INTEGER,
			about_me TEXT,
			FOREIGN KEY (user_id) REFERENCES users (user_id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}
