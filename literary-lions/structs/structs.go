package structs

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DateCreated string `json:"date_created"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type Post struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Author   string `json:"author"`
	Created  string `json:"created"`
	Likes    uint   `json:"likes"`
	Dislikes uint   `json:"dislikes"`
}

type Comment struct {
	ID       string `json:"id"`
	PostID   string `json:"post_id"`
	Content  string `json:"content"`
	Author   string `json:"author"`
	Created  string `json:"created"`
	Likes    uint   `json:"likes"`
	Dislikes uint   `json:"dislikes"`
}
