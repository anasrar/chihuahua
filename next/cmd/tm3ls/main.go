package main

import (
	"fmt"
	"log"

	"github.com/anasrar/chihuahua/next/pkg/tm3"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"
)

var (
	PathTM3 string
	Offset  int64
	Length  int64

	IsRelativeOffset bool
	Filter           string

	Version string
)

var cmd = &cobra.Command{
	Use:     "tm3ls",
	Short:   "List Content From TM3",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		tm3Container, err := tm3.FromPath(PathTM3, Offset, Length)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range tm3Container.ToEntries() {
			name := entry.Name
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
				"F %-8s  offset=0x%X length=%d",
				name,
				offset,
				entry.Length,
			)
			fmt.Println()
		}
	},
}

func init() {
	cmd.Flags().StringVarP(&PathTM3, "tm3", "i", "", "path to TM3")
	cmd.MarkFlagsOneRequired("tm3")

	cmd.Flags().Int64VarP(&Offset, "offset", "s", 0, "offset to AFS")
	cmd.Flags().Int64VarP(&Length, "length", "l", 0, "length of AFS")
	cmd.MarkFlagsRequiredTogether("offset", "length")

	cmd.Flags().BoolVarP(&IsRelativeOffset, "relative", "r", false, "show offset as relative offset")
	cmd.Flags().StringVarP(&Filter, "filter", "f", "", "filter list content of TM3")
}

func main() {
	cmd.Execute()
}
