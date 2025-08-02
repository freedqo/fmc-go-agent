package ujwtx

import (
	"github.com/freedqo/fmc-go-agents/pkg/fjwt"
	"sync"
)

var uJwt fjwt.If
var once sync.Once
var mu sync.Mutex

const jwtKey = "9d6d0c6d2f0p7a4d8c6b0e4caf2d4a7c9c6d0p6d2f0b7a4d8c6l0e4cyf2d4a7c"

func GetJWT() fjwt.If {
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		if uJwt == nil {
			uJwt = fjwt.New("DisplayServer", jwtKey)
		}
	})
	return uJwt
}
