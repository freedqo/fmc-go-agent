package iconsts

import (
	"github.com/freedqo/fmc-go-agent/internal/fmc-go-agent-server/model/dalm/dbm/urtyg_ai_agent/model"
	"time"
)

// 会话类型（提示词模板）固化ID
type PromptID string

const (
	PromptID_IntelligentAssistant PromptID = "PromptID_IntelligentAssistant-001" // 智能助手固化模板ID
	PromptID_DrawAssistant        PromptID = "PromptID_DrawAssistant-002"        // 图表绘制助手固化模板ID
)

type PromptType string

const ()

var PromptTypeDist = map[PromptType]string{}

var PromptDist = map[PromptID]model.Ai_prompt{
	PromptID_IntelligentAssistant: {
		ID:          string(PromptID_IntelligentAssistant),
		UserID:      "1",
		Type:        "BaseAssistant",
		Name:        "基础智能助手",
		Description: "智能助手基础提示模板",
		Content:     systemPrompt,
		IsShared:    true,
		SharedAt:    time.Date(2025, 7, 16, 0, 0, 0, 0, time.Local),
	},
	PromptID_DrawAssistant: {
		ID:          string(PromptID_DrawAssistant),
		UserID:      "1",
		Type:        "BaseDraw",
		Name:        "基础图表绘制助手",
		Description: "图表绘制助手基础提示模板",
		Content:     "智能助手",
		IsShared:    true,
		SharedAt:    time.Date(2025, 7, 16, 0, 0, 0, 0, time.Local),
	},
}

var systemPrompt = `
# 角色定义: 
- 你是小U,一个产品AI智能专家, 你在回答用户的问题的时候, 会提供准确、简洁、有用的答案,你在执行用户提出的动作的时候,能准确、快速、稳定的执行,并返回执行结果。

## 核心能力
- 电力系统知识
- 电力系统优化
- 电力系统调度
- 电力系统安全
- 电力系统故障诊断
- 电力系统维护

## 工具调用指南
 你(大模型)需要调用工具,来执行用户提出的问题的时候,不能输出文字思考,不能暴露具体的参数名称和值，而是直接调用（二次确认除外）

## 互动指南
- 再提供帮助的时候的指南:
  你要用词恰当,不卑不亢,落落大方,不能使用抹黑的内容对用户进行任何的评价,你必须使用符合企业核心价值观的内容进行正面回答。
- 如果某个请求超出了你的能力范围:
  你要明确告知用户,并解释原因,并提供替代方案,给出你具备的核心能力。
`
