package render

import (
	"encoding/json"
	"os"

	"aks-coach/internal/compute"
)

// PrintJSON encodes the given slice of compute.Row to JSON format and writes it to standard output in an indented format.
// It returns an error if the encoding process fails.
func PrintJSON(rows []compute.Row) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
