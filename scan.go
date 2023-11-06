//nolint:wrapcheck,nilerr,ireturn,structcheck,nonamedreturns,golint
package scan

import (
	"database/sql"
	"encoding/json"
	"errors"
)

var (
	ErrNoRows      = errors.New("sql: no rows in result set")
	ErrTooManyRows = errors.New("sql: too many rows in result set")
)

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Columns() ([]string, error)
	Err() error
	Close() error
}

type Scanner[T any] interface {
	Scan() (any, func(*T) error)
}

type Func[T, V any] func(*T, V) error

func (f Func[T, V]) Scan() (any, func(*T) error) {
	var v V

	return &v, func(t *T) error {
		return f(t, v)
	}
}

func Any[T, V any](scan func(*T, V)) Func[T, V] {
	return func(t *T, v V) error {
		scan(t, v)

		return nil
	}
}

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

func JSON[T, V any](scan func(*T, V)) Func[T, []byte] {
	return func(typ *T, b []byte) error {
		var value V

		err := json.Unmarshal(b, &value)
		if err != nil {
			return nil
		}

		scan(typ, value)

		return nil
	}
}

func First[T any](rows Rows, columns map[string]Scanner[T]) (t T, err error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return t, err
	}

	if !iter.Next() {
		return t, errors.Join(iter.Close(), sql.ErrNoRows)
	}

	return t, errors.Join(iter.Scan(&t), iter.Err(), iter.Close())
}

func One[T any](rows Rows, columns map[string]Scanner[T]) (t T, err error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return t, err
	}

	if !iter.Next() {
		return t, errors.Join(iter.Close(), sql.ErrNoRows, ErrNoRows)
	}

	err = iter.Scan(&t)
	if err != nil {
		return t, errors.Join(err, iter.Err())
	}

	if iter.Next() {
		return t, errors.Join(iter.Close(), ErrTooManyRows)
	}

	return t, errors.Join(iter.Err(), iter.Close())
}

func All[T any](rows Rows, columns map[string]Scanner[T]) ([]T, error) {
	iter, err := Iter(rows, columns)
	if err != nil {
		return nil, err
	}

	var (
		list  []T
		index = 0
	)

	for iter.Next() {
		list = append(list, *new(T))

		err = iter.Scan(&list[index])
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		index++
	}

	return list, errors.Join(iter.Err(), iter.Close())
}

func Iter[T any](rows Rows, columns map[string]Scanner[T]) (Iterator[T], error) {
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

type Iterator[T any] struct {
	rows     Rows
	dest     []any
	scanners []func(*T) error
}

func (i Iterator[T]) Close() error {
	return i.rows.Close()
}

func (i Iterator[T]) Err() error {
	return i.rows.Err()
}

func (i Iterator[T]) Next() bool {
	return i.rows.Next()
}

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
