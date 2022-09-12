# Example

```go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/wroge/scan"
	_ "modernc.org/sqlite"
)

type Post struct {
	ID      int64
	Title   string
	Authors []Author
}

type Author struct {
	ID   int64
	Name string
}

func main() {
	db, _ := sql.Open("sqlite", ":memory:")

	var rows scan.Rows
	var posts []Post

	rows, _ = db.Query(`SELECT * FROM (VALUES 
		(1, null, JSON_ARRAY(JSON_OBJECT('id', 1, 'name', 'Jim'), JSON_OBJECT('id', 2, 'name', 'Tim'))),
		(2, 'Post Two', JSON_ARRAY(JSON_OBJECT('id', 2, 'name', 'Tim'))))`)

	posts, _ = scan.All[Post](rows,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.AnyErr(func(post *Post, authors []byte) error { return json.Unmarshal(authors, &post.Authors) }),
	)

	fmt.Println(posts)
	// [{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}]

	rows, _ = db.Query(`SELECT * FROM (VALUES 
		(1, null, JSON_ARRAY(1, 2), JSON_ARRAY('Jim','Tim')),
		(2, 'Post Two', JSON_ARRAY(2), JSON_ARRAY('Tim')))`)

	posts, _ = scan.All[Post](rows,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, ids []int64) {
			post.Authors = make([]Author, len(ids))

			for i, id := range ids {
				post.Authors[i].ID = id
			}
		}),
		scan.JSON(func(post *Post, names []string) {
			for i, name := range names {
				post.Authors[i].Name = name
			}
		}),
	)

	fmt.Println(posts)
	// [{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}]

	var row scan.Row
	var post Post

	row = db.QueryRow(
		`SELECT 1, 'Post One', JSON_ARRAY(JSON_OBJECT('id', 1, 'name', 'Jim'), 
			JSON_OBJECT('id', 2, 'name', 'Tim'))`)

	post, _ = scan.One[Post](row,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Any(func(post *Post, title string) { post.Title = title }),
		scan.AnyErr(func(post *Post, authors []byte) error { return json.Unmarshal(authors, &post.Authors) }),
	)

	fmt.Println(post)
	// {1 Post One [{1 Jim} {2 Tim}]}
}
```