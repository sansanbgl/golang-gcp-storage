package models

import "time"

// CloudStorageFile (optional) can be used if you need to store additional information about uploaded files
type GCPFile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Bucket    string    `json:"bucket"`
	CreatedAt time.Time `json:"created_at"`
}
