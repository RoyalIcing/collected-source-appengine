package main

import (
	"encoding/csv"
)

// Connection represents a collection of items which can be enumerated through
type Connection interface {
	WriteToCSV(w csv.Writer) error
}
