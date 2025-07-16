package ujwtx

import (
	"github.com/freedqo/fmc-go-agent/pkg/ujwt"
	"sync"
)

var uJwt ujwt.If
var once sync.Once
var mu sync.Mutex

const jwtKey = "9d6d0c6d2f0p7a4d8c6b0e4caf2d4a7c9c6d0p6d2f0b7a4d8c6l0e4cyf2d4a7c"

func GetJWT() ujwt.If {
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		if uJwt == nil {
			uJwt = ujwt.New("DisplayServer", jwtKey)
		}
	})
	return uJwt
}
