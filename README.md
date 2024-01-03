[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/scan)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/scan)](https://goreportcard.com/report/github.com/wroge/scan)
![golangci-lint](https://github.com/wroge/scan/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/scan/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/scan)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/scan.svg?style=social)](https://github.com/wroge/scan/tags)

## Scan sql rows into any type powered by generics.

- Don't write the same code over and over again,
- Auto closing,
- No Reflection (faster than any reflection based mappers),
- Best practices for error handling.

### Example

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

- Scan all rows:

```go 
rows, err := db.Query("SELECT ...")
// handle error

posts, err := scan.All(rows, columns)
// handle error
```

- Scan first row:

```go 
post, err := scan.First(rows, columns)
if err != nil {
	if errors.Is(err, scan.ErrNoRows) {
		// handle no rows
	}

	// handle other error
}
```

- Scan exact one row:

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

- Scan a known number of rows:

```go 
rows, err := db.Query("SELECT ... LIMIT 10")
// handle error

posts, err := scan.Limit(10, rows, columns)
// handle error
```

- Scan rows using the underlying Iterator:

```go 
iter, err := scan.Iter(rows, columns)
// handle error

defer iter.Close()

for iter.Next() {
	err = iter.Scan(&posts[index])
	// handle error

	// Or use the Value method:
	// post, err := iter.Value()
	// handle error
}
```
