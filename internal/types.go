package internal

import (
	"time"
)

type StateEntry struct {
	Type        string     `json:"type"`
	Filename    string     `json:"filename"`
	Description string     `json:"description,omitempty"`
	MD5SUM      string     `json:"filesum,omitempty"`
	ChainSum    string     `json:"chainsum,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
}

type DatabaseState struct {
	IndexLocation  string       `json:"index"`
	Date           time.Time    `json:"date"`
	CurrentVersion int          `json:"version"`
	Entries        []StateEntry `json:"entries"`
}

type IndexState struct {
	IndexFile string       `json:"indexfile"`
	Entries   []StateEntry `json:"entries"`
}
