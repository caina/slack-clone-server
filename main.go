package main

import (
	"fmt"
	"github.com/slack-clone-server/config"
	"github.com/slack-clone-server/handlers/channel"
	"github.com/slack-clone-server/handlers/message"
	"github.com/slack-clone-server/handlers/user"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
	"log"
	"net/http"
)

type User struct {
	Id   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}

func main() {
	session, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:32772",
		Database: "rtsupport",
	})

	if err != nil {
		log.Panic("Connection problem", err.Error())
	}

	router := config.NewRouter(session)

	router.Handle("channel add", channel.Add)
	router.Handle("channel subscribe", channel.Subscribe)
	router.Handle("channel unsubscribe", channel.Unsubscribe)

	router.Handle("user edit", user.Edit)
	router.Handle("user subscribe", user.Subscribe)
	router.Handle("user unsubscribe", user.Unsubscribe)

	router.Handle("message add", message.Add)
	router.Handle("message subscribe", message.Subscribe)
	router.Handle("message unsubscribe", message.Unsubscrive)

	http.Handle("/", router)
	if err := http.ListenAndServe(":4000", nil); err != nil {
		fmt.Println(err.Error())
	}
}
