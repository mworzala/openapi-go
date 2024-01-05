package oapi_rt

import (
	"encoding/json"
	"net/http"
)

func ReadExplodedQuery(r *http.Request, target interface{}) error {
	// This goes to json and back to avoid a dependency on something like mapstructure, may change in the future

	entries := make(map[string]interface{})
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue // sanity check
		}
		entries[key] = values[0]
	}

	raw, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, target)
}
