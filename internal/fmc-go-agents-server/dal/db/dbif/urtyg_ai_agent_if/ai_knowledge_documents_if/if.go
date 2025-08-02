package ai_knowledge_documents_if

import "context"

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
	UpdateDocumentsStatus(ctx context.Context, documentsId int64, status int) error
}
