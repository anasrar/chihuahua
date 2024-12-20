package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/utils"
)

func unpack(
	ctx context.Context,
	datPath string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	d := dat.New()
	if err := dat.FromPath(d, datPath); err != nil {
		return err
	}

	outputMetadataPath := filepath.Join(
		utils.ParentDirectory(datPath),
		fmt.Sprintf("UNPACK_%s", utils.Basename(datPath)),
		"METADATA.json",
	)

	md := dat.Metadata{
		EntryTotal: d.EntryTotal,
		Entries:    []*dat.MetadataEntry{},
	}

	for i, entry := range d.Entries {
		if entry.IsNull {
			md.Entries = append(
				md.Entries,
				&dat.MetadataEntry{
					IsNull: true,
					Source: "",
					Type:   "\x00\x00\x00\x00",
				},
			)
		} else {
			normalizeType := utils.FilterUnprintableString(entry.Type)
			source := filepath.Join(
				"FILES",
				normalizeType,
				fmt.Sprintf("%s_%03d.%s", normalizeType, i, strings.ToLower(normalizeType)),
			)
			md.Entries = append(
				md.Entries,
				&dat.MetadataEntry{
					IsNull: false,
					Source: source,
					Type:   entry.Type,
				},
			)
		}
	}

	outputFilesDirPath := filepath.Join(
		utils.ParentDirectory(datPath),
		fmt.Sprintf("UNPACK_%s", utils.Basename(datPath)),
		"FILES",
	)
	if err := os.MkdirAll(outputFilesDirPath, os.ModePerm); err != nil {
		return err
	}

	if err := d.Unpack(ctx, outputFilesDirPath, onStart, onDone); err != nil {
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
