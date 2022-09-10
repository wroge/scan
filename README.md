# scan - sql rows into any type

- Don't write the same code over and over again.
- Any rows implementation is supported (*sql.Rows, pgx.Rows, ...).
- Auto closing.
- No reflection, only generics.
- Aggregation of rows is not a goal of this module and should be performed in the database.
- [Benchmarks](#benchmarks).

```go
package main

import (
	"database/sql"
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
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)

	fmt.Println(posts)
	// [{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}]

	rows, _ = db.Query(`SELECT * FROM (VALUES 
		(1, null, json_array(1, 2), json_array('Jim','Tim')),
		(2, 'Post Two', json_array(2), json_array('Tim')))`)

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

	row := db.QueryRow(
		`SELECT 1, 'Post One', JSON_ARRAY(JSON_OBJECT('id', 1, 'name', 'Jim'), 
			JSON_OBJECT('id', 2, 'name', 'Tim'))`)

	post, err := scan.One[Post](row,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)

	fmt.Println(post, err)
	// {1 Post One [{1 Jim} {2 Tim}]}
}
```

## Benchmarks

- ```Standard``` scans rows by hand.
- Criticism and possible improvements to the benchmarks are very welcome.

```sh
➜ go test -bench=. -benchtime=10s -benchmem
goos: darwin
goarch: amd64
pkg: github.com/wroge/scan
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkExample1WrogeScan-8      686976             17547 ns/op            7176 B/op        149 allocs/op
BenchmarkExample1Standard-8       754894             14519 ns/op            5528 B/op        112 allocs/op
BenchmarkExample2WrogeScan-8      759638             15721 ns/op            8936 B/op        201 allocs/op
BenchmarkExample2Standard-8       810042             12435 ns/op            9748 B/op        117 allocs/op
BenchmarkExample3WrogeScan-8     5651865              2190 ns/op             848 B/op         24 allocs/op
BenchmarkExample3Standard-8      7641949              1571 ns/op             432 B/op         12 allocs/op
PASS
ok      github.com/wroge/scan   94.167s
```