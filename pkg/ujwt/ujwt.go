package ujwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

// New 创建一个UJwt实例
// 入参：name string 签发者
// 入参：jwtKey string  签名密钥
// 返回：If  UJwt实例
func New(name, jwtKey string) If {
	return &UJwt{
		Name:   name,
		jwtKey: []byte(jwtKey),
	}
}

type UJwt struct {
	Name   string
	jwtKey []byte
	jwt    jwt.RegisteredClaims
}

// UClaims 自定义声明结构体，包含用户名、用户id、过期时间
type UClaims struct {
	UserName string `json:"userName"`
	UserId   string `json:"userId"`
	Option   string `json:"option"`
	jwt.RegisteredClaims
}

// GenToken 生成token
// 入参：userName string 用户名
// 入参：userId string 用户id
// 入参：effectiveDuration int token有效时长(分钟)
// 返回：string token
func (j *UJwt) GenToken(userName, userId string, effectiveDuration int) (string, error) {
	if strings.TrimSpace(userName) == "" || strings.TrimSpace(userId) == "" {
		return "", errors.New("用户名和用户Id不能为空")
	}
	if effectiveDuration <= 0 {
		return "", errors.New("token有效时长必须大于0(分钟)")
	}
	// 创建自定义声明
	now := time.Now()
	claims := &UClaims{
		UserName: userName,
		UserId:   userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Name,
			Subject:   userId,
			Audience:  nil,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(effectiveDuration) * time.Minute)), // 设置过期时间为 effectiveDuration 分钟后
			NotBefore: nil,
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        "",
		},
	}
	// 使用 HS256 算法创建令牌对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 对令牌进行签名
	tokenString, err := token.SignedString(j.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken 验证token
// 入参：token string token
// 返回：*UClaims 自定义声明结构体
// 返回：error 错误信息
func (j *UJwt) VerifyToken(token string) (*UClaims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("token不能为空")
	}
	// 解析令牌
	claims := &UClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtKey, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrInvalidKey):
			return nil, errors.New("token 键无效")
		case errors.Is(err, jwt.ErrInvalidKeyType):
			return nil, errors.New("token 键类型无效")
		case errors.Is(err, jwt.ErrHashUnavailable):
			return nil, errors.New("token 请求的哈希函数不可用")
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, errors.New("token 令牌格式错误")
		case errors.Is(err, jwt.ErrTokenUnverifiable):
			return nil, errors.New("token 令牌不可验证")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, errors.New("token 令牌签名无效")
		case errors.Is(err, jwt.ErrTokenRequiredClaimMissing):
			return nil, errors.New("token 令牌缺少必需的声明")
		case errors.Is(err, jwt.ErrTokenInvalidAudience):
			return nil, errors.New("token 令牌的受众无效")
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, errors.New("token 令牌已过期")
		case errors.Is(err, jwt.ErrTokenUsedBeforeIssued):
			return nil, errors.New("token 令牌在发布之前被使用")
		case errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, errors.New("token 令牌的发行者无效")
		case errors.Is(err, jwt.ErrTokenInvalidSubject):
			return nil, errors.New("token 令牌的主题无效")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, errors.New("token 令牌尚未生效")
		case errors.Is(err, jwt.ErrTokenInvalidId):
			return nil, errors.New("token 令牌的 ID 无效")
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			return nil, errors.New("token 令牌的声明无效")
		case errors.Is(err, jwt.ErrInvalidType):
			return nil, errors.New("声明的类型无效")
		default:
			{
				return nil, fmt.Errorf("token 校验失败:%w", err)
			}
		}
	}
	if tkn == nil {
		return nil, errors.New("token 无效")
	}
	if !tkn.Valid {
		return nil, errors.New("token 无效")
	} else {
		return claims, nil
	}
}
func (j *UJwt) GetSession(userId, sessionId string) (*string, error) {
	if strings.TrimSpace(userId) == "" {
		return nil, errors.New("用户名和用户Id不能为空")
	}
	// 创建自定义声明
	now := time.Now()
	claims := &UClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Name,
			Subject:   userId, //+ "|" + sessionId,
			Audience:  nil,
			ExpiresAt: nil, // 设置永不过期
			NotBefore: nil,
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        sessionId,
		},
	}
	// 使用 HS256 算法创建令牌对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 对令牌进行签名
	tokenString, err := token.SignedString(j.jwtKey)
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}

func (j *UJwt) VerifySession(sessionID string) (*UClaims, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("token不能为空")
	}
	// 解析令牌
	claims := &UClaims{}
	tkn, err := jwt.ParseWithClaims(sessionID, claims, func(token *jwt.Token) (interface{}, error) {
		return j.jwtKey, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrInvalidKey):
			return nil, errors.New("session键无效")
		case errors.Is(err, jwt.ErrInvalidKeyType):
			return nil, errors.New("session键类型无效")
		case errors.Is(err, jwt.ErrHashUnavailable):
			return nil, errors.New("session请求的哈希函数不可用")
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, errors.New("session令牌格式错误")
		case errors.Is(err, jwt.ErrTokenUnverifiable):
			return nil, errors.New("session令牌不可验证")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, errors.New("session令牌签名无效")
		case errors.Is(err, jwt.ErrTokenRequiredClaimMissing):
			return nil, errors.New("session令牌缺少必需的声明")
		case errors.Is(err, jwt.ErrTokenInvalidAudience):
			return nil, errors.New("session令牌的受众无效")
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, errors.New("session令牌已过期")
		case errors.Is(err, jwt.ErrTokenUsedBeforeIssued):
			return nil, errors.New("session令牌在发布之前被使用")
		case errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, errors.New("session令牌的发行者无效")
		case errors.Is(err, jwt.ErrTokenInvalidSubject):
			return nil, errors.New("session令牌的主题无效")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, errors.New("session令牌尚未生效")
		case errors.Is(err, jwt.ErrTokenInvalidId):
			return nil, errors.New("session令牌的 ID 无效")
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			return nil, errors.New("session令牌的声明无效")
		case errors.Is(err, jwt.ErrInvalidType):
			return nil, errors.New("声明的类型无效")
		default:
			{
				return nil, fmt.Errorf("session校验失败:%w", err)
			}
		}
	}
	if tkn == nil {
		return nil, errors.New("session无效")
	}
	if !tkn.Valid {
		return nil, errors.New("session无效")
	} else {
		return claims, nil
	}
}

var _ If = &UJwt{}
