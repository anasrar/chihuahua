package tm3

type MetadataEntry struct {
	Source string `json:"source"`
	Name   string `json:"name"`
}

type Metadata struct {
	EntryTotal uint32           `json:"entry_total"`
	Entries    []*MetadataEntry `json:"entries"`
}
