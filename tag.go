package main

import (
	"time"
)

// Tag of a repository tagged via GitHub.
type Tag struct {
	ID            string
	Name          string
	CommittedDate time.Time
}
