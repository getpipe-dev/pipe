package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/destis/pipe/internal/config"
	"github.com/destis/pipe/internal/model"
	"gopkg.in/yaml.v3"
)

func LoadPipeline(name string) (*model.Pipeline, error) {
	path := filepath.Join(config.FilesDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pipeline %q: %w", name, err)
	}

	var p model.Pipeline
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing pipeline %q: %w", name, err)
	}

	if p.Name == "" {
		p.Name = name
	}

	if err := validate(&p); err != nil {
		return nil, fmt.Errorf("validating pipeline %q: %w", name, err)
	}
	return &p, nil
}

func validate(p *model.Pipeline) error {
	ids := make(map[string]bool)
	for i, s := range p.Steps {
		if s.ID == "" {
			return fmt.Errorf("step %d: missing id", i)
		}
		if ids[s.ID] {
			return fmt.Errorf("step %d: duplicate id %q", i, s.ID)
		}
		ids[s.ID] = true

		if !s.Run.IsSingle() && !s.Run.IsStrings() && !s.Run.IsSubRuns() {
			return fmt.Errorf("step %q: missing run field", s.ID)
		}
	}
	return nil
}
