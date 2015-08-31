package main

import (
	_ "fmt"
	"github.com/gorilla/websocket"
	"github.com/lfq618/golearn/learn3/trace"
	"github.com/stretchr/objx"
	"log"
	"net/http"
)

type room struct {
	//forward is a channel that holds incoming messages
	//that should be forwarded to the other clients.
	forward chan *message

	//join is a channel for clients wishing to join the room
	join chan *client

	//leave is a channel for clients wishing to leave the room
	leave chan *client

	//clients holds all current clients in this room
	clients map[*client]bool

	//tracer will receive trace infomation of activity
	//in the room.
	tracer trace.Tracer

	//avatar is how avatar infomation will be obtained.
	avatar Avatar
}

//newRoom makes a new room that is ready to go.
func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			//joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			//leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", msg.Message)
			//forward message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					//send the message
					r.tracer.Trace(" -- sent to client")
				default:
					//failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- failed to send, cleaned up client.")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
