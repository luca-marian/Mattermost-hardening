package main

import (
    "os"
    "log"
    "github.com/mattermost/mattermost/server/public/model"
    "time"
    "context"
)


var MM_AUTHTOKEN string = "9bdc282343efec4fe1f35324d2"
var MM_SITEURL string = "https://mattermost.beg.05.berylia.org"
var MM_TEAM_NAME string = "beg"
var SAT_CHANNEL string = "satellite-feed"
var SAT_DELAY time.Duration = 10 * time.Minute
var CREATION_WINDOW time.Duration = 90 * time.Minute
var API_SECRET = "UzP4f8gP6u994vNddXfBrGyjS10yE1vm"
var API_KEY = "x5mml5TbUnK2nk6B0vlq4psaBSuVw7Tj"

var XROAD_URL = "https://xroad-beg.beg.05.berylia.org/r1/BY-05/GOV/1001/satellite/satellite_data/"
var XROAD_CLIENT = "BY-05/GOV/1003/recon"


func postImages(client *ApiClient) {
    seenBefore := make(map[string]bool)

    for {
        log.Println("Seen before: ", seenBefore)

        data := getSatelliteData()
        channel, err := client.GetChannel(SAT_CHANNEL)

        if err != nil {
            log.Print("Failed to get channel id ", err)
            continue
        }

        newImages := make([]map[string]any, 0)
        recentImages := getRecentImages(data, CREATION_WINDOW)

        log.Println("Recent images:", recentImages)

        for _, img := range recentImages {
            fileId := img["fileId"].(string)

            _, exists := seenBefore[fileId]
            if exists == false {
                newImages = append(newImages, img)
                seenBefore[fileId] = true
            }
        }

        log.Println("New images: ", newImages)

        for _, img := range newImages {
            message := formatSatelliteData(img)
            downloadImage(img)

            fileName := img["fileId"].(string)
            fileBytes, err := os.ReadFile(fileName)

            if err != nil {
                log.Println("Error reading file:", err)
                continue
            }

            uploadResp, _, err := client.Client4.UploadFile(context.Background(), fileBytes, channel.Id, fileName + ".png")
            if err != nil {
                log.Println("Error uploading file:", err)
                continue
            }

            fileId := uploadResp.FileInfos[0].Id
            log.Println("File uploaded successfully, file ID:", fileId)

            post := model.Post {
                ChannelId: channel.Id,
                Message: message,
                FileIds: []string{fileId},
            }

            _ , _ , err = client.CreatePost(&post)
            if err != nil {
                log.Print("Could not post message ", err)
            }

            os.Remove(fileName)
        }

        time.Sleep(SAT_DELAY)
    }
}


func main() {

    // create client
    api_client := ApiClient{Url: MM_SITEURL, Token: MM_AUTHTOKEN, Team: MM_TEAM_NAME}
    api_client.Setup()

    go postImages(&api_client)

    select {}

}