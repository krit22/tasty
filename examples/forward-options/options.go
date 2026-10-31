package render

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ParseOptions decodes JSON options and rejects a negative limit.
// Unknown fields are ignored, and missing fields retain their zero values.
func ParseOptions(data []byte) (Options, error) {
	var options Options
	if err := json.Unmarshal(data, &options); err != nil {
		return Options{}, fmt.Errorf("parse options: %w", err)
	}
	if options.Limit < 0 {
		return Options{}, errors.New("limit must be non-negative")
	}
	return options, nil
}
