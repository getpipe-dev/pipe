package cli

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/idestis/pipe/internal/hub"
	"github.com/idestis/pipe/internal/parser"
	"github.com/idestis/pipe/internal/resolve"
	"github.com/spf13/cobra"
)

var pullForce bool

func init() {
	pullCmd.Flags().BoolVarP(&pullForce, "force", "f", false, "overwrite local changes")
}

var pullCmd = &cobra.Command{
	Use:   "pull <owner>/<name>[:<tag>]",
	Short: "Pull a pipeline from PipeHub",
	Args:  exactArgs(1, "pipe pull <owner>/<name>[:<tag>]"),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := requireAuth()
		if err != nil {
			return err
		}

		owner, name, tag := resolve.ParsePipeArg(args[0])
		if owner == "" {
			return fmt.Errorf("owner required — use \"pipe pull <owner>/<name>[:<tag>]\"")
		}
		if tag == "" {
			tag = "latest"
		}

		// Check for local modifications before overwriting
		if !pullForce {
			dirty, err := hub.IsDirty(owner, name, tag)
			if err != nil {
				log.Warn("could not check for local changes", "err", err)
			} else if dirty {
				return fmt.Errorf("local changes to %s/%s:%s would be overwritten — push first or use --force", owner, name, tag)
			}
		}

		client := newHubClient(creds)

		log.Info("fetching tag metadata", "pipe", owner+"/"+name, "tag", tag)
		detail, err := client.GetTag(owner, name, tag)
		if err != nil {
			return fmt.Errorf("fetching tag info: %w", err)
		}

		log.Info("downloading content", "size", detail.SizeBytes)
		content, err := client.DownloadTag(owner, name, tag)
		if err != nil {
			return fmt.Errorf("downloading content: %w", err)
		}

		// Verify checksum
		sha, _ := hub.ComputeChecksums(content)
		if sha != detail.SHA256 {
			return fmt.Errorf("checksum mismatch — expected %s, got %s", detail.SHA256, sha)
		}

		// Write content to disk
		if err := hub.SaveContent(owner, name, tag, content); err != nil {
			return fmt.Errorf("saving content: %w", err)
		}

		// Update index
		if err := hub.UpdateIndex(owner, name, tag, detail.SHA256, detail.MD5, detail.SizeBytes); err != nil {
			return fmt.Errorf("updating index: %w", err)
		}

		// Validate YAML
		path := hub.ContentPath(owner, name, tag)
		if _, err := parser.LoadPipelineFromPath(path, owner+"/"+name); err != nil {
			log.Warn("pulled content has validation issues", "err", err)
		}

		log.Info("pulled successfully", "pipe", owner+"/"+name, "tag", tag, "sha256", sha[:12])
		return nil
	},
}
