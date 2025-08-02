package ai_mcp_if

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
}
