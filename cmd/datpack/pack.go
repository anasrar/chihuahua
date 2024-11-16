package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/anasrar/chihuahua/pkg/utils"
)

func pack(
	ctx context.Context,
	metadataPath string,
	onStart,
	onDone func(total uint32, current uint32, name string),
) error {
	metadataBuf, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var m dat.Metadata
	if err := json.Unmarshal(metadataBuf, &m); err != nil {
		return err
	}

	d := dat.New()

	parentDir := utils.ParentDirectory(metadataPath)

	for _, entry := range m.Entries {
		if entry.IsNull {
			d.AddNullEntry()
		} else {
			if err := d.AddEntryFromPathWithType(
				filepath.Join(
					parentDir, entry.Source,
				),
				entry.Type,
			); err != nil {
				return err
			}
		}
	}

	if err := d.Pack(
		ctx,
		filepath.Join(parentDir, "OUTPUT.dat"),
		onStart,
		onDone,
	); err != nil {
		return err
	}

	return nil
}
