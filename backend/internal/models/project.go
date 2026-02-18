package models

import "time"

type Project struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	GitURL            *string    `json:"git_url,omitempty"`
	GitBranch         *string    `json:"git_branch,omitempty"`
	GitTokenEncrypted *string    `json:"-"`
	LastSyncedAt      *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}
