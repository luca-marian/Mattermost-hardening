package main

import (
    "crypto/tls"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    "bytes"
    "fmt"
    "path"
)

type IP struct {
    ID    int    `json:"id"`
    IP  string `json:"ip"`
}


func getKeycloakToken() (string, error) {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    data := url.Values{}
    data.Set("grant_type", "client_credentials")
    data.Set("client_id", IPDENY_CLIENT_ID)
    data.Set("client_secret", IPDENY_CLIENT_SECRET)

    req, err := http.NewRequest("POST", KEYCLOAK_TOKEN_URL, bytes.NewBufferString(data.Encode()))
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }


    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return "", err
    }

    accessToken, ok := result["access_token"].(string)
    if !ok {
        return "", fmt.Errorf("Access token missing from Keycloak response")
    }

    return accessToken, nil

}

func getIPDenyData() ([]IP, error) {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
    var result []IP

    accessToken, err := getKeycloakToken()
    if err != nil {
        return result, err
    }

    client := &http.Client{}

    req, err := http.NewRequest("GET", IPDENY_API_URL, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return result, err
    }

    req.Header.Set("Authorization", "Bearer " + accessToken)

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making request:", err)
        return result, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return result, err
    }

    if err := json.Unmarshal(body, &result); err != nil {
        fmt.Println("Error parsing JSON:", err)
        return result, err
    }

    return result, nil
}

func addIP(item string) error {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    accessToken, err := getKeycloakToken()
    if err != nil {
        return err
    }

    payload := map[string]string{
        "ip": item,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", IPDENY_API_URL, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer " + accessToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}

func deleteIP(id string) error {
    // 1️⃣ Validate length
    if len(id) > 10 {
        return fmt.Errorf("id too long: got %d characters, want ≤ 5", len(id))
    }

    // 2️⃣ Validate that it contains only digits
    for _, r := range id {
        if r < '0' || r > '9' {
            return fmt.Errorf("id contains invalid character %q: only digits are allowed", r)
        }
    }

    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    accessToken, err := getKeycloakToken()
    if err != nil {
        return err
    }


    client := &http.Client{}

    // url := fmt.Sprintf("%s/%s", IPDENY_API_URL, id)

    // 5️⃣ Build URL safely
    u, err := url.Parse(IPDENY_API_URL)
    if err != nil {
        return err
    }
    u.Path = path.Join(u.Path, id)
    req, _ := http.NewRequest("DELETE", u.String(), nil)

    // req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return err
    }

    req.Header.Set("Authorization", "Bearer " + accessToken)

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making request:", err)
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}

func formatIPDenyData(items []IP) string {
    var sb strings.Builder
    for _, item := range items {
        sb.WriteString(fmt.Sprintf("ID: `%d` - IP: `%s`\n", item.ID, item.IP))
    }
    return sb.String()
}
