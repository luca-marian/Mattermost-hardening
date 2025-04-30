package main

import (
    "log"
    "github.com/mattermost/mattermost/server/public/model"
    "time"
)


var MM_AUTHTOKEN string = "ef65fea3cf106a233bfcff0074"
var MM_SITEURL string = "https://mattermost.beg.05.berylia.org"
var MM_WSURL string = "wss://mattermost.beg.05.berylia.org"
var MM_TEAM_NAME string = "beg"
var METRICS_CHANNEL string = "infra-metrics"
var METRICS_DELAY time.Duration = 1 * time.Hour


func postMetrics(client *ApiClient) {
    for {
        metrics, _ := getSystemMetrics()

        channel, err := client.GetChannel(METRICS_CHANNEL)

        if err != nil {
            log.Print("Failed to get channel id ", err)
            continue
        }

        post := model.Post {
            ChannelId: channel.Id,
            Message: metrics,
        }
        
        _ , _ , err = client.CreatePost(&post)
        if err != nil {
            log.Print("Could not post message ", err)
        }

        time.Sleep(METRICS_DELAY)    
    }
}


func main() {

    // create client
    api_client := ApiClient{Url: MM_SITEURL, Token: MM_AUTHTOKEN, Team: MM_TEAM_NAME}
    api_client.Setup()
    
    go postMetrics(&api_client)


    select {}

}
