package ai_tool_if

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
}
