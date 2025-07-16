package utils

import (
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tjfoc/gmsm/sm2"
)

// Sm2Encrypt 使用SM2公钥加密数据
func Sm2Encrypt(encryptValue []byte, publicKey string) (string, error) {
	publicKeyObj, err := ParsePublicKey(publicKey)
	if err != nil {
		return "", err
	}
	var encryptByte []byte
	encryptByte, err = sm2.Encrypt(publicKeyObj, encryptValue, nil, sm2.C1C2C3) //Vue前端JS库与后端刚好相反,JS库0:C1C2C3;1:C1C3C2
	if err != nil {
		return "", errors.New(fmt.Sprint("sm2加密内容异常:", err))
	}
	return hex.EncodeToString(encryptByte), nil
}

// ParsePublicKey 解析公钥字符串为PublicKey对象
func ParsePublicKey(publicKey string) (*sm2.PublicKey, error) {
	publicKeyDec, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Hex解码公钥字符串异常:", err))
	}
	curve := sm2.P256Sm2()
	x, y := elliptic.Unmarshal(curve, publicKeyDec)
	publicKeyObj := &sm2.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
	return publicKeyObj, nil
}
