package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "crypto/tls"
    "encoding/base64"
    "encoding/json"
    "net/http"
    "io/ioutil"
    "time"
    "fmt"
    "os"
    "io"
)


func getSatelliteData() []map[string]interface{} {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
    var result []map[string]interface{}

    h := hmac.New(sha256.New, []byte(API_SECRET))
    payload := "GET+/api/payload/files/BeryliaSat/BE-SAT/"

    h.Write([]byte(payload))
    sha := base64.StdEncoding.EncodeToString(h.Sum(nil))

    fmt.Println("SHA256 HMAC (Base64):", sha)

    client := &http.Client{}

    req, err := http.NewRequest("GET", XROAD_URL, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return result
    }

    req.Header.Set("Accept", "application/json")
    req.Header.Set("Authorization", "ApiKey " + API_KEY)
    req.Header.Set("MCS-Request-Digest", sha)
    req.Header.Set("X-Road-Client", XROAD_CLIENT)

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making request:", err)
        return result
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return result
    }

    if err := json.Unmarshal(body, &result); err != nil {
        fmt.Println("Error parsing JSON:", err)
        return result
    }

    return result
}


func getRecentImages(satelliteData []map[string]any, creationWindow time.Duration) []map[string]any {
    recentImages := make([]map[string]any, 0)

    for _, img := range satelliteData {
        if img["status"].(string) != "TransferCompleted" {
            continue
        }

        currentTime := time.Now()
        imgTime, _ := time.Parse(time.RFC3339, img["createdAt"].(string))
        timeDiff := currentTime.Sub(imgTime)
        fmt.Println("Time difference: M - ", timeDiff.Minutes(), " H - ", timeDiff.Hours())
        
        if timeDiff > creationWindow {
            continue
        }

        recentImages = append(recentImages, img)
    }

    return recentImages
}

func downloadImage(data map[string]any) {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    fileName := data["fileId"].(string)
    h := hmac.New(sha256.New, []byte(API_SECRET))

    payload := "GET+/api/payload/files/BeryliaSat/BE-SAT/" + fileName

    h.Write([]byte(payload))
    sha := base64.StdEncoding.EncodeToString(h.Sum(nil))

    fmt.Println("SHA256 HMAC (Base64):", sha)

    client := &http.Client{}

    req, err := http.NewRequest("GET", XROAD_URL + fileName, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
    }

    req.Header.Set("Accept", "application/json")
    req.Header.Set("Authorization", "ApiKey " + API_KEY)
    req.Header.Set("MCS-Request-Digest", sha)
    req.Header.Set("X-Road-Client", XROAD_CLIENT)

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making request:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Println("Error: received non-200 status code", resp.StatusCode)
        return
    }

    file, err := os.Create(fileName)
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    if err != nil {
        fmt.Println("Error saving file:", err)
        return
    }

    fmt.Println("File downloaded successfully:", fileName)

}

func formatSatelliteData(data map[string]any) string {
    result := fmt.Sprintf("Asset: %s\nMission: %s\nFile name: %s\nFile Id: %s\nSize: %d\nStatus: %s\nCreation date: %s",
    data["asset"], data["mission"], data["fileName"], data["fileId"], int(data["size"].(float64)), data["status"], data["createdAt"])

    return result
}