package cli

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/idestis/pipe/internal/parser"
	"github.com/idestis/pipe/internal/resolve"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <name>",
	Short: "Validate a pipeline",
	Args:  exactArgs(1, "pipe validate <name>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := resolve.Resolve(args[0])
		if err != nil {
			return err
		}

		pipeline, err := parser.LoadPipelineFromPath(ref.Path, ref.Name)
		if err != nil {
			if isYAMLError(err) {
				return fmt.Errorf("invalid YAML in pipeline %q: %v", ref.Name, unwrapYAMLError(err))
			}
			return err
		}
		for _, w := range parser.Warnings(pipeline) {
			log.Warn(w)
		}
		fmt.Printf("pipeline %q is valid\n", ref.Name)
		return nil
	},
}
