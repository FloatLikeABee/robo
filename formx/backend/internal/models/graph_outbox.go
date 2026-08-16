package models

import "time"

// GraphSyncOutbox queues Neo4j / GraphRAG sync jobs after primary writes.
type GraphSyncOutbox struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Source     string    `json:"source" gorm:"size:64;not null"`
	EntityType string    `json:"entity_type" gorm:"size:64;not null"`
	EntityID   string    `json:"entity_id" gorm:"size:128;not null"`
	Op         string    `json:"op" gorm:"size:32;not null"`
	CreatedAt  time.Time `json:"created_at"`
}

func (GraphSyncOutbox) TableName() string { return "graph_sync_outbox" }
