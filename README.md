[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/scan)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/scan)](https://goreportcard.com/report/github.com/wroge/scan)
![golangci-lint](https://github.com/wroge/scan/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/scan/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/scan)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/scan.svg?style=social)](https://github.com/wroge/scan/tags)

# Scan

The scan package offers a powerful and efficient way to scan SQL rows into any Go type, leveraging the power of generics. This package emphasizes simplicity, performance, and best practices in error handling.

## Features

- **Efficient and Reusable**: Avoid repetitive code with generalized scanning functions.
- **Automatic Resource Management**: Automatic closing of iterators to prevent resource leaks.
- **Non-Reflective Operations**: Offers faster performance compared to reflection-based mappers.
- **Robust Error Handling**: Adheres to best practices for managing and reporting errors.

## Usage

```go
import "github.com/wroge/scan"

type Author struct {
	ID   int64
	Name string
}

type Post struct {
	ID      int64
	Title   string
	Authors []Author
}

// Define mapping of database columns to struct fields.
var columns = scan.Columns[Post]{
	// Map the 'id' column to the 'ID' field in the 'Post' struct.
	// Uses the 'scan.Any' function for direct assignment without additional processing.
	"id": scan.Any(func(p *Post, id int64) { p.ID = id }),

	// Map the 'title' column to the 'Title' field in the 'Post' struct.
	// The 'scan.Null' function allows handling of nullable database columns.
	// If the 'title' column is null, 'default title' is used as the value.
	"title": scan.Null("default title", func(p *Post, title string) { p.Title = title }),

	// Map the 'authors' column, expected to be in JSON format, to the 'Authors' field in the 'Post' struct.
	// The 'scan.JSON' function automatically handles unmarshalling of the JSON data into the 'Author' struct slice.
	"authors": scan.JSON(func(p *Post, authors []Author) { p.Authors = authors }),

	// Or you could create a custom scanner with this function.
	// "column": scan.Func[Post, V](func(p *Post, value V) error {
	// 	return nil
	// }),
}

rows, err := db.Query("SELECT ...")
// handle error
```

#### Scanning all rows

```go 
posts, err := scan.All(rows, columns)
// handle error
```

#### Scanning the first row

```go 
post, err := scan.First(rows, columns)
if err != nil {
	if errors.Is(err, scan.ErrNoRows) {
		// handle no rows
	}

	// handle other error
}
```

#### Scanning exactly one row

```go 
post, err := scan.One(rows, columns)
if err != nil {
	if errors.Is(err, scan.ErrTooManyRows) {
		// handle too many rows
	}
	if errors.Is(err, scan.ErrNoRows) {
		// handle no rows
	}

	// handle other error
}
```

#### Scanning a limited number of rows

```go 
posts, err := scan.Limit(10, rows, columns)
if err != nil {
	if errors.Is(err, scan.ErrTooManyRows) {
		// ignore if result set has more than 10 rows
	}

	// handle other error
}
```

#### Using the Iterator Directly

```go 
iter, err := scan.Iter(rows, columns)
// handle error

defer iter.Close()

for iter.Next() {
	var post Post

	err = iter.Scan(&post)
	// handle error

	// Or use the Value method:
	post, err := iter.Value()
	// handle error
}
```
