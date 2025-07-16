package uaivectordb

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/compose"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"io/fs"
	"path/filepath"
	"strings"
)

func New(ctx context.Context, option *Option, log *zap.SugaredLogger) If {
	irv := &IRVector{
		log: log,
		opt: option,
		rdb: redis.NewClient(&redis.Options{
			Addr:     option.RedisStack.Addr,
			Protocol: option.RedisStack.Protocol,
			DB:       option.RedisStack.Db,
		}),
	}

	eb, err := newEmbedding(ctx, option)
	if err != nil {
		irv.log.Fatalw("Failed to create embedding", "error", err)
	}
	irv.eb = eb
	irv.ldr, err = newLoader(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create loader: %v", err))
	}
	irv.idr, err = newIndexer(ctx, option, irv.rdb)
	if err != nil {
		panic(fmt.Sprintf("Failed to create indexer: %v", err))
	}
	irv.tfr, err = newDocumentTransformer(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to create document transformer: %v", err))
	}
	r, err := irv.buildIRVector(ctx, irv.opt, irv.ldr, irv.idr, irv.tfr)
	if err != nil {
		irv.log.Fatalw("Failed to build VectorDb", "error", err)
	}
	irv.r = r
	if option.LoadMdFilePloy.IsLoadMdFiles {
		err = irv.BuildDir(ctx, option.LoadMdFilePloy.Dir)
		if err != nil {
			irv.log.Fatalw("Failed to build local knowledge db", "error", err)
		}
	}
	return irv
}

type IRVector struct {
	log *zap.SugaredLogger
	opt *Option
	rdb *redis.Client
	eb  embedding.Embedder
	ldr document.Loader
	idr indexer.Indexer
	tfr document.Transformer
	r   compose.Runnable[document.Source, []string]
}

func (i *IRVector) Redis() *redis.Client {
	return i.rdb
}

func (i *IRVector) Embedder() embedding.Embedder {
	return i.eb
}

func (i *IRVector) Runnable() compose.Runnable[document.Source, []string] {
	return i.r
}

func (i *IRVector) BuildDir(ctx context.Context, dir string) (err error) {
	i.log.Infow("Starting to build local knowledge db", "directory", dir)
	// 遍历 dir 下的所有 markdown 文件
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			i.log.Errorw("Failed to walk directory", "error", err)
			return fmt.Errorf("walk dir failed: %w", err)
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			i.log.Infow("Skipping non - markdown file", "file", path)
			return nil
		}

		i.log.Infow("Starting to index file", "file", path)

		ids, err := i.r.Invoke(ctx, document.Source{URI: path})
		if err != nil {
			i.log.Errorw("Failed to invoke index graph", "error", err)
			return fmt.Errorf("invoke index graph failed: %w", err)
		}

		i.log.Infow("Finished indexing file", "file", path, "number of parts", len(ids))

		return nil
	})
	if err != nil {
		i.log.Errorw("Failed to build local knowledge db", "error", err)
		return fmt.Errorf("build local knowledge db failed: %w", err)
	}
	i.log.Infow("Successfully build local knowledge db")
	return nil
}
func (i *IRVector) BuildFile(ctx context.Context, filepath string) (err error) {
	i.log.Infow("Starting to build local knowledge db", "file path", filepath)
	if !strings.HasSuffix(filepath, ".md") {
		i.log.Infow("Skipping non - markdown file", "file", filepath)
		return nil
	}
	ids, err := i.r.Invoke(ctx, document.Source{URI: filepath})
	if err != nil {
		i.log.Errorw("Failed to invoke index graph", "error", err)
		return fmt.Errorf("invoke index graph failed: %w", err)
	}

	i.log.Infow("Finished indexing file", "file", filepath, "number of parts", len(ids))
	return nil
}
