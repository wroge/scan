// Scan sql rows into any type powered by generics with proper error handling and automatic resource cleanup.
//
//nolint:wrapcheck,ireturn,structcheck,golint,varnamelen
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
// It returns a pair: destination for scan and a function to process the scan result.
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
	return func(t *T, b []byte) error {
		var v V

		err := json.Unmarshal(b, &v)
		if err != nil {
			return err
		}

		scan(t, v)

		return nil
	}
}

// Columns are used by the utility functions.
type Columns[T any] map[string]Scanner[T]

// First retrieves the first row, scans it, and closes the iterator.
func First[T any](rows Rows, columns Columns[T]) (T, error) {
	var t T

	iter, err := Iter(rows, columns)
	if err != nil {
		return t, err
	}

	if !iter.Next() {
		return t, errors.Join(iter.Err(), iter.Close(), ErrNoRows)
	}

	return t, errors.Join(iter.Scan(&t), iter.Err(), iter.Close())
}

// One retrieves a single row, scans it, and closes the iterator.
func One[T any](rows Rows, columns Columns[T]) (T, error) {
	var t T

	iter, err := Iter(rows, columns)
	if err != nil {
		return t, err
	}

	if !iter.Next() {
		return t, errors.Join(iter.Err(), iter.Close(), ErrNoRows)
	}

	if err = iter.Scan(&t); err != nil {
		return t, errors.Join(err, iter.Err())
	}

	if iter.Next() {
		return t, errors.Join(iter.Err(), iter.Close(), ErrTooManyRows)
	}

	return t, errors.Join(iter.Err(), iter.Close())
}

// All retrieves all rows, scans them into a slice, and closes the iterator.
func All[T any](rows Rows, columns Columns[T]) ([]T, error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return nil, err
	}

	var (
		index = 0
		list  []T
	)

	for iter.Next() {
		list = append(list, *new(T))

		if err = iter.Scan(&list[index]); err != nil {
			return list, errors.Join(err, iter.Err(), iter.Close())
		}

		index++
	}

	return list, errors.Join(iter.Err(), iter.Close())
}

// Limit retrieves up to a specified number of rows, scans them, and closes the iterator.
func Limit[T any](limit int, rows Rows, columns Columns[T]) ([]T, error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return nil, err
	}

	var (
		index = 0
		list  = make([]T, limit)
	)

	for iter.Next() {
		if index >= limit {
			return list, errors.Join(ErrTooManyRows, iter.Err(), iter.Close())
		}

		if err = iter.Scan(&list[index]); err != nil {
			return list, errors.Join(err, iter.Err(), iter.Close())
		}

		index++
	}

	if index < limit {
		list = list[:index]
	}

	return list, errors.Join(iter.Err(), iter.Close())
}

// Iter creates a new iterator.
func Iter[T any](rows Rows, columns Columns[T]) (Iterator[T], error) {
	names, err := rows.Columns()
	if err != nil {
		return Iterator[T]{}, errors.Join(err, rows.Close())
	}

	var (
		dest     = make([]any, len(names))
		scanners = make([]func(*T) error, len(names))
	)

	for i, n := range names {
		if s, ok := columns[n]; ok {
			dest[i], scanners[i] = s.Scan()
		} else {
			dest[i] = new(any)
		}
	}

	return Iterator[T]{
		rows:     rows,
		dest:     dest,
		scanners: scanners,
	}, nil
}

// Iterator for scanning SQL rows.
type Iterator[T any] struct {
	rows     Rows
	dest     []any
	scanners []func(*T) error
}

// Close releases resources of the iterator.
func (i Iterator[T]) Close() error {
	return i.rows.Close()
}

// Err returns any error from iteration process.
func (i Iterator[T]) Err() error {
	return i.rows.Err()
}

// Next advances the iterator to the next row.
func (i Iterator[T]) Next() bool {
	return i.rows.Next()
}

// Scan scans the current row into a value of type T.
func (i Iterator[T]) Scan(t *T) error {
	err := i.rows.Scan(i.dest...)
	if err != nil {
		return err
	}

	for _, s := range i.scanners {
		if s != nil {
			if err = s(t); err != nil {
				return err
			}
		}
	}

	return nil
}

// Value retrieves the current row value.
func (i Iterator[T]) Value() (T, error) {
	var t T

	return t, i.Scan(&t)
}
