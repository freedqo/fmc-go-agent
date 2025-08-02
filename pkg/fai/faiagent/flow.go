package faiagent

import (
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// newLambda1 component initialization function of node 'ReactAgent' in graph 'FAiAgent'
func (u *FAiAgent) newLambda1(ins *react.Agent) (lba *compose.Lambda, err error) {

	// 创建一个Lambda实例
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}
