package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/anasrar/chihuahua/pkg/tm3"
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

	var m tm3.Metadata
	if err := json.Unmarshal(metadataBuf, &m); err != nil {
		return err
	}

	tim := tm3.New()

	parentDir := utils.ParentDirectory(metadataPath)

	for _, entry := range m.Entries {
		if err := tim.AddEntryFromPathWithName(
			filepath.Join(parentDir, entry.Source),
			entry.Name,
		); err != nil {
			return err
		}
	}

	if err := tim.Pack(
		ctx,
		filepath.Join(parentDir, "OUTPUT.tm3"),
		onStart,
		onDone,
	); err != nil {
		return err
	}

	return nil
}
