package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/anasrar/chihuahua/next/pkg/afs"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

var (
	PathAFS string
	Offset  int64
	Length  int64

	IsRelativeOffset bool
	Filter           string

	Version string
)

var cmd = &cobra.Command{
	Use:     "afsls",
	Short:   "List Content From AFS",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		afsContainer, err := afs.FromPath(PathAFS, Offset, Length)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range afsContainer.Entries {
			name := string(bytes.Trim(entry.Description.Name[:], "\x00"))
			if Filter != "" {
				if !fuzzy.Match(Filter, name) {
					continue
				}
			}

			offset := entry.Offset
			if !IsRelativeOffset {
				offset = offset + uint32(Offset)
			}

			fmt.Printf(
				"F %-32s  offset=0x%X length=%d",
				name,
				offset,
				entry.Length,
			)
			fmt.Println()
		}
	},
}

func init() {
	cmd.Flags().StringVarP(&PathAFS, "afs", "i", "", "path to AFS")
	cmd.MarkFlagsOneRequired("afs")

	cmd.Flags().Int64VarP(&Offset, "offset", "s", 0, "offset to AFS")
	cmd.Flags().Int64VarP(&Length, "length", "l", 0, "length of AFS")
	cmd.MarkFlagsRequiredTogether("offset", "length")

	cmd.Flags().BoolVarP(&IsRelativeOffset, "relative", "r", false, "show offset as relative offset")
	cmd.Flags().StringVarP(&Filter, "filter", "f", "", "filter list content of AFS")
}

func main() {
	cmd.Execute()
}
