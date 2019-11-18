package server

import (
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"fmt"
	"math/rand"
	"time"
)

type User struct {
	connectionId string
	gameChan     chan httpPayloadTypes.JSONResponseData
	appChan      chan httpPayloadTypes.JSONResponseData
	name         string
	lastAlive    int64
	notAliveChan chan *User
	isJoined     bool
}

func NewUser(name string, ttl int, notAliveChan chan *User) *User {
	u := &User{connectionId: createPlayerIdentificationString(), name: name, notAliveChan: notAliveChan,
		isJoined: false, gameChan:nil, appChan: nil}
	u.receivedAlive()
	go u.checkIsAlive(ttl)
	return u
}

func (this *User) receivedAlive() {
	this.lastAlive = time.Now().Unix()
}

func (this *User) checkIsAlive(ttl int) {
	output.Spit(fmt.Sprintf("go routine - monitoring if user %s is alive - start", this))
	defer func() {
		output.Spit(fmt.Sprintf("go routine - monitoring if user %s is alive - ended", this))
	}()
	for {
		now := time.Now().Unix()
		if now - this.lastAlive > int64(ttl) {
			this.notAliveChan <- this
			return
		}
		time.Sleep(time.Duration(ttl) * time.Second)
	}
}

func createPlayerIdentificationString() string {
	letters := configuration.GetString("ClientIdLetters")
	length := configuration.GetInt("ClientIdLength")
	b := make([]byte, length)
	var s string
	for doesCodeExist(s) {
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		s = string(b)
	}
	return s
}

func (this *User) String() string {
	if this.name != "" {
		return fmt.Sprintf("%s(%s)", this.name, this.connectionId)
	} else {
		return fmt.Sprintf("%s", this.connectionId)
	}
}

func doesCodeExist(c string) bool {
	// This func is called in a loop, so first call should return true
	if c == "" {
		return true
	}

	for _, u := range users {
		if c == u.connectionId {
			return true
		}
	}

	return false
}