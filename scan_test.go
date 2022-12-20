//nolint:exhaustivestruct,exhaustruct,varnamelen,gocritic,goerr113,wrapcheck
package scan_test

import (
	"context"
	"encoding/json"
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

//nolint:gochecknoglobals
var (
	postsResult []Post
	postResult  Post
)

func rows1() *fakeRows {
	return &fakeRows{
		index: -1,
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

var scan1 []scan.Column[Post] = []scan.Column[Post]{
	scan.Any(func(post *Post, id int64) { post.ID = id }),
	scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
	scan.AnyErr(func(post *Post, authors []byte) error { return json.Unmarshal(authors, &post.Authors) }),
}

func TestExample1(t *testing.T) {
	t.Parallel()

	posts, err := scan.All(rows1(), scan1...)
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

func TestExample1Each(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var posts []Post

	err := scan.Each(ctx, func(ctx context.Context, post Post) error {
		posts = append(posts, post)

		return nil
	}, rows1(), scan1...)
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

func BenchmarkExample1WrogeScanAll(b *testing.B) {
	var (
		posts []Post
		err   error
	)

	for n := 0; n < b.N; n++ {
		posts, err = scan.All(rows1(), scan1...)
		if err != nil {
			b.Fatal(err)
		}

		postsResult = posts
	}
}

func BenchmarkExample1WrogeScanEach(b *testing.B) {
	var (
		posts []Post
		ctx   = context.Background()
	)

	for n := 0; n < b.N; n++ {
		err := scan.Each(ctx, func(ctx context.Context, post Post) error {
			posts = append(posts, post)

			return nil
		}, rows1(), scan1...)
		if err != nil {
			b.Fatal(err)
		}

		postsResult = posts
	}
}

func BenchmarkExample1Standard(b *testing.B) {
	var (
		rows        *fakeRows
		err         error
		post        Post
		title       *string
		authorsJSON []byte
	)

	for n := 0; n < b.N; n++ {
		rows = rows1()

		var posts []Post

		for rows.Next() {
			err = rows.Scan(&post.ID, &title, &authorsJSON)
			if err != nil {
				b.Fatal(err)
			}

			if title != nil {
				post.Title = *title
			} else {
				post.Title = "No Title"
			}

			err = json.Unmarshal(authorsJSON, &post.Authors)
			if err != nil {
				b.Error(err)
			}

			posts = append(posts, post)
		}

		err = rows.Err()
		if err != nil {
			b.Error(err)
		}

		postsResult = posts
	}
}

func rows2() *fakeRows {
	return &fakeRows{
		index: -1,
		data: [][]any{
			{1, nil, []byte(`[1,2]`), []byte(`["Jim","Tim"]`)},
			{2, "Post Two", []byte(`[2]`), []byte(`["Tim"]`)},
			{3, "Post Three", []byte(`[2,3]`), []byte(`["Tim","Tom"]`)},
			{4, "Post Four", []byte(`[1,2]`), []byte(`["Jim","Tim"]`)},
			{5, "Post Five", []byte(`[1,3]`), []byte(`["Jim","Tom"]`)},
			{6, "Post Six", []byte(`[2]`), []byte(`["Tim"]`)},
			{7, "Post Seven", []byte(`[3]`), []byte(`["Tom"]`)},
			{8, "Post Eight", []byte(`[1]`), []byte(`["Jim"]`)},
			{9, "Post Nine", []byte(`[1,2,3]`), []byte(`["Jim","Tim","Tom"]`)},
			{10, "Post Ten", []byte(`[3]`), []byte(`["Tom"]`)},
		},
	}
}

var scan2 []scan.Column[Post] = []scan.Column[Post]{
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

func TestExample2(t *testing.T) {
	t.Parallel()

	posts, err := scan.All(rows2(), scan2...)
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

func TestExample2Each(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var posts []Post

	err := scan.Each(ctx, func(ctx context.Context, post Post) error {
		posts = append(posts, post)

		return nil
	}, rows1(), scan1...)
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

func BenchmarkExample2WrogeScanAll(b *testing.B) {
	var (
		posts []Post
		err   error
	)

	for n := 0; n < b.N; n++ {
		posts, err = scan.All(rows2(), scan2...)
		if err != nil {
			b.Fatal(err)
		}

		postsResult = posts
	}
}

func BenchmarkExample2WrogeScanEach(b *testing.B) {
	var (
		posts []Post
		ctx   = context.Background()
	)

	for n := 0; n < b.N; n++ {
		err := scan.Each(ctx, func(ctx context.Context, post Post) error {
			posts = append(posts, post)

			return nil
		}, rows1(), scan1...)
		if err != nil {
			b.Fatal(err)
		}

		postsResult = posts
	}
}

func BenchmarkExample2Standard(b *testing.B) {
	var (
		err             error
		rows            *fakeRows
		post            Post
		title           *string
		authorIdsJSON   []byte
		authorNamesJSON []byte
		authorIds       []int64
		authorNames     []string
		posts           []Post
	)

	for n := 0; n < b.N; n++ {
		rows = rows2()

		for rows.Next() {
			err = rows.Scan(&post.ID, &title, &authorIdsJSON, &authorNamesJSON)
			if err != nil {
				b.Fatal(err)
			}

			if title != nil {
				post.Title = *title
			} else {
				post.Title = "No Title"
			}

			err = json.Unmarshal(authorIdsJSON, &authorIds)
			if err != nil {
				b.Error(err)
			}

			for _, id := range authorIds {
				post.Authors = append(post.Authors, Author{ID: id})
			}

			err = json.Unmarshal(authorNamesJSON, &authorNames)
			if err != nil {
				b.Error(err)
			}

			for i, name := range authorNames {
				post.Authors[i].Name = name
			}

			posts = append(posts, post)
		}

		err = rows.Err()
		if err != nil {
			b.Error(err)
		}

		postsResult = posts
	}
}

func row3() *fakeRows {
	return &fakeRows{
		index: 0,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}
}

var scan3 []scan.Column[Post] = []scan.Column[Post]{
	scan.Any(func(post *Post, id int64) { post.ID = id }),
	scan.Any(func(post *Post, title string) { post.Title = title }),
	scan.AnyErr(func(post *Post, authors []byte) error { return json.Unmarshal(authors, &post.Authors) }),
}

func TestExample3(t *testing.T) {
	t.Parallel()

	post, err := scan.One(row3(), scan3...)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(post) != "{1 Post One [{1 Jim} {2 Tim}]}" {
		t.Fatal(post)
	}
}

func BenchmarkExample3WrogeScanOne(b *testing.B) {
	var (
		post Post
		err  error
	)

	for n := 0; n < b.N; n++ {
		post, err = scan.One(row3(), scan3...)
		if err != nil {
			b.Fatal(err)
		}

		postResult = post
	}
}

func BenchmarkExample3Standard(b *testing.B) {
	var (
		post        Post
		authorsJSON []byte
		row         *fakeRows
		err         error
	)

	for n := 0; n < b.N; n++ {
		row = row3()

		err = row.Scan(&post.ID, &post.Title, &authorsJSON)
		if err != nil {
			b.Fatal(err)
		}

		err = json.Unmarshal(authorsJSON, &post.Authors)
		if err != nil {
			b.Fatal(err)
		}

		postResult = post
	}
}

func TestJSONErr(t *testing.T) {
	t.Parallel()

	_, err := scan.One[Post](row3(),
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSONErr(func(post *Post, authors Author) error { return nil }),
	)
	if err == nil {
		t.Fail()
	}

	ute := &json.UnmarshalTypeError{}

	if !errors.As(err, &ute) {
		t.Fail()
	}
}

func TestCloseErr(t *testing.T) {
	t.Parallel()

	rows := &fakeRows{
		closeErr: fmt.Errorf("closing failed"),
		index:    -1,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}

	_, err := scan.All[Post](rows,
		//nolint:nlreturn
		scan.AnyErr(func(post *Post, id int64) error { post.ID = id; return nil }),
		//nolint:nlreturn
		scan.NullErr("No Title", func(post *Post, title string) error { post.Title = title; return nil }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)
	if err == nil || err.Error() != "wroge/scan error: closing failed" {
		t.Fail()
	}
}

func TestScanErr(t *testing.T) {
	t.Parallel()

	rows := &fakeRows{
		scanErr: fmt.Errorf("scan failed"),
		index:   -1,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}

	_, err := scan.All[Post](rows,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)
	if err == nil {
		t.Fail()
	}
}

func TestEachScanErr(t *testing.T) {
	t.Parallel()

	rows := &fakeRows{
		scanErr: fmt.Errorf("scan failed"),
		index:   -1,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}

	ctx := context.Background()

	err := scan.Each[Post](ctx, func(ctx context.Context, p Post) error {
		return nil
	}, rows,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)
	if err == nil {
		t.Fail()
	}
}

func TestErr(t *testing.T) {
	t.Parallel()

	rows := &fakeRows{
		err:   fmt.Errorf("failed"),
		index: -1,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}

	_, err := scan.All[Post](rows,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)
	if err == nil {
		t.Fail()
	}
}

func TestErrEach(t *testing.T) {
	t.Parallel()

	rows := &fakeRows{
		index: -1,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}

	ctx := context.Background()

	err := scan.Each[Post](ctx, func(ctx context.Context, p Post) error {
		return fmt.Errorf("fail")
	}, rows,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)
	if err == nil {
		t.Fail()
	}
}

func TestRowErr(t *testing.T) {
	t.Parallel()

	row := &fakeRows{
		scanErr: fmt.Errorf("scan failed"),
		index:   0,
		data: [][]any{
			{1, "Post One", []byte(`[{"id": 1, "name": "Jim"},{"id": 2, "name": "Tim"}]`)},
		},
	}

	_, err := scan.One[Post](row,
		scan.Any(func(post *Post, id int64) { post.ID = id }),
		scan.Null("No Title", func(post *Post, title string) { post.Title = title }),
		scan.JSON(func(post *Post, authors []Author) { post.Authors = authors }),
	)
	if err == nil {
		t.Fail()
	}
}

type fakeRows struct {
	closeErr error
	scanErr  error
	err      error
	index    int
	data     [][]any
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

	for i, d := range dest {
		switch t := d.(type) {
		case *[]byte:
			switch s := r.data[r.index][i].(type) {
			case []byte:
				*t = s
			}
		case *string:
			switch s := r.data[r.index][i].(type) {
			case string:
				*t = s
			}
		case *int64:
			switch s := r.data[r.index][i].(type) {
			case int64:
				*t = s
			case int:
				*t = int64(s)
			}
		case *int:
			switch s := r.data[r.index][i].(type) {
			case int:
				*t = s
			case int64:
				*t = int(s)
			}
		case **[]byte:
			switch s := r.data[r.index][i].(type) {
			case []byte:
				*t = &s
			}
		case **string:
			switch s := r.data[r.index][i].(type) {
			case string:
				*t = &s
			}
		case **int64:
			switch s := r.data[r.index][i].(type) {
			case int64:
				*t = &s
			case int:
				i := int64(s)

				*t = &i
			}
		case **int:
			switch s := r.data[r.index][i].(type) {
			case int:
				*t = &s
			case int64:
				i := int(s)

				*t = &i
			}
		}
	}

	return nil
}
