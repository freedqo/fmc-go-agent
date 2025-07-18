package uaivectordb

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/components/document"
)

// newDocumentTransformer component initialization function of node 'MarkdownSplitter' in graph 'VectorDb'
func newDocumentTransformer(ctx context.Context) (tfr document.Transformer, err error) {
	// TODO Modify component configuration here.
	config := &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":    "title",      // 一级标题映射为 title
			"##":   "chapter",    // 二级标题映射为 chapter
			"###":  "section",    // 三级标题映射为 section
			"####": "subsection", // 四级标题映射为 subsection
		},
		TrimHeaders: false,
	}
	tfr, err = markdown.NewHeaderSplitter(ctx, config)
	if err != nil {
		return nil, err
	}
	return tfr, nil
}
