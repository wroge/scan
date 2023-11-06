# scan - sql rows into any type

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/scan)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/scan)](https://goreportcard.com/report/github.com/wroge/scan)
![golangci-lint](https://github.com/wroge/scan/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/scan/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/scan)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/scan.svg?style=social)](https://github.com/wroge/scan/tags)

- Don't write the same code over and over again.
- Define the mapping (Columns) in one place.

## Examples

- ```All[T](Rows, ...Column[T])``` scans rows into ```[]T``` and performs the closing.
- ```AnyErr``` is a type-safe Column with a pointer to ```V``` and a setter function ```func(*T, V)```.
- ```Any``` is like ```AnyErr```, but without the returned error.
- ```Null``` scans nullable values and uses a default if the scanned value is nil.

```go
type Author struct {
	ID   int64
	Name string
}

type Post struct {
	ID      int64
	Title   string
	Authors []Author
}

rows, _ = db.Query(`SELECT * FROM (VALUES 
	(1, null, JSON_ARRAY(JSON_OBJECT('id', 1, 'name', 'Jim'), JSON_OBJECT('id', 2, 'name', 'Tim'))),
	(2, 'Post Two', JSON_ARRAY(JSON_OBJECT('id', 2, 'name', 'Tim'))))`)

columns := []scan.Column[Post]{
	scan.Any(func(post *Post, id int64) { post.ID = id }),
	scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
	scan.AnyErr(func(post *Post, authors []byte) error { return json.Unmarshal(authors, &post.Authors) }),
}

posts, _ = scan.All(rows, columns...)
// [{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}]
```

- Column ```JSON``` scans a value into ```[]byte``` and unmarshals it into ```V```.
- The setter functions are executed in order, so the following is possible.

```go
rows, _ = db.Query(`SELECT * FROM (VALUES 
	(1, null, JSON_ARRAY(1, 2), JSON_ARRAY('Jim','Tim')),
	(2, 'Post Two', JSON_ARRAY(2), JSON_ARRAY('Tim')))`)

columns = []scan.Column[Post]{
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
}

posts, _ = scan.All(rows, columns...)
// [{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}]
```

- A single ```Row``` can be scanned via ```One[T](Row, ...Column[T])```.
- Full example [here](https://github.com/wroge/scan/blob/main/EXAMPLE.md).

```go
row = db.QueryRow(
	`SELECT 1, 'Post One', JSON_ARRAY(JSON_OBJECT('id', 1, 'name', 'Jim'), 
		JSON_OBJECT('id', 2, 'name', 'Tim'))`)

columns = []scan.Column[Post]{
	scan.Any(func(post *Post, id int64) { post.ID = id }),
	scan.Any(func(post *Post, title string) { post.Title = title }),
	scan.AnyErr(func(post *Post, authors []byte) error { return json.Unmarshal(authors, &post.Authors) }),
}

post, _ = scan.One(row, columns...)
// {1 Post One [{1 Jim} {2 Tim}]}
```

## Each

- With ```Each[T](ctx, func(ctx, T) error, Rows, ...Column[T])``` it is possible to scan large number of rows.

## Benchmarks

- ```Standard``` scans rows by hand.
- Criticism and possible improvements to the [benchmarks and tests](https://github.com/wroge/scan/blob/main/scan_test.go) are very welcome.

```sh
➜ go test -bench=. -benchtime=10s -benchmem 
goos: darwin
goarch: amd64
pkg: github.com/wroge/scan
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz
BenchmarkExample1WrogeScanAll-8           635017             18906 ns/op            6936 B/op        137 allocs/op
BenchmarkExample1WrogeScanEach-8          564099             21634 ns/op            8590 B/op        142 allocs/op
BenchmarkExample1Standard-8               714564             16244 ns/op            5576 B/op        112 allocs/op
BenchmarkExample2WrogeScanAll-8           689020             16916 ns/op            8928 B/op        198 allocs/op
BenchmarkExample2WrogeScanEach-8          590502             20411 ns/op            8471 B/op        142 allocs/op
BenchmarkExample2Standard-8               922117             13368 ns/op            9640 B/op        117 allocs/op
BenchmarkExample3WrogeScanOne-8          5183390              2435 ns/op             816 B/op         20 allocs/op
BenchmarkExample3Standard-8              6530119              1922 ns/op             480 B/op         12 allocs/op
PASS
ok      github.com/wroge/scan   105.128s
```
