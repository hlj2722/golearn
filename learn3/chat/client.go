package main

import (
	"github.com/gorilla/websocket"
	"time"
)

//client represents a single chatting user
type client struct {
	//socket is the web socket for this client.
	socket *websocket.Conn
	//send is a channel on which message are sent.
	send chan *message
	//room is the room this client is chatting in
	room *room
	//userData holds infomation about the user
	userData map[string]interface{}
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			if avatarUrl, ok := c.userData["avatar_url"]; ok {
				msg.AvatarUrl = avatarUrl.(string)
			}
			//msg.AvatarUrl, _ = c.room.avatar.GetAvatarURL(c)
			c.room.forward <- msg
		} else {
			break
		}
	}

	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
