package user

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/slack-clone-server/config"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

const (
	userStop = 9876128
)

func Edit(client *config.Client, data interface{}) {

	var user User
	user.Id = client.Id
	if err := mapstructure.Decode(data, &user); err != nil {
		client.Send <- config.Message{"error", err.Error()}
	}

	go func() {
		err := r.Table("users").Update(user).Exec(client.Session)

		if err != nil {
			client.Send <- config.Message{"error", err.Error()}
		}
	}()

}

func Subscribe(client *config.Client, data interface{}) {
	stop := client.NewStopChannel(userStop)
	result := make(chan r.ChangeResponse)

	cursor, err := r.Table("users").
		Changes(r.ChangesOpts{IncludeInitial: true}).
		Run(client.Session)

	if err != nil {
		client.Send <- config.Message{"error", err.Error()}
		return
	}

	go func() {
		var change r.ChangeResponse
		for cursor.Next(&change) {
			result <- change
		}
	}()

	go func() {
		for {
			select {
			case <-stop:
				cursor.Close()
				return
			case change := <-result:
				if change.NewValue != nil && change.OldValue == nil {
					client.Send <- config.Message{"user add", change.NewValue}
				} else if change.NewValue != nil && change.OldValue != nil {
					fmt.Println("edit")
					fmt.Println(change.NewValue)
					client.Send <- config.Message{"user edit", change.NewValue}
				}
			}
		}
	}()
}

func Unsubscribe(client *config.Client, data interface{}) {
	client.StopForKey(userStop)
}
