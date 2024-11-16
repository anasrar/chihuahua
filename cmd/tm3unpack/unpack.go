package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/anasrar/chihuahua/pkg/utils"
)

func unpack(
	ctx context.Context,
	datPath string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	tim := tm3.New()
	if err := tm3.FromPath(tim, datPath); err != nil {
		return err
	}

	outputMetadataPath := filepath.Join(utils.ParentDirectory(datPath), fmt.Sprintf("UNPACK_%s", utils.Basename(datPath)), "METADATA.json")

	md := tm3.Metadata{
		EntryTotal: tim.EntryTotal,
		Entries:    []*tm3.MetadataEntry{},
	}

	for i, entry := range tim.Entries {
		normalizeName := utils.FilterUnprintableString(entry.Name)
		source := filepath.Join("FILES", fmt.Sprintf("%s_%03d.tm3", normalizeName, i))
		md.Entries = append(
			md.Entries,
			&tm3.MetadataEntry{
				Source: source,
				Name:   entry.Name,
			},
		)
	}

	outputFilesDirPath := filepath.Join(utils.ParentDirectory(datPath), fmt.Sprintf("UNPACK_%s", utils.Basename(datPath)), "FILES")
	if err := os.MkdirAll(outputFilesDirPath, os.ModePerm); err != nil {
		return err
	}

	if err := tim.Unpack(ctx, outputFilesDirPath, onStart, onDone); err != nil {
		return err
	}

	buf, err := json.MarshalIndent(md, "", "\t")
	if err != nil {
		return err
	}

	metadataFile, err := os.OpenFile(outputMetadataPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer metadataFile.Close()

	if _, err := metadataFile.Write(buf); err != nil {
		return err
	}

	return nil
}
