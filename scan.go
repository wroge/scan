//nolint:ireturn,wrapcheck
package scan

import (
	"encoding/json"
	"fmt"
)

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Err() error
	Next() bool
	Row
}

// Column provides a stable pointer via Scan, so that
// Set can access the value and set it into *T.
type Column[T any] interface {
	Scan() any
	Set(*T) error
}

// AnyColumn is a typesafe Column to Scan and Set V for each Row.
type AnyColumn[T, V any] struct {
	Setter func(typ *T, value V) error

	scan V
}

func (c *AnyColumn[T, V]) Scan() any {
	return &c.scan
}

func (c *AnyColumn[T, V]) Set(typ *T) error {
	return c.Setter(typ, c.scan)
}

// AnyErr produces a Column.
func AnyErr[T, V any](setter func(*T, V) error) *AnyColumn[T, V] {
	return &AnyColumn[T, V]{
		Setter: setter,
	}
}

// Any is like AnyErr but omits the error.
func Any[T, V any](setter func(*T, V)) *AnyColumn[T, V] {
	return AnyErr(func(typ *T, value V) error {
		setter(typ, value)

		return nil
	})
}

// NullErr produces a Column that can scan nullable values and
// sets a default value if its null.
func NullErr[T, V any](def V, setter func(*T, V) error) *AnyColumn[T, *V] {
	return AnyErr(func(typ *T, value *V) error {
		if value == nil {
			return setter(typ, def)
		}

		return setter(typ, *value)
	})
}

// Null is like NullErr but omits the error.
func Null[T, V any](def V, setter func(*T, V)) *AnyColumn[T, *V] {
	return Any(func(typ *T, value *V) {
		if value == nil {
			setter(typ, def)
		} else {
			setter(typ, *value)
		}
	})
}

// JSONErr produces a Column that scans json into bytes and
// unmarshals it into V.
func JSONErr[T, V any](setter func(*T, V) error) *AnyColumn[T, []byte] {
	return AnyErr(func(typ *T, b []byte) error {
		var value V

		err := json.Unmarshal(b, &value)
		if err != nil {
			return err
		}

		return setter(typ, value)
	})
}

// JSON is like JSONErr but omits the error.
func JSON[T, V any](setter func(*T, V)) *AnyColumn[T, []byte] {
	return JSONErr(func(typ *T, value V) error {
		setter(typ, value)

		return nil
	})
}

func doClose(rows any, wrap error) error {
	switch r := rows.(type) {
	case interface{ Close() }:
		r.Close()
	case interface{ Close() error }:
		err := r.Close()
		if err != nil {
			return CloseError{
				Err:  err,
				Wrap: wrap,
			}
		}
	}

	if wrap != nil {
		return fmt.Errorf("wroge/scan error: %w", wrap)
	}

	return nil
}

// All returns a slice of T from rows and columns.
// Close is called automatically.
func All[T any](rows Rows, columns ...Column[T]) ([]T, error) {
	var (
		err  error
		out  []T
		dest = make([]any, len(columns))
	)

	for i, column := range columns {
		dest[i] = column.Scan()
	}

	count := 0

	for rows.Next() {
		//nolint:gocritic
		out = append(out, *new(T))

		err = rows.Scan(dest...)
		if err != nil {
			return nil, doClose(rows, err)
		}

		for _, column := range columns {
			err = column.Set(&out[count])
			if err != nil {
				return nil, doClose(rows, err)
			}
		}

		count++
	}

	return out, doClose(rows, rows.Err())
}

// All returns T from a row and columns.
func One[T any](row Row, columns ...Column[T]) (T, error) {
	var out T

	dest := make([]any, len(columns))

	for i, column := range columns {
		dest[i] = column.Scan()
	}

	err := row.Scan(dest...)
	if err != nil {
		return out, fmt.Errorf("wroge/scan error: %w", err)
	}

	for _, column := range columns {
		err = column.Set(&out)
		if err != nil {
			return out, fmt.Errorf("wroge/scan error: %w", err)
		}
	}

	return out, nil
}

// CloseError is returned if the closing of rows fails.
type CloseError struct {
	Err  error
	Wrap error
}

func (ce CloseError) Error() string {
	return fmt.Sprintf("wroge/scan error: %s", ce.Err)
}

func (ce CloseError) Unwrap() error {
	return ce.Wrap
}
