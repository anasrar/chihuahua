package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/anasrar/chihuahua/next/pkg/afs"
	"github.com/spf13/cobra"
)

var (
	PathAFS string
	Offset  int64
	Length  int64
)

var cmd = &cobra.Command{
	Use:   "afsls",
	Short: "List Content From AFS",
	Run: func(cmd *cobra.Command, args []string) {
		afsContainer, err := afs.FromPath(PathAFS, Offset, Length)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range afsContainer.Entries {
			fmt.Printf(
				"F %-32s  offset=0x%X length=%d",
				bytes.Trim(entry.Description.Name[:], "\x00"),
				entry.Offset,
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
}

func main() {
	cmd.Execute()
}
