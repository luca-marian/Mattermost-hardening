package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"context"
	"log"
	"time"
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

