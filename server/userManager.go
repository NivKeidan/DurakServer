package server

import (
	"DurakGo/output"
	"fmt"
	"math/rand"
)

type UserManager struct {
	users []*User
	notAliveChan chan *User
	ttl int
}


func NewUserManager(ttl int) *UserManager {
	return &UserManager{
		notAliveChan: make(chan *User),
		users: make([]*User, 0),
		ttl: ttl,
	}
}


func (this *UserManager) CreateNewUser() *User {
	u := &User{connectionId: this.createUserIdentificationString(), notAliveChan: this.notAliveChan,
		isJoined: false, gameChan:nil, appChan: nil}
	output.Spit(fmt.Sprintf("New User Created: %s", u))
	u.receivedAlive()
	go u.checkIsAlive(this.ttl)
	return u
}

func (this *UserManager) doesCodeExist(c string) bool {
	if c == "" {  // empty string always exists
		return true
	}

	for _, u := range this.users {
		if c == u.connectionId {
			return true
		}
	}

	return false
}

func (this *UserManager) createUserIdentificationString() string {
	letters := configuration.GetString("ClientIdLetters")
	length := configuration.GetInt("ClientIdLength")
	b := make([]byte, length)
	var s string
	for this.doesCodeExist(s) {
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		s = string(b)
	}
	return s
}

func (this *UserManager) GetUserByConnectionId(connId string) *User {
	for _, u := range this.users {
		if u.connectionId == connId {
			return u
		}
	}
	return nil
}