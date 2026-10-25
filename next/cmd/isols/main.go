package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/anasrar/chihuahua/next/pkg/iso"
	"github.com/spf13/cobra"
)

var (
	PathISO string
	Offset  int64
	Length  int64
)

func walk(n *iso.Node, depth int) {
	kind := "F"
	if n.IsDirectory {
		kind = "D"
	}
	fmt.Printf(
		"%s%s %-16s  lba=%d offset=0x%X length=%d date=%s",
		strings.Repeat("  ", depth),
		kind,
		n.Name,
		n.LBA,
		n.Offset,
		n.Length,
		n.Date.Format("2006-01-02 15:04:05"),
	)
	if n.Error != "" {
		fmt.Printf(" ERROR=%s", n.Error)
	}
	fmt.Println()
	for _, c := range n.Children {
		walk(c, depth+1)
	}
}

var cmd = &cobra.Command{
	Use:   "isols",
	Short: "List Content From PS2 ISO",
	Run: func(cmd *cobra.Command, args []string) {
		isoContainer, err := iso.FromPath(PathISO, Offset, Length)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(
			"Volume: %q SystemID: %q LogicalVolumeID: %q VolumeID: %q FileSetID: %q Sectors: %d\n",
			isoContainer.Image.Volume.Name,
			isoContainer.Image.Volume.SystemID,
			isoContainer.Image.Volume.LogicalVolumeID,
			isoContainer.Image.Volume.VolumeID,
			isoContainer.Image.Volume.FileSetID,
			isoContainer.Image.Sectors,
		)
		walk(isoContainer.Image.Root, 0)
	},
}

func init() {
	cmd.Flags().StringVarP(&PathISO, "iso", "i", "", "path to ISO")
	cmd.MarkFlagsOneRequired("iso")

	cmd.Flags().Int64VarP(&Offset, "offset", "s", 0, "offset to ISO")
	cmd.Flags().Int64VarP(&Length, "length", "l", 0, "length of ISO")
	cmd.MarkFlagsRequiredTogether("offset", "length")
}

func main() {
	cmd.Execute()
}
