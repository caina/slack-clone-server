package channel

import (
	"github.com/mitchellh/mapstructure"
	"github.com/slack-clone-server/config"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// iota vai adicionar os valores: 1,2,3 automaticamente
const (
	ChannelStop = iota
	UserStop
	MessageStop
)

func Add(client *config.Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.Send <- config.Message{"error", err.Error()}
	}

	go func() {
		err = r.Table("channel").
			Insert(channel).
			Exec(client.Session)

		if err != nil {
			client.Send <- config.Message{"error", err.Error()}
		}
	}()
}

func Subscribe(client *config.Client, data interface{}) {
	stop := client.NewStopChannel(ChannelStop)
	result := make(chan r.ChangeResponse)

	cursor, err := r.Table("channel").
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
					client.Send <- config.Message{"channel add", change.NewValue}
				}
			}
		}
	}()
}

func Unsubscribe(client *config.Client, data interface{}) {
	client.StopForKey(ChannelStop)
}
