package v1m

type ReqCompletions struct {
	// 默认为1 在服务器端生成best_of个补全,并返回“最佳”补全(每个令牌的日志概率最高的那个)。无法流式传输结果。
	// 与n一起使用时,best_of控制候选补全的数量,n指定要返回的数量 – best_of必须大于n。
	// 注意:因为这个参数会生成许多补全,所以它可以快速消耗您的令牌配额。请谨慎使用,并确保您对max_tokens和stop有合理的设置。
	BestOf *int64 `json:"best_of,omitempty"`
	// 默认为false 除了补全之外,还回显提示
	Echo *bool `json:"echo,omitempty"`
	// 默认为0 -2.0和2.0之间的数字。正值根据文本目前的现有频率处罚新令牌,降低模型逐字重复相同行的可能性。
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	// 默认为null 修改完成中指定令牌出现的可能性。
	// 接受一个JSON对象,该对象将令牌(由GPT令牌化器中的令牌ID指定)映射到关联偏差值,-100到100。您可以使用这个令牌化器工具(适用于GPT-2和GPT-3)将文本转换为令牌ID。从数学上讲,偏差在对模型进行采样之前添加到生成的logit中。确切效果因模型而异,但-1至1之间的值应降低或提高选择的可能性;像-100或100这样的值应导致相关令牌的禁用或专属选择。
	// 例如,您可以传递{"50256": -100}来防止生成<|endoftext|>令牌。
	LogitBias map[string]interface{} `json:"logit_bias,omitempty"`
	// 默认为null
	// 包括logprobs个最可能令牌的日志概率,以及所选令牌。例如,如果logprobs为5,API将返回5个最有可能令牌的列表。
	// API总会返回采样令牌的logprob,因此响应中最多可能有logprobs+1个元素。
	//
	// logprobs的最大值是5。
	Logprobs interface{} `json:"logprobs"`
	// 默认为16
	// 在补全中生成的最大令牌数。
	//
	// 提示的令牌计数加上max_tokens不能超过模型的上下文长度。 计数令牌的Python代码示例。
	MaxTokens *int64 `json:"max_tokens,omitempty"`
	// 要使用的模型的 ID。您可以使用[List models](https://platform.openai.com/docs/api-reference/models/list)
	// API 来查看所有可用模型，或查看我们的[模型概述](https://platform.openai.com/docs/models/overview)以了解它们的描述。
	Model string `json:"model"`
	// 默认为1
	// 为每个提示生成的补全数量。
	//
	// 注意:因为这个参数会生成许多补全,所以它可以快速消耗您的令牌配额。请谨慎使用,并确保您对max_tokens和stop有合理的设置。
	N *int64 `json:"n,omitempty"`
	// 默认为0 -2.0和2.0之间的数字。正值根据它们是否出现在目前的文本中来惩罚新令牌,增加模型讨论新话题的可能性。  有关频率和存在惩罚的更多信息,请参阅。
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`
	// 生成完成的提示，编码为字符串、字符串数组、标记数组或标记数组数组。  请注意，<|endoftext|>
	// 是模型在训练期间看到的文档分隔符，因此如果未指定提示，模型将生成新文档的开头。
	Prompt string `json:"prompt"`
	// 如果指定,我们的系统将尽最大努力确定性地进行采样,以便使用相同的种子和参数的重复请求应返回相同的结果。
	// 不保证确定性,您应该参考system_fingerprint响应参数来监视后端的更改。
	Seed *int64 `json:"seed,omitempty"`
	// 默认为null 最多4个序列,API将停止在其中生成更多令牌。返回的文本不会包含停止序列。
	Stop *string `json:"stop,omitempty"`
	// 默认为false 是否流回部分进度。如果设置,令牌将作为可用时发送为仅数据的服务器发送事件,流由数据 Terminated by a data: [DONE] message.
	// 对象消息终止。 Python代码示例。
	Stream *bool `json:"stream,omitempty"`
	// 默认为null 在插入文本的补全之后出现的后缀。
	Suffix *string `json:"suffix,omitempty"`
	// 默认为1 要使用的采样温度,介于0和2之间。更高的值(如0.8)将使输出更随机,而更低的值(如0.2)将使其更集中和确定。  我们通常建议更改这个或top_p,而不是两者都更改。
	Temperature *int64 `json:"temperature,omitempty"`
	// 表示最终用户的唯一标识符,这可以帮助OpenAI监控和检测滥用。 了解更多。
	TopP *int64 `json:"top_p,omitempty"`
	User string `json:"user"`
}

type ResCompletions struct {
	Choices           []Choice         `json:"choices"`
	Created           int64            `json:"created"`
	ID                string           `json:"id"`
	Model             string           `json:"model"`
	Object            string           `json:"object"`
	SystemFingerprint string           `json:"system_fingerprint"`
	Usage             CompletionsUsage `json:"usage"`
}

type Choice struct {
	FinishReason *string     `json:"finish_reason,omitempty"`
	Index        *int64      `json:"index,omitempty"`
	Logprobs     interface{} `json:"logprobs"`
	Text         *string     `json:"text,omitempty"`
}

type CompletionsUsage struct {
	CompletionTokens int64 `json:"completion_tokens"`
	PromptTokens     int64 `json:"prompt_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}
