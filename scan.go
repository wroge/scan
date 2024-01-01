// Scan provides utility functions for scanning SQL rows into custom types using a flexible iterator approach.
// It supports custom scanners for individual columns, error handling, and resource cleanup.
// The package is designed to handle common scenarios such as retrieving the first row, all rows, or a single row from a
// result set. Custom scanning functions can be easily integrated to handle specific data types or processing logic.
// The Iterator type facilitates efficient row iteration, scanning, and error management.
// Users can create iterators using the Iter function, and then use the provided methods like All, One, or First for
// different use cases.
//
//nolint:wrapcheck,ireturn,structcheck,golint
package scan

import (
	"encoding/json"
	"errors"
)

// ErrNoRows represents the error indicating no rows in the result set.
var ErrNoRows = errors.New("sql: no rows in result set")

// ErrTooManyRows represents the error indicating too many rows in the result set.
var ErrTooManyRows = errors.New("sql: too many rows in result set")

// Rows defines the interface for a SQL result set, including methods for iterating, scanning, and error handling.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Columns() ([]string, error)
	Err() error
	Close() error
}

// Scanner defines the interface for a custom column scanner.
type Scanner[T any] interface {
	Scan() (any, func(*T) error)
}

// Func is a function type that defines a custom scanner for a column.
type Func[T, V any] func(*T, V) error

// Scan implements the Scan method for the Func type.
func (f Func[T, V]) Scan() (any, func(*T) error) {
	var v V

	return &v, func(t *T) error {
		return f(t, v)
	}
}

// Any creates a custom scanner for a column with a specified scan function.
func Any[T, V any](scan func(*T, V)) Func[T, V] {
	return func(t *T, v V) error {
		scan(t, v)

		return nil
	}
}

// Null creates a custom scanner for a nullable column with a specified default value and scan function.
func Null[T, V any](def V, scan func(*T, V)) Func[T, *V] {
	return func(t *T, v *V) error {
		if v != nil {
			scan(t, *v)
		} else {
			scan(t, def)
		}

		return nil
	}
}

// JSON creates a custom scanner for a column containing JSON data with a specified scan function.
func JSON[T, V any](scan func(*T, V)) Func[T, []byte] {
	return func(typ *T, b []byte) error {
		var value V

		err := json.Unmarshal(b, &value)
		if err != nil {
			return err
		}

		scan(typ, value)

		return nil
	}
}

// Columns are used by the utility functions.
type Columns[T any] map[string]Scanner[T]

// First retrieves the first row from the iterator, scans it into a value of type T, and closes the iterator.
// It returns the scanned value and any encountered error during scanning, closing, or if no rows are found.
// The method handles errors gracefully by using error accumulation and specifically identifies the case
// of no rows found using the ErrNoRows error.
func First[T any](rows Rows, columns Columns[T]) (T, error) {
	var t T

	iter, err := Iter(rows, columns)
	if err != nil {
		return t, err
	}

	return iter.First()
}

// One retrieves a single row from the iterator, scans it into a value of type T, and closes the iterator.
// It returns the scanned value and any encountered error during scanning, closing, or if no rows are found.
// The method handles errors gracefully by using error accumulation and provides specific error types for
// cases such as no rows found or multiple rows encountered.
func One[T any](rows Rows, columns Columns[T]) (T, error) {
	var t T

	iter, err := Iter(rows, columns)
	if err != nil {
		return t, err
	}

	return iter.One()
}

// All retrieves all rows from the iterator, scans them into a slice of type T, and closes the iterator.
// It returns the populated slice and any encountered error during scanning or closing.
// The method efficiently handles errors by using error accumulation and ensures proper resource cleanup.
func All[T any](rows Rows, columns Columns[T]) ([]T, error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return nil, err
	}

	return iter.All()
}

// Limit retrieves a maximum number of rows from the iterator, scans them into a slice of type T, and
// closes the iterator. It returns the populated slice and any encountered error during scanning or closing.
// The method efficiently handles errors by using error accumulation and ensures proper resource cleanup.
func Limit[T any](rows Rows, columns Columns[T], limit int) ([]T, error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return nil, err
	}

	return iter.Limit(limit)
}

