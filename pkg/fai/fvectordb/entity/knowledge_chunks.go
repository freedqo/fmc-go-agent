// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// KnowledgeChunks is the golang structure for table knowledge_chunks.
type KnowledgeChunks struct {
	Id             int64      `json:"id"             orm:"id"               description:""` //
	KnowledgeDocId int64      `json:"knowledgeDocId" orm:"knowledge_doc_id" description:""` //
	ChunkId        string     `json:"chunkId"        orm:"chunk_id"         description:""` //
	Content        string     `json:"content"        orm:"content"          description:""` //
	Ext            string     `json:"ext"            orm:"ext"              description:""` //
	Status         int        `json:"status"         orm:"status"           description:""` //
	CreatedAt      *time.Time `json:"createdAt"      orm:"created_at"       description:""` //
	UpdatedAt      *time.Time `json:"updatedAt"      orm:"updated_at"       description:""` //
}
