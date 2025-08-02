package fvectordb

import (
	"context"
	"github.com/freedqo/fmc-go-agents/pkg/fai/fvectordb/entity"
)

type DocDb struct {
	UpdateDocumentsStatusFunc func(ctx context.Context, documentsId int64, status int) error
	SaveChunksDataFunc        func(ctx context.Context, documentsId int64, chunks []entity.KnowledgeChunks) error
}

func (d *DocDb) UpdateDocumentsStatus(ctx context.Context, documentsId int64, status int) error {
	return d.UpdateDocumentsStatusFunc(ctx, documentsId, status)
}

func (d *DocDb) SaveChunksData(ctx context.Context, documentsId int64, chunks []entity.KnowledgeChunks) error {
	return d.SaveChunksDataFunc(ctx, documentsId, chunks)
}

var _ DocDbIf = &DocDb{}

func NewDocDbIf(UpdateDocumentsStatusFunc func(ctx context.Context, documentsId int64, status int) error, SaveChunksDataFunc func(ctx context.Context, documentsId int64, chunks []entity.KnowledgeChunks) error) DocDbIf {
	d := &DocDb{
		UpdateDocumentsStatusFunc: UpdateDocumentsStatusFunc,
		SaveChunksDataFunc:        SaveChunksDataFunc,
	}
	return d
}
