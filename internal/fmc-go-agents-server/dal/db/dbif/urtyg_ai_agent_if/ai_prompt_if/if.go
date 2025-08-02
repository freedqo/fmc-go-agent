package ai_prompt_if

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
}
