# scan - sql rows into any type

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/scan)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/scan)](https://goreportcard.com/report/github.com/wroge/scan)
![golangci-lint](https://github.com/wroge/scan/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/scan/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/scan)
[![tippin.me](https://badgen.net/badge/%E2%9A%A1%EF%B8%8Ftippin.me/@_wroge/F0918E)](https://tippin.me/@_wroge)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/scan.svg?style=social)](https://github.com/wroge/scan/tags)

- Don't write the same code over and over again.
- Any rows implementation is supported (*sql.Rows, pgx.Rows, ...).
- Auto closing.
- No reflection, only generics.
- Aggregation of rows is not a goal of this module and should be performed in the database.
- Take a look at [wroge/superbasic](https://github.com/wroge/superbasic) for query building.
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
- Criticism and possible improvements to the [benchmarks and tests](https://github.com/wroge/scan/blob/main/scan_test.go) are very welcome.

```sh
➜ go test -bench=. -benchtime=10s -benchmem
goos: darwin
goarch: amd64
pkg: github.com/wroge/scan
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkExample1WrogeScan-8      691414             17149 ns/op            7224 B/op        149 allocs/op
BenchmarkExample1Standard-8       778964             14490 ns/op            5576 B/op        112 allocs/op
BenchmarkExample2WrogeScan-8      755938             16308 ns/op            8984 B/op        201 allocs/op
BenchmarkExample2Standard-8       805694             13108 ns/op            9823 B/op        117 allocs/op
BenchmarkExample3WrogeScan-8     5683458              2158 ns/op             896 B/op         24 allocs/op
BenchmarkExample3Standard-8      7509338              1592 ns/op             480 B/op         12 allocs/op
PASS
ok      github.com/wroge/scan   95.894s
```