package main

import (
    "fmt"
    "log"
    "time"
    "github.com/mattermost/mattermost/server/public/model"
)


var MM_AUTHTOKEN string = "593f1c0f1354d7409c97419946" 
var MM_SITEURL string = "https://mattermost.beg.05.berylia.org"

var MM_WSURL string = "wss://mattermost.beg.05.berylia.org"

var MM_TEAM_NAME string = "beg"
var IPDENY_CHANNEL string = "ipdeny-feed"


var KEYCLOAK_TOKEN_URL = "https://keycloak.bps.05.berylia.org/realms/bps.05.berylia.org/protocol/openid-connect/token"
var IPDENY_API_URL = "https://ipdeny.bps.05.berylia.org/api"

var IPDENY_CLIENT_ID = "ipdeny"
var IPDENY_CLIENT_SECRET = "rvKZ08yikGnbLv3o2MMbTXSLRrG07Rjb"


var IPDENY_DELAY time.Duration = 10 * time.Minute



func postIPs(client *ApiClient) {
    var lastResponse string

    for {
        time.Sleep(IPDENY_DELAY)

        data, err := getIPDenyData()
        if err != nil {
            log.Print("Failed to get IPDeny data ", err)
            continue
        }

        channel, err := client.GetChannel(IPDENY_CHANNEL)

        if err != nil {
            log.Print("Failed to get channel id ", err)
            continue
        }

        ipList := formatIPDenyData(data)

        if ipList == lastResponse {
            log.Print("No new IP info")
            continue
        }

        message := fmt.Sprintf("Current IPDeny IP list\n%s", ipList)

        post := model.Post {
            ChannelId: channel.Id,
            Message: message,
        }

        _ , _ , err = client.CreatePost(&post)
        if err != nil {
            log.Print("Could not post message ", err)
        }

        lastResponse = ipList
    }
}


func main() {
    // create client
    api_client := ApiClient{Url: MM_SITEURL, Token: MM_AUTHTOKEN, Team: MM_TEAM_NAME}
    api_client.Setup()
    
    go postIPs(&api_client)


    ws_client := WSClient{Url: MM_WSURL, Token: MM_AUTHTOKEN}
    err := ws_client.Connect()
    
    if err != nil {
        log.Printf("Failed to connect to WebSocket: %v", err)
        
    }

    ws_client.Listen()
    defer ws_client.Close()

    bot := Bot{}
    
    for {
        err = ws_client.HandleEvents(&api_client, &bot)
        if err != nil {
            log.Print("Error occured while handling websocket events, reconnecting: ", err)
        }

        ws_client.Connect()
        ws_client.Listen()
    }

}