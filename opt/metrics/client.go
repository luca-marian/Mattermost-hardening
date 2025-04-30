package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"context"
	"log"
	"time"
	"encoding/json"
)

// api client

type ApiClient struct {
	Url string
	Token string
	Team string
	Client4 *model.Client4
	
	// values set in Setup method
	RetryCount int
	RetryDuration time.Duration
}

// ws client

type WSClient struct {
	Url string
	Token string
	Client4 *model.WebSocketClient
}


func (client *ApiClient) GetChannel(channelName string) (*model.Channel, error) {
    etag := ""
    channel, _, err := client.Client4.GetChannelByNameForTeamName(context.Background(), channelName, client.Team, etag)
    if err != nil {
        log.Printf("Failed to get channel: %v\n", err)
    }

	for retry := 0; retry < client.RetryCount; retry++ {
		sleep_duration := client.RetryDuration
		channel, _, err = client.Client4.GetChannelByNameForTeamName(context.Background(), channelName, client.Team, etag)
		if err != nil {
			log.Printf("Failed to get channel: %v. Sleeping %v\n", err, sleep_duration)
			time.Sleep(sleep_duration)
		}
	}

	if err == nil {
    	log.Print("Found channel ", channel.Id," with name ", channel.Name,"\n")
	}

	return channel, err
} 

func (client *ApiClient) Setup() {
	client.Client4 = model.NewAPIv4Client(client.Url)
    client.Client4.SetToken(client.Token)
	
	client.RetryCount = 10
	client.RetryDuration = 1 * time.Minute
}

func (client *ApiClient) CreatePost(post *model.Post) (*model.Post, *model.Response, error) {
	return client.Client4.CreatePost(context.Background(), post)
}


func (client *WSClient) Connect() (err error) {
	client.Client4, err = model.NewWebSocketClient4(client.Url, client.Token)

	for err != nil {
		log.Printf("Failed to connect to WebSocket: %v", err)
		time.Sleep(1 * time.Minute)
		client.Client4, err = model.NewWebSocketClient4(client.Url, client.Token)
	}
	
	return err
}

func (client *WSClient) Listen() error {
	client.Client4.Listen()
	// TODO: check for listen errors
	
	return nil
}

func (client *WSClient) Close() {
	client.Client4.Close()
}

func (client *WSClient) HandleEvents(api_client *ApiClient, agent *Agent) error {

	for {
		select {
		case event := <-client.Client4.EventChannel:
			if event != nil && event.EventType() == model.WebsocketEventPosted {
				
				post := model.Post{}
				json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)


				if post.RootId == "" {
					result, err := agent.handleCommands(&post)

					if err != nil {
						log.Print(err)
						break
					}
					if result == "" {
						break
					}

					duration := random_ms_duration(1000)
					time.Sleep(duration)
					retry(3, duration, func() error {

						new_post := model.Post{
							ChannelId: post.ChannelId,
							Message: result,
							RootId:  post.Id,
						}
						_, _, err = api_client.CreatePost(&new_post)
						if err != nil {
							log.Println("could not post message ", err)
						}
						return err
					})
					log.Println("Done")
				}
			}

		case ping := <-client.Client4.PingTimeoutChannel:
			log.Printf("New ping: %s\n", ping)

		case resp := <-client.Client4.ResponseChannel:
			listen_error := client.Client4.ListenError

			if listen_error != nil {
				log.Print("Listen error ", listen_error)
				time.Sleep(1 * time.Minute)
				return listen_error
			}

			log.Printf("New resp: %s err: %v\n", resp.Status, resp.Error)

		}
	}

	return nil
}

