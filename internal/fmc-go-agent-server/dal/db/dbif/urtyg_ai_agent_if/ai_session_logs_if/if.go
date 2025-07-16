package ai_session_logs_if

type If interface {
	Gen() GenIf
	Self() SelfIf
}

type SelfIf interface {
}
