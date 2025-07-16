package uaivectordb

import (
	"context"
	"github.com/cloudwego/eino/components/indexer"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

func (i *IRVector) buildIRVector(ctx context.Context, opt *Option, fileLoaderKeyOfLoader document.Loader, idr indexer.Indexer, tfr document.Transformer) (r compose.Runnable[document.Source, []string], err error) {
	const (
		FileLoader       = "FileLoader"
		MarkdownSplitter = "MarkdownSplitter"
		RedisIndexer     = "RedisIndexer"
	)
	// 初始化RedisStack Vector数据库
	err = initRedisIndex(ctx, opt, i.rdb)
	if err != nil {
		return nil, err
	}
	g := compose.NewGraph[document.Source, []string]()
	// 添加FileLoader节点
	_ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)
	// 添加MarkdownSplitter节点
	_ = g.AddDocumentTransformerNode(MarkdownSplitter, tfr)
	// 添加RedisIndexer节点
	_ = g.AddIndexerNode(RedisIndexer, idr)
	// 添加节点之间的边
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, MarkdownSplitter)
	_ = g.AddEdge(MarkdownSplitter, RedisIndexer)
	_ = g.AddEdge(RedisIndexer, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("VectorDb"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
