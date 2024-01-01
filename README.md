# scan - sql rows into any type

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/scan)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/scan)](https://goreportcard.com/report/github.com/wroge/scan)
![golangci-lint](https://github.com/wroge/scan/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/scan/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/scan)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/scan.svg?style=social)](https://github.com/wroge/scan/tags)

- Don't write the same code over and over again,
- Auto closing,
- Mapping (Columns) in one place,
- Best practices for error handling.

## Define Columns

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
	"id": scan.Any(func(p *Post, id int64) { p.ID = id }),
	// Nullable values are scanned into pointers.
	// If pointer is nil, the default value is used.
	"title": scan.Null("default value", func(p *Post, title string) { p.Title = title }),
	// JSON values are scanned into bytes and unmarshalled into any type.
	"authors": scan.JSON(func(p *Post, authors []Author) { p.Authors = authors }),

	// Or you could create a custom scanner with this function.
	// "column": scan.Func[Post, V](func(p *Post, value V) error {
	// 	return nil
	// }),
}
```

## Scan All

```go 
rows, err := db.Query("SELECT ...")
if err != nil {
	// handle error
}

posts, err := scan.All(rows, columns)
if err != nil {
	// handle error
}
```

## Scan First

```go 
post, err := scan.First(rows, columns)
if err != nil {
	if errors.Is(err, scan.ErrNoRows) {
		// handle no rows
	}

	// handle other error
}
```

## Scan One

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

## Iterator

```go 
rows, err := db.Query("SELECT ... LIMIT 10")
if err != nil {
	// handle error
}

iter, err := scan.Iter(rows, columns)
if err != nil {
	// handle error
}

defer iter.Close()

var (
	posts = make([]Post, 10)
	index int
)

for iter.Next() {
	err = iter.Scan(&posts[index])
	if err != nil {
		// handle error
	}

	index++
}
```
