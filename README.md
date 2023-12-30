# scan - sql rows into any type

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/scan)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/scan)](https://goreportcard.com/report/github.com/wroge/scan)
![golangci-lint](https://github.com/wroge/scan/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/scan/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/scan)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/scan.svg?style=social)](https://github.com/wroge/scan/tags)

- Don't write the same code over and over again.
- Define the mapping (Columns) in one place.

## Example

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

var columns = scan.Columns[Post]{
	// Any value supported by your database driver can be used.
	"id":      scan.Any(func(p *Post, id int64) { p.ID = id }),
	// Nullable data is scanned into a pointer (*string).
	// If pointer is nil, the default value is used.
	"title":   scan.Null("default value", func(p *Post, title string) { p.Title = title }),
	// JSON data is scanned into bytes and unmarshalled into []Author.
	"authors": scan.JSON(func(p *Post, authors []Author) { p.Authors = authors }),
}

rows, err := db.Query("SELECT ...")
// handle error

posts, err := scan.All(rows, columns)
// handle error
```