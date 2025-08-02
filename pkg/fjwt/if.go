package fjwt

type If interface {
	GenToken(userName, userId string, effectiveDuration int) (string, error)
	VerifyToken(token string) (*UClaims, error)
	GetSession(userId, sessionId string) (*string, error)
	VerifySession(sessionID string) (*UClaims, error)
}
