package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/idestis/pipe/internal/hub"
	"github.com/idestis/pipe/internal/resolve"
	"github.com/spf13/cobra"
)

var (
	tagDelete bool
	tagForce  bool
)

func init() {
	tagCmd.Flags().BoolVarP(&tagDelete, "delete", "d", false, "delete the specified tag")
	tagCmd.Flags().BoolVarP(&tagForce, "force", "f", false, "overwrite an existing tag")
}

var tagCmd = &cobra.Command{
	Use:   "tag <owner>/<name> [tag]",
	Short: "List, create, or delete tags for a hub pipeline",
	Long: `Manage tags for a hub pipeline.

Without a tag argument, lists all tags.
With a tag argument, creates a new tag pointing to the same content as HEAD.
With -d, deletes the specified tag.`,
	Args: rangeArgs(1, 2, "pipe tag <owner>/<name> [tag]"),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, name, _ := resolve.ParsePipeArg(args[0])
		if owner == "" {
			return fmt.Errorf("owner required — use \"pipe tag <owner>/<name> [tag]\"")
		}

		idx, err := hub.LoadIndex(owner, name)
		if err != nil {
			return err
		}
		if idx == nil {
			return fmt.Errorf("no index found for %s/%s — run \"pipe pull %s/%s\" first", owner, name, owner, name)
		}

		// No tag argument → list tags
		if len(args) < 2 {
			return listTags(owner, name, idx)
		}

		tag := args[1]

		// -d flag → delete
		if tagDelete {
			return deleteTag(owner, name, tag, idx)
		}

		// Create new tag from HEAD content
		return createTag(owner, name, tag, idx)
	},
}

func listTags(owner, name string, idx *hub.Index) error {
	if len(idx.Tags) == 0 {
		fmt.Printf("no tags for %s/%s\n", owner, name)
		return nil
	}

	headRef, _ := hub.ReadHeadRef(owner, name)

	tags := sortedTags(idx)
	for _, tag := range tags {
		rec := idx.Tags[tag]

		pointer := "  "
		if headRef != nil && headRef.Kind == hub.HeadKindTag && tag == headRef.Value {
			pointer = "* "
		}

		tagType := "symlink"
		if rec.Editable {
			tagType = "editable"
		}

		dirtyMarker := ""
		dirty, derr := hub.IsDirty(owner, name, tag)
		if derr == nil && dirty {
			dirtyMarker = " [dirty]"
		}

		fmt.Printf("%s%-16s [%s] sha256:%s%s\n", pointer, tag, tagType, rec.SHA256[:12], dirtyMarker)
	}

	// Show detached HEAD if pointing to a blob
	if headRef != nil && headRef.Kind == hub.HeadKindBlob {
		fmt.Printf("* %-16s sha256:%s\n", "(detached)", headRef.Value[:12])
	}

	return nil
}

func deleteTag(owner, name, tag string, idx *hub.Index) error {
	if _, ok := idx.Tags[tag]; !ok {
		return fmt.Errorf("tag %q not found for %s/%s", tag, owner, name)
	}

	if len(idx.Tags) == 1 {
		return fmt.Errorf("cannot delete the only tag — use \"pipe rm %s/%s\" to remove the entire pipe", owner, name)
	}

	if err := hub.DeleteTag(owner, name, tag); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}

	log.Info("deleted tag", "pipe", owner+"/"+name, "tag", tag)

	// If we deleted the active tag, suggest switching
	if tag == idx.ActiveTag {
		log.Warn("deleted the active tag — run \"pipe switch\" to select a new one")
	}
	return nil
}

func createTag(owner, name, tag string, idx *hub.Index) error {
	if _, ok := idx.Tags[tag]; ok && !tagForce {
		return fmt.Errorf("tag %q already exists for %s/%s — use -f to overwrite", tag, owner, name)
	}

	var content []byte
	var sha, md5h string
	sourceLabel := idx.ActiveTag

	// Resolve HEAD to get current content
	if idx.ActiveTag == "" {
		// No active tag — check if HEAD points to a blob
		headRef, err := hub.ReadHeadRef(owner, name)
		if err != nil || headRef.Value == "" {
			return fmt.Errorf("no active tag — run \"pipe switch %s/%s <tag>\" first", owner, name)
		}
		if headRef.Kind == hub.HeadKindBlob {
			blobPath := hub.BlobPath(owner, name, headRef.Value)
			content, err = os.ReadFile(blobPath)
			if err != nil {
				return fmt.Errorf("reading blob %s: %w", headRef.Value[:12], err)
			}
			sha = headRef.Value
			_, md5h = hub.ComputeChecksums(content)
			sourceLabel = "sha256:" + headRef.Value[:12]

			// Create tag symlink pointing to the existing blob
			if err := hub.CreateTagSymlink(owner, name, tag, sha); err != nil {
				return fmt.Errorf("creating tag symlink: %w", err)
			}
		} else {
			// HeadKindTag but not in index — try to read it
			sourceLabel = headRef.Value
			content, err = hub.LoadContent(owner, name, headRef.Value)
			if err != nil {
				return fmt.Errorf("no active tag — run \"pipe switch %s/%s <tag>\" first", owner, name)
			}
			sha, md5h = hub.ComputeChecksums(content)
			if err := hub.CreateTagSymlink(owner, name, tag, sha); err != nil {
				return fmt.Errorf("creating tag symlink: %w", err)
			}
		}
	} else {
		activeRec, ok := idx.Tags[idx.ActiveTag]
		if !ok {
			return fmt.Errorf("active tag %q not in index", idx.ActiveTag)
		}

		// Read current content from the active tag (follows symlinks)
		var err error
		content, err = hub.LoadContent(owner, name, idx.ActiveTag)
		if err != nil {
			return fmt.Errorf("reading active tag %q: %w", idx.ActiveTag, err)
		}

		// Compute checksum of actual on-disk content
		sha, md5h = hub.ComputeChecksums(content)

		// If active tag is a symlink and content hasn't been modified, we can
		// point the new tag at the same blob. Otherwise write a new blob.
		if sha == activeRec.SHA256 {
			// Clean — reuse same blob via symlink
			if err := hub.CreateTagSymlink(owner, name, tag, sha); err != nil {
				return fmt.Errorf("creating tag symlink: %w", err)
			}
		} else {
			// Dirty — write current content as a new blob and point to it
			newSha, err := hub.WriteBlob(owner, name, content)
			if err != nil {
				return fmt.Errorf("writing blob: %w", err)
			}
			if err := hub.CreateTagSymlink(owner, name, tag, newSha); err != nil {
				return fmt.Errorf("creating tag symlink: %w", err)
			}
			sha = newSha
			_, md5h = hub.ComputeChecksums(content)
		}
	}

	// Add to index
	idx.Tags[tag] = hub.TagRecord{
		SHA256:    sha,
		MD5:       md5h,
		SizeBytes: int64(len(content)),
		CreatedAt: time.Now(),
	}
	if err := hub.SaveIndex(idx); err != nil {
		return fmt.Errorf("saving index: %w", err)
	}

	// Garbage collect orphaned blobs
	if err := hub.GarbageCollectBlobs(owner, name); err != nil {
		log.Warn("garbage collection failed", "err", err)
	}

	log.Info("tagged", "pipe", owner+"/"+name, "tag", tag, "from", sourceLabel, "sha256", sha[:12])
	return nil
}
