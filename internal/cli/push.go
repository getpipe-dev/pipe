package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/idestis/pipe/internal/config"
	"github.com/idestis/pipe/internal/hub"
	"github.com/idestis/pipe/internal/resolve"
	"github.com/spf13/cobra"
)

var pushTags []string

func init() {
	pushCmd.Flags().StringArrayVarP(&pushTags, "tag", "t", nil, "tags to assign (repeatable, e.g. -t latest -t v2.0.0)")
}

var pushCmd = &cobra.Command{
	Use:   "push <owner>/<name>[:<tag>]",
	Short: "Push a pipeline to PipeHub",
	Args:  exactArgs(1, "pipe push <owner>/<name>[:<tag>]"),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := requireAuth()
		if err != nil {
			return err
		}

		owner, name, inlineTag := resolve.ParsePipeArg(args[0])
		if owner == "" {
			return fmt.Errorf("owner required — use \"pipe push <owner>/<name>[:<tag>]\"")
		}

		// Build tag list: -t flags take precedence, then inline :tag, then default "latest"
		tags := pushTags
		if len(tags) == 0 {
			if inlineTag != "" {
				tags = []string{inlineTag}
			} else {
				tags = []string{"latest"}
			}
		}

		// Resolve source content: try exact tag file, then active tag, then local files
		var content []byte
		var sourceIsHub bool

		// Try the first requested tag's file directly
		hubPath := hub.ContentPath(owner, name, tags[0])
		if _, err := os.Stat(hubPath); err == nil {
			content, err = os.ReadFile(hubPath)
			if err != nil {
				return fmt.Errorf("reading hub pipe: %w", err)
			}
			sourceIsHub = true
		} else {
			// Fall back to active tag
			idx, _ := hub.LoadIndex(owner, name)
			if idx != nil && idx.ActiveTag != "" {
				activeHubPath := hub.ContentPath(owner, name, idx.ActiveTag)
				if data, err := os.ReadFile(activeHubPath); err == nil {
					content = data
					sourceIsHub = true
				}
			}
		}

		if content == nil {
			// Try local files
			localPath := filepath.Join(config.FilesDir, name+".yaml")
			data, err := os.ReadFile(localPath)
			if err != nil {
				return fmt.Errorf("pipe %q not found in hub store or local files", owner+"/"+name)
			}
			content = data
		}

		// Dirty check for editable tags
		if sourceIsHub {
			idx, _ := hub.LoadIndex(owner, name)
			if idx != nil {
				activeTag := idx.ActiveTag
				if activeTag != "" {
					rec, ok := idx.Tags[activeTag]
					if ok {
						dirty, derr := hub.IsDirty(owner, name, activeTag)
						if derr == nil && dirty {
							if rec.Editable {
								log.Warn("editable tag has local modifications, pushing current content", "tag", activeTag)
							} else {
								log.Warn("local modifications detected, pushing current content", "tag", activeTag)
							}
						}
					}
				}
			}
		}

		client := newHubClient(creds)

		// Check if pipe exists, auto-create if not
		meta, err := client.GetPipe(owner, name)
		if err != nil {
			return fmt.Errorf("checking pipe: %w", err)
		}
		if meta == nil {
			log.Info("pipe not found on hub, creating", "pipe", owner+"/"+name)
			_, err := client.CreatePipe(owner, &hub.CreatePipeRequest{
				Name:     name,
				IsPublic: true,
			})
			if err != nil {
				return fmt.Errorf("creating pipe: %w", err)
			}
		}

		log.Info("pushing", "pipe", owner+"/"+name, "tags", tags, "size", len(content))
		resp, err := client.Push(owner, name, content, tags)
		if err != nil {
			return fmt.Errorf("pushing: %w", err)
		}

		// Verify response digest against local checksum
		localSHA, localMD5 := hub.ComputeChecksums(content)
		expectedDigest := "sha256:" + localSHA
		if resp.Digest != expectedDigest {
			return fmt.Errorf("digest mismatch after push — local %s, remote %s", expectedDigest, resp.Digest)
		}

		// Re-snapshot: write pushed content as a correctly-named blob,
		// re-point the active tag symlink, and update its index record.
		if sourceIsHub {
			idx, _ := hub.LoadIndex(owner, name)
			if idx != nil && idx.ActiveTag != "" {
				newSha, err := hub.WriteBlob(owner, name, content)
				if err != nil {
					return fmt.Errorf("writing blob after push: %w", err)
				}
				if err := hub.CreateTagSymlink(owner, name, idx.ActiveTag, newSha); err != nil {
					return fmt.Errorf("re-pointing active tag: %w", err)
				}
				idx.Tags[idx.ActiveTag] = hub.TagRecord{
					SHA256:    localSHA,
					MD5:       localMD5,
					SizeBytes: resp.SizeBytes,
					PulledAt:  idx.Tags[idx.ActiveTag].PulledAt,
				}
				if err := hub.SaveIndex(idx); err != nil {
					return fmt.Errorf("updating index: %w", err)
				}
				if err := hub.GarbageCollectBlobs(owner, name); err != nil {
					log.Warn("garbage collection failed", "err", err)
				}
			}
		}

		if resp.Created {
			log.Info("pushed successfully", "pipe", owner+"/"+name, "tags", resp.Tags, "digest", resp.Digest[:19])
		} else {
			log.Info("content already exists", "pipe", owner+"/"+name, "tags", resp.Tags, "digest", resp.Digest[:19])
		}
		return nil
	},
}