// Iter creates a new iterator for the given SQL rows and a map of column names to custom scanners.
// It returns the initialized iterator and any encountered error during column retrieval or iterator creation.
// The method efficiently handles errors by using error accumulation and ensures proper resource cleanup.
func Iter[T any](rows Rows, columns Columns[T]) (Iterator[T], error) {
	names, err := rows.Columns()
	if err != nil {
		return Iterator[T]{}, errors.Join(err, rows.Close())
	}

	var (
		dest     = make([]any, len(names))
		scanners = make([]func(*T) error, len(names))
		ignore   any
	)

	for i, n := range names {
		if s, ok := columns[n]; ok {
			dest[i], scanners[i] = s.Scan()
		} else {
			dest[i] = &ignore
		}
	}

	return Iterator[T]{
		rows:     rows,
		dest:     dest,
		scanners: scanners,
	}, nil
}

// Iterator represents an iterator for scanning rows from a SQL result set.
type Iterator[T any] struct {
	rows     Rows
	dest     []any
	scanners []func(*T) error
}

// Close closes the underlying SQL rows, releasing associated resources.
func (i Iterator[T]) Close() error {
	return i.rows.Close()
}

// Err returns any error encountered during the iteration process.
func (i Iterator[T]) Err() error {
	return i.rows.Err()
}

// Next advances the iterator to the next row in the result set.
func (i Iterator[T]) Next() bool {
	return i.rows.Next()
}

// Scan scans the current row of the iterator into the provided value of type T.
// It internally uses the underlying SQL rows.Scan method and then applies any custom scanners
// provided during the iterator's initialization. It returns any encountered error during scanning.
func (i Iterator[T]) Scan(typ *T) error {
	err := i.rows.Scan(i.dest...)
	if err != nil {
		return err
	}

	for _, s := range i.scanners {
		if s != nil {
			err = s(typ)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Value uses the Scan method to return T.
func (i Iterator[T]) Value() (T, error) {
	var t T

	return t, i.Scan(&t)
}

// All retrieves all rows from the iterator, scans them into a slice of type T, and closes the iterator.
// It returns the populated slice and any encountered error during scanning or closing.
// The method efficiently handles errors by using error accumulation and ensures proper resource cleanup.
func (i Iterator[T]) All() ([]T, error) {
	var (
		index = 0
		list  []T
		err   error
	)

	for i.Next() {
		list = append(list, *new(T))

		err = i.Scan(&list[index])
		if err != nil {
			return nil, errors.Join(err, i.Err(), i.Close())
		}

		index++
	}

	return list, errors.Join(i.Err(), i.Close())
}

// Max retrieves a maximal number of rows from the iterator, scans them into a slice of type T, and closes
// the iterator. It returns the populated slice and any encountered error during scanning or closing.
// The method efficiently handles errors by using error accumulation and ensures proper resource cleanup.
func (i Iterator[T]) Limit(limit int) ([]T, error) {
	var (
		index = 0
		list  = make([]T, limit)
		err   error
	)

	for i.Next() {
		if index >= limit {
			return nil, errors.Join(err, i.Err(), i.Close(), ErrTooManyRows)
		}

		err = i.Scan(&list[index])
		if err != nil {
			return nil, errors.Join(err, i.Err(), i.Close())
		}

		index++
	}

	if index < limit {
		list = list[:index]
	}

	return list, errors.Join(i.Err(), i.Close())
}

// One retrieves a single row from the iterator, scans it into a value of type T, and closes the iterator.
// It returns the scanned value and any encountered error during scanning, closing, or if no rows are found.
// The method handles errors gracefully by using error accumulation and provides specific error types for
// cases such as no rows found or multiple rows encountered.
func (i Iterator[T]) One() (T, error) {
	var (
		typ T
		err error
	)

	if !i.Next() {
		return typ, errors.Join(i.Err(), i.Close(), ErrNoRows)
	}

	err = i.Scan(&typ)
	if err != nil {
		return typ, errors.Join(err, i.Err())
	}

	if i.Next() {
		return typ, errors.Join(i.Err(), i.Close(), ErrTooManyRows)
	}

	return typ, errors.Join(i.Err(), i.Close())
}

// First retrieves the first row from the iterator, scans it into a value of type T, and closes the iterator.
// It returns the scanned value and any encountered error during scanning, closing, or if no rows are found.
// The method handles errors gracefully by using error accumulation and specifically identifies the case
// of no rows found using the ErrNoRows error.
func (i Iterator[T]) First() (T, error) {
	var typ T

	if !i.Next() {
		return typ, errors.Join(i.Err(), i.Close(), ErrNoRows)
	}

	return typ, errors.Join(i.Scan(&typ), i.Err(), i.Close())
}
