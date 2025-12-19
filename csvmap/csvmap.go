package csvmap

import (
	"encoding/csv"
	"io"
	"iter"
	"slices"
)

// Reads CSV rows as map[string]string.
type Reader struct {
	reader  *csv.Reader
	columns []string
	err     error
}

// Create a new reader.
func NewReader(r *csv.Reader) *Reader {
	return &Reader{
		reader: r,
	}
}

// Returns a slice containing the columns. Will be nil before rows are read.
func (r *Reader) Columns() []string {
	return slices.Clone(r.columns)
}

// Returns the error encountered during Iter().
func (r *Reader) Err() error {
	return r.err
}

// Returns an iterator of row data maps.
func (r *Reader) Iter() iter.Seq[map[string]string] {
	return func(yield func(map[string]string) bool) {
		cols, err := readHeader(r.reader)
		if err == io.EOF {
			return
		} else if err != nil {
			r.err = err
			return
		}

		r.columns = cols

		for {
			row, err := r.reader.Read()
			if err == io.EOF {
				return
			} else if err != nil {
				r.err = err
				return
			}

			asMap := rowToMap(r.columns, row)
			if !yield(asMap) {
				return
			}
		}
	}
}

func readHeader(r *csv.Reader) ([]string, error) {
	record, err := r.Read()

	if err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(record))

	for _, col := range record {
		columns = append(columns, col)
	}

	return columns, nil
}

func rowToMap(cols []string, row []string) map[string]string {
	colMap := make(map[string]string, len(cols))

	for i, col := range cols {
		if i < len(row) {
			colMap[col] = row[i]
		}
	}

	return colMap
}
