package config

import (
	"github.com/gorilla/websocket"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
	"log"
)

type FindHandler func(string) (Handler, bool)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Client struct {
	Send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	Session      *r.Session
	stopChannels map[int]chan bool
	Id           string
	UserName     string
}

func (c *Client) NewStopChannel(stopKey int) chan bool {
	c.StopForKey(stopKey)
	stop := make(chan bool)
	c.stopChannels[stopKey] = stop
	return stop
}

func (c *Client) StopForKey(key int) {
	if ch, found := c.stopChannels[key]; found {
		ch <- true
		delete(c.stopChannels, key)
	}
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}

		if handler, found := client.findHandler(message.Name); found {
			handler(client, message.Data)
		}
	}

	client.socket.Close()
}

func (client *Client) Write() {
	for msg := range client.Send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	client.socket.Close()
}

func (c *Client) Close() {
	for _, ch := range c.stopChannels {
		ch <- true
	}
	close(c.Send)
}

type User struct {
	Id   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {

	var userModel User
	userModel.Name = "Anonymous"
	res, err := r.Table("users").Insert(userModel).RunWrite(session)
	if err != nil {
		log.Printf(err.Error())
	}

	if len(res.GeneratedKeys) > 0 {
		userModel.Id = res.GeneratedKeys[0]
	}

	return &Client{
		Send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		Session:      session,
		stopChannels: make(map[int]chan bool),
		Id:           userModel.Id,
		UserName:     userModel.Name,
	}
}
