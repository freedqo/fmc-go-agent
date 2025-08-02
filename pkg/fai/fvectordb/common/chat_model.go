package common

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

var (
	embeddingModel model.BaseChatModel
	rerankModel    model.BaseChatModel
	rewriteModel   model.BaseChatModel
	qaModel        model.BaseChatModel
)

func GetRewriteModel(ctx context.Context, cfg *openai.ChatModelConfig) (model.BaseChatModel, error) {
	if rewriteModel != nil {
		return rewriteModel, nil
	}
	if cfg == nil {
		panic("qa model config is nil")
	}
	cm, err := openai.NewChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}
	rewriteModel = cm
	return cm, nil
}

func GetQAModel(ctx context.Context, cfg *openai.ChatModelConfig) (model.BaseChatModel, error) {
	if qaModel != nil {
		return qaModel, nil
	}
	if cfg == nil {
		panic("qa model config is nil")
	}
	cm, err := openai.NewChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}
	qaModel = cm
	return cm, nil
}
