package csvmap_test

import (
	"bytes"
	"encoding/csv"
	"maps"
	"slices"
	"testing"

	"github.com/Open-Event-Systems/go-utils/csvmap"
)

func TestCSV(t *testing.T) {
	data := []byte("a,b,c\na1,b1,c1\na2,\"b,2\",c2\n")
	buf := bytes.NewBuffer(data)

	reader := csvmap.NewReader(csv.NewReader(buf))
	rows := slices.Collect(reader.Iter())

	if reader.Err() != nil {
		t.Fatal(reader.Err())
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	expectedCols := []string{"a", "b", "c"}
	if !slices.Equal(reader.Columns(), expectedCols) {
		t.Fatalf("expected %v, got %v", expectedCols, reader.Columns())
	}

	expected := map[string]string{"a": "a1", "b": "b1", "c": "c1"}
	if !maps.Equal(rows[0], expected) {
		t.Fatalf("expected %v, got %v", expected, rows[0])
	}

	expected = map[string]string{"a": "a2", "b": "b,2", "c": "c2"}
	if !maps.Equal(rows[1], expected) {
		t.Fatalf("expected %v, got %v", expected, rows[1])
	}
}

func TestCSVEmpty(t *testing.T) {
	data := []byte("")
	buf := bytes.NewBuffer(data)

	reader := csvmap.NewReader(csv.NewReader(buf))
	rows := slices.Collect(reader.Iter())

	if reader.Err() != nil {
		t.Fatal(reader.Err())
	}

	if len(rows) != 0 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if len(reader.Columns()) != 0 {
		t.Fatalf("expected 0 columns, got %d", len(reader.Columns()))
	}
}

func TestCSVError(t *testing.T) {
	data := []byte("a,b,c\nx")
	buf := bytes.NewBuffer(data)

	reader := csvmap.NewReader(csv.NewReader(buf))
	for range reader.Iter() {
	}

	if reader.Err() == nil {
		t.Fatal("expected error")
	}
}
