package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/anasrar/chihuahua/next/pkg/dat"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

var (
	PathDAT string
	Offset  int64
	Length  int64

	IsRelativeOffset bool
	Filter           string

	Version string
)

var cmd = &cobra.Command{
	Use:     "datls",
	Short:   "List Content From DAT",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		datContainer, err := dat.FromPath(PathDAT, Offset, Length)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range datContainer.Entries {
			name := string(bytes.Trim([]byte(entry.Type), "\x00"))
			if Filter != "" {
				if !fuzzy.Match(Filter, name) {
					continue
				}
			}

			offset := entry.AbsoluteOffset
			if IsRelativeOffset {
				offset = entry.Offset
			}

			fmt.Printf(
				"F %-4s  offset=0x%X length=%d",
				name,
				offset,
				entry.Length,
			)
			fmt.Println()
		}
	},
}

func init() {
	cmd.Flags().StringVarP(&PathDAT, "dat", "i", "", "path to DAT")
	cmd.MarkFlagsOneRequired("dat")

	cmd.Flags().Int64VarP(&Offset, "offset", "s", 0, "offset to DAT")
	cmd.Flags().Int64VarP(&Length, "length", "l", 0, "length of DAT")
	cmd.MarkFlagsRequiredTogether("offset", "length")

	cmd.Flags().BoolVarP(&IsRelativeOffset, "relative", "r", false, "show offset as relative offset")
	cmd.Flags().StringVarP(&Filter, "filter", "f", "", "filter list content of DAT")
}

func main() {
	cmd.Execute()
}
