package ai_knowledge_base_if

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
}
