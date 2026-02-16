package model

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// CacheField supports two YAML forms:
//   - bool:    cached: true → {Enabled: true, ExpireAfter: ""}
//   - mapping: cached: {expireAfter: "1h"} → {Enabled: true, ExpireAfter: "1h"}
type CacheField struct {
	Enabled     bool
	ExpireAfter string
}

func (c *CacheField) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var b bool
		if err := value.Decode(&b); err != nil {
			return fmt.Errorf("cached: expected true/false, got %q", value.Value)
		}
		c.Enabled = b
		return nil

	case yaml.MappingNode:
		var m struct {
			ExpireAfter string `yaml:"expireAfter"`
		}
		if err := value.Decode(&m); err != nil {
			return fmt.Errorf("cached: decoding mapping: %w", err)
		}
		c.Enabled = true
		c.ExpireAfter = m.ExpireAfter
		return nil

	default:
		return fmt.Errorf("cached: must be a bool or a mapping with expireAfter")
	}
}
