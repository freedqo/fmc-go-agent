package uaiagent

import (
	"context"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 定义一个函数，用于构建一个可运行的节点
func (u *UAiAgent) buildRunnable(ctx context.Context, ctp prompt.ChatTemplate, lba *compose.Lambda, rtr retriever.Retriever) (g *compose.Graph[*UserMessage, *schema.Message], r compose.Runnable[*UserMessage, *schema.Message], err error) {
	// 定义一些常量
	const (
		InputToQuery   = "InputToQuery"
		ChatTemplate   = "ChatTemplate"
		ReactAgent     = "ReactAgent"
		RedisRetriever = "RedisRetriever"
		InputToHistory = "InputToHistory"
	)
	// 创建一个新的图
	g = compose.NewGraph[*UserMessage, *schema.Message]()
	// 添加一个Lambda节点
	_ = g.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(newLambda), compose.WithNodeName("UserMessageToQuery"))

	// 添加一个ChatTemplate节点
	_ = g.AddChatTemplateNode(ChatTemplate, ctp)

	// 添加一个Lambda节点
	_ = g.AddLambdaNode(ReactAgent, lba, compose.WithNodeName("ReActAgent"))

	// 添加一个Retriever节点
	_ = g.AddRetrieverNode(RedisRetriever, rtr, compose.WithOutputKey("documents"))
	// 添加一个Lambda节点
	_ = g.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(u.newLambda2), compose.WithNodeName("UserMessageToVariables"))
	// 添加边
	_ = g.AddEdge(compose.START, InputToQuery)
	_ = g.AddEdge(compose.START, InputToHistory)
	_ = g.AddEdge(ReactAgent, compose.END)
	_ = g.AddEdge(InputToQuery, RedisRetriever)
	_ = g.AddEdge(RedisRetriever, ChatTemplate)
	_ = g.AddEdge(InputToHistory, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ReactAgent)
	// 编译图
	r, err = g.Compile(ctx, compose.WithGraphName("UAiAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, nil, err
	}
	return g, r, err
}
