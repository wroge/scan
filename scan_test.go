//nolint:gocritic,depguard,gochecknoglobals,exhaustivestruct,exhaustruct
package scan_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/wroge/scan"
)

type Author struct {
	ID   int64
	Name string
}

type Post struct {
	ID      int64
	Title   string
	Authors []Author
}

func rows1() *fakeRows {
	return &fakeRows{
		index:   -1,
		columns: []string{"id", "title", "authors"},
		data: [][]any{
			{1, nil, []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
			{2, "Post Two", []byte(`[{"id": 2, "name": "Tim"}]`)},
			{3, "Post Three", []byte(`[{"id": 2, "name": "Tim"},{"id": 3, "name": "Tom"}]`)},
			{4, "Post Four", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
			{5, "Post Five", []byte(`[{"id": 1, "name": "Jim"},{"id": 3, "name": "Tom"}]`)},
			{6, "Post Six", []byte(`[{"id": 2, "name": "Tim"}]`)},
			{7, "Post Seven", []byte(`[{"id": 3, "name": "Tom"}]`)},
			{8, "Post Eight", []byte(`[{"id": 1, "name": "Jim"}]`)},
			{9, "Post Nine", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"},{"id": 3, "name": "Tom"}]`)},
			{10, "Post Ten", []byte(`[{"id": 3, "name": "Tom"}]`)},
		},
	}
}

func rows2() *fakeRows {
	return &fakeRows{
		index:   -1,
		columns: []string{"id", "title", "authors"},
		data: [][]any{
			{1, nil, []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}
}

//nolint:goerr113
func rows3() *fakeRows {
	return &fakeRows{
		index:      -1,
		columnsErr: fmt.Errorf("columns error"),
		columns:    []string{"id", "title", "authors"},
		data:       [][]any{},
	}
}

//nolint:goerr113
func rows4() *fakeRows {
	return &fakeRows{
		index:   -1,
		scanErr: fmt.Errorf("scan error"),
		columns: []string{"id", "title", "authors"},
		data: [][]any{
			{1, nil, []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}
}

func rows5() *fakeRows {
	return &fakeRows{
		index:   -1,
		columns: []string{"id", "title", "authors"},
		data:    [][]any{},
	}
}

func rows6() *fakeRows {
	return &fakeRows{
		index:   -1,
		columns: []string{"id", "title", "authors"},
		data: [][]any{
			{1, nil, []byte(`[{"id": 1, name: "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}
}

var columns1 = scan.Columns[Post]{
	"id":      scan.Any(func(post *Post, id int64) { post.ID = id }),
	"title":   scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
	"authors": scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
}

var columns2 = scan.Columns[Post]{
	"id":      scan.Any(func(post *Post, id int64) { post.ID = id }),
	"authors": scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
}

func TestAll(t *testing.T) {
	t.Parallel()

	posts, err := scan.All[Post](rows1(), columns1)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(posts) != `[{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}`+
		` {3 Post Three [{2 Tim} {3 Tom}]} {4 Post Four [{1 Jim} {2 Tim}]}`+
		` {5 Post Five [{1 Jim} {3 Tom}]} {6 Post Six [{2 Tim}]} {7 Post Seven [{3 Tom}]}`+
		` {8 Post Eight [{1 Jim}]} {9 Post Nine [{1 Jim} {2 Tim} {3 Tom}]} {10 Post Ten [{3 Tom}]}]` {
		t.Fatal(posts)
	}
}

func TestLimit(t *testing.T) {
	t.Parallel()

	posts, err := scan.Limit[Post](10, rows1(), columns1)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(posts) != `[{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}`+
		` {3 Post Three [{2 Tim} {3 Tom}]} {4 Post Four [{1 Jim} {2 Tim}]}`+
		` {5 Post Five [{1 Jim} {3 Tom}]} {6 Post Six [{2 Tim}]} {7 Post Seven [{3 Tom}]}`+
		` {8 Post Eight [{1 Jim}]} {9 Post Nine [{1 Jim} {2 Tim} {3 Tom}]} {10 Post Ten [{3 Tom}]}]` {
		t.Fatal(posts)
	}
}

func TestLimit2(t *testing.T) {
	t.Parallel()

	posts, err := scan.Limit[Post](12, rows1(), columns1)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(posts) != `[{1 No Title [{1 Jim} {2 Tim}]} {2 Post Two [{2 Tim}]}`+
		` {3 Post Three [{2 Tim} {3 Tom}]} {4 Post Four [{1 Jim} {2 Tim}]}`+
		` {5 Post Five [{1 Jim} {3 Tom}]} {6 Post Six [{2 Tim}]} {7 Post Seven [{3 Tom}]}`+
		` {8 Post Eight [{1 Jim}]} {9 Post Nine [{1 Jim} {2 Tim} {3 Tom}]} {10 Post Ten [{3 Tom}]}]` {
		t.Fatal(posts, len(posts))
	}
}

func TestLimit3(t *testing.T) {
	t.Parallel()

	posts, err := scan.Limit[Post](1, rows1(), columns1)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(posts) != `[{1 No Title [{1 Jim} {2 Tim}]}]` {
		t.Fatal(posts, len(posts))
	}
}

func TestFirst(t *testing.T) {
	t.Parallel()

	post, err := scan.First[Post](rows1(), columns1)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(post) != `{1 No Title [{1 Jim} {2 Tim}]}` {
		t.Fatal(post)
	}
}

func TestFirst2(t *testing.T) {
	t.Parallel()

	_, err := scan.First[Post](rows5(), columns1)
	if !errors.Is(err, scan.ErrNoRows) {
		t.Fatal(err)
	}
}

func TestOne(t *testing.T) {
	t.Parallel()

	post, err := scan.One[Post](rows2(), columns2)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(post) != `{1  [{1 Jim} {2 Tim}]}` {
		t.Fatal(post)
	}
}

func TestOneError(t *testing.T) {
	t.Parallel()

	_, err := scan.One[Post](rows1(), columns1)
	if !errors.Is(err, scan.ErrTooManyRows) {
		t.Fatal(err)
	}
}

func TestOneError2(t *testing.T) {
	t.Parallel()

	_, err := scan.One[Post](rows5(), columns1)
	if !errors.Is(err, scan.ErrNoRows) {
		t.Fatal(err)
	}
}

func TestOneError3(t *testing.T) {
	t.Parallel()

	_, err := scan.One[Post](rows4(), columns1)
	if err == nil {
		t.Fatal(err)
	}
}

func TestAllError(t *testing.T) {
	t.Parallel()

	_, err := scan.All[Post](rows3(), columns1)
	if err == nil {
		t.Fatal("error is nil")
	}
}

func TestLimitError(t *testing.T) {
	t.Parallel()

	_, err := scan.Limit[Post](10, rows3(), columns1)
	if err == nil {
		t.Fatal("error is nil")
	}
}

func TestFirstError(t *testing.T) {
	t.Parallel()

	_, err := scan.First[Post](rows3(), columns1)
	if err == nil {
		t.Fatal("error is nil")
	}
}

func TestAllError2(t *testing.T) {
	t.Parallel()

	_, err := scan.All[Post](rows4(), columns1)
	if err == nil {
		t.Fatal("error is nil")
	}
}

func TestFirstError2(t *testing.T) {
	t.Parallel()

	_, err := scan.First[Post](rows4(), columns1)
	if err == nil {
		t.Fatal("error is nil")
	}
}

func TestFirstError3(t *testing.T) {
	t.Parallel()

	post, err := scan.First[Post](rows6(), columns1)
	if err == nil {
		t.Fatal("error is nil", post)
	}
}

func TestIterator(t *testing.T) {
	t.Parallel()

	iter, err := scan.Iter[Post](rows1(), columns1)
	if err != nil {
		t.Fatal(err)
	}

	for iter.Next() {
		_, err = iter.Value()
		if err != nil {
			t.Fatal(err)
		}
	}

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}
}

type fakeRows struct {
	columnsErr error
	closeErr   error
	scanErr    error
	err        error
	index      int
	columns    []string
	data       [][]any
}

func (r *fakeRows) Columns() ([]string, error) {
	return r.columns, r.columnsErr
}

func (r *fakeRows) Close() error {
	return r.closeErr
}

func (r *fakeRows) Err() error {
	return r.err
}

func (r *fakeRows) Next() bool {
	r.index++

	return r.index < len(r.data)
}

//nolint:funlen,cyclop
func (r *fakeRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}

	for index, d := range dest {
		switch value := d.(type) {
		case *[]byte:
			switch s := r.data[r.index][index].(type) {
			case []byte:
				*value = s
			}
		case *string:
			switch s := r.data[r.index][index].(type) {
			case string:
				*value = s
			}
		case *int64:
			switch s := r.data[r.index][index].(type) {
			case int64:
				*value = s
			case int:
				*value = int64(s)
			}
		case *int:
			switch s := r.data[r.index][index].(type) {
			case int:
				*value = s
			case int64:
				*value = int(s)
			}
		case **[]byte:
			switch s := r.data[r.index][index].(type) {
			case []byte:
				*value = &s
			}
		case **string:
			switch s := r.data[r.index][index].(type) {
			case string:
				*value = &s
			}
		case **int64:
			switch s := r.data[r.index][index].(type) {
			case int64:
				*value = &s
			case int:
				i := int64(s)

				*value = &i
			}
		case **int:
			switch s := r.data[r.index][index].(type) {
			case int:
				*value = &s
			case int64:
				i := int(s)

				*value = &i
			}
		}
	}

	return nil
}
