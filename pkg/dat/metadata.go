package dat

type MetadataEntry struct {
	IsNull bool   `json:"is_null"`
	Source string `json:"source"`
	Type   string `json:"type"`
}

type Metadata struct {
	EntryTotal uint32           `json:"entry_total"`
	Entries    []*MetadataEntry `json:"entries"`
}
