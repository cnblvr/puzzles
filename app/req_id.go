package app

import (
	"crypto/rand"
	"fmt"
)

func GenerateReqID() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "0000-0000"
	}
	return fmt.Sprintf("%x-%x", buf[0:3], buf[3:6])
}
