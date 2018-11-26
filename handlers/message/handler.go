package message

import (
	"github.com/mitchellh/mapstructure"
	"github.com/slack-clone-server/config"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
	"time"
)

const (
	channelStop = 9871265
)

func Add(client *config.Client, data interface{}) {
	var posting Posting
	posting.CreatedAt = time.Now()
	posting.Author = client.UserName

	err := mapstructure.Decode(data, &posting)
	if err != nil {
		client.Send <- config.Message{"error", err.Error()}
	}

	go func() {
		err := r.Table("posting").
			Insert(posting).
			Exec(client.Session)

		if err != nil {
			client.Send <- config.Message{"error", err.Error()}
		}
	}()
}

func Subscribe(client *config.Client, data interface{}) {
	stop := client.NewStopChannel(channelStop)
	result := make(chan r.ChangeResponse)

	var messageFilter MessageFilter
	if err := mapstructure.Decode(data, &messageFilter); err != nil {
		client.Send <- config.Message{"error", err.Error()}
	}

	cursor, err := r.Table("posting").
		OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}). //precisa de um index na tabela pra isso funcionar!
		Filter(messageFilter).
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
					client.Send <- config.Message{"message add", change.NewValue}
				}
			}
		}
	}()

}

func Unsubscrive(client *config.Client, data interface{}) {
	client.StopForKey(channelStop)
}
