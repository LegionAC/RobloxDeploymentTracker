package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func getDeployment() {
	var latestHash string

	clientVersion := "https://clientsettingscdn.roblox.com/v1/client-version/WindowsPlayer"

	client := &http.Client{
		Timeout: time.Minute * 5,
	}

	for {
		version, _ := os.ReadFile("latestVersion.txt")

		latestHash = string(version)
		method := "GET"

		req, err := http.NewRequest(method, clientVersion, nil)
		if err != nil {
			fmt.Println("Error on request creation: ", err)
			return
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Accept", "text/plain")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error on request made: ", err)
		}

		var data map[string]string

		json.NewDecoder(resp.Body).Decode(&data)

		defer resp.Body.Close()

		hash := data["clientVersionUpload"]

		if hash == latestHash {
			fmt.Println(hash)
			time.Sleep(3 * time.Second)
			fmt.Println("Hash identical, skipping...")
			continue
		}

		fmt.Println("Getting deployment...")

		deploymentURL := "https://setup.rbxcdn.com/" + hash + "-RobloxApp.zip"

		req, err = http.NewRequest(method, deploymentURL, nil)
		if err != nil {
			fmt.Println("Error on request creation: ", err)
			return
		}

		req.Header.Add("Accept", "application/zip")

		resp, err = client.Do(req)
		if err != nil {
			fmt.Println("Error on request made: ", err)
			return
		}
		out, _ := os.Create(hash + "-RobloxApp.zip")

		io.Copy(out, resp.Body)
		out.Close()

		defer resp.Body.Close()

		os.WriteFile("latestVersion.txt", []byte(hash), 0644)
	}
}

func main() {
	_, err := os.Stat("./Deployments")
	if os.IsNotExist(err) {
		os.Mkdir("./Deployments", 0755)
	}

	_, err = os.Stat("./Deployments/latestVersion.txt")
	if os.IsNotExist(err) {
		os.Create("./Deployments/latestVersion.txt")
	}

	os.Chdir("./Deployments")
	getDeployment()
}
