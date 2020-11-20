package server

import (
	"DurakGo/output"
	"DurakGo/server/httpPayloadTypes"
	"fmt"
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



func (this *User) String() string {
	if this.name != "" {
		return fmt.Sprintf("%s(%s)", this.name, this.connectionId)
	} else {
		return fmt.Sprintf("%s", this.connectionId)
	}
}



