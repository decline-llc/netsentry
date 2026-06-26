package rule

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/decline-llc/netsentry/pkg/model"
)

type rulesFile struct {
	Rules []*model.Rule `json:"rules"`
}

// LoadFromFile reads a rules JSON file and returns the parsed rules.
func LoadFromFile(path string) ([]*model.Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules %s: %w", path, err)
	}
	var rf rulesFile
	if err := json.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("parse rules %s: %w", path, err)
	}
	// Set sensible defaults.
	for _, r := range rf.Rules {
		if r.Priority == 0 {
			r.Priority = 100
		}
		if r.Config == nil {
			r.Config = json.RawMessage("{}")
		}
	}
	return rf.Rules, nil
}
