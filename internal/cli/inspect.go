package cli

import (
	"fmt"

	"github.com/idestis/pipe/internal/hub"
	"github.com/idestis/pipe/internal/parser"
	"github.com/idestis/pipe/internal/resolve"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Show detailed info about a pipeline",
	Args:  exactArgs(1, "pipe inspect <name>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := resolve.Resolve(args[0])
		if err != nil {
			return err
		}

		pipeline, err := parser.LoadPipelineFromPath(ref.Path, ref.Name)
		if err != nil {
			return fmt.Errorf("loading pipeline: %w", err)
		}

		fmt.Printf("Name:        %s\n", ref.Name)
		if ref.Alias != "" {
			fmt.Printf("Alias:       %s\n", ref.Alias)
		}
		fmt.Printf("Source:      %s\n", kindStr(ref.Kind))
		fmt.Printf("Path:        %s\n", ref.Path)

		if pipeline.Description != "" {
			fmt.Printf("Description: %s\n", pipeline.Description)
		}
		fmt.Printf("Steps:       %d\n", len(pipeline.Steps))
		fmt.Printf("Vars:        %d\n", len(pipeline.Vars))

		if ref.Kind == resolve.KindHub {
			// Show HEAD pointer
			headRef, _ := hub.ReadHeadRef(ref.Owner, ref.Pipe)
			if headRef != nil {
				if headRef.Kind == hub.HeadKindBlob {
					fmt.Printf("HEAD:        sha256:%s (detached)\n", headRef.Value[:12])
				} else {
					fmt.Printf("HEAD:        %s\n", headRef.Value)
				}
			}
			fmt.Printf("Active Tag:  %s\n", ref.Tag)

			idx, err := hub.LoadIndex(ref.Owner, ref.Pipe)
			if err == nil && idx != nil && len(idx.Tags) > 0 {
				fmt.Println("\nPulled Tags:")
				for tag, rec := range idx.Tags {
					active := ""
					if tag == idx.ActiveTag {
						active = " (active)"
					}

					// Tag type
					tagType := "symlink"
					if rec.Editable {
						tagType = "editable"
					}

					// Dirty check
					dirtyMarker := ""
					dirty, derr := hub.IsDirty(ref.Owner, ref.Pipe, tag)
					if derr == nil && dirty {
						dirtyMarker = " [dirty]"
					}

					pulledAt := ""
					if !rec.PulledAt.IsZero() {
						pulledAt = fmt.Sprintf("  pulled=%s", rec.PulledAt.Format("2006-01-02 15:04"))
					}
					createdAt := ""
					if !rec.CreatedAt.IsZero() {
						createdAt = fmt.Sprintf("  created=%s", rec.CreatedAt.Format("2006-01-02 15:04"))
					}

					fmt.Printf("  %-16s [%s] sha256=%s%s%s%s%s\n",
						tag, tagType, rec.SHA256[:12], pulledAt, createdAt, active, dirtyMarker)
				}
			}
		}

		return nil
	},
}

func kindStr(k resolve.PipeKind) string {
	switch k {
	case resolve.KindLocal:
		return "local"
	case resolve.KindHub:
		return "hub"
	default:
		return "unknown"
	}
}
