package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func waitForFile(path string) {
	for {
		_, err := os.Stat(path)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}
}

func fileDownload(path string) {
	var lastSize int64
	for {
		fileInfo, _ := os.Stat(path)
		if fileInfo.Size() == lastSize && lastSize > 0 {
			break
		}

		lastSize = fileInfo.Size()
		time.Sleep(time.Second * 5)
	}
}

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
			time.Sleep(time.Minute * 10)
			fmt.Println("Hash identical, skipping...")
			continue
		}

		fmt.Println("Getting deployment...")

		deploymentURL := "https://www.roblox.com/download/client"

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

		out, _ := os.Create("RobloxPlayerInstaller.exe")

		io.Copy(out, resp.Body)
		out.Close()

		defer resp.Body.Close()
		os.WriteFile("latestVersion.txt", []byte(hash), 0644)

		waitForFile("./RobloxPlayerInstaller.exe")

		fileDownload("./RobloxPlayerInstaller.exe")

		cmd := exec.Command("./RobloxPlayerInstaller.exe")
		err = cmd.Run()

		cmd = exec.Command("taskkill", "/F", "/IM", "RobloxPlayerBeta.exe")
		err = cmd.Run()

		home, _ := os.UserHomeDir()
		pathToRoblox := filepath.Join(home, "AppData", "Local", "Roblox", "Versions", hash)
		pathToStudio := filepath.Join(home, "AppData", "Local", "Roblox", "Versions", "RobloxStudioInstaller.exe")

		fmt.Println("Compressing file...")

		cmd = exec.Command("powershell", "-Command", "Compress-Archive -Path '"+pathToRoblox+"' -DestinationPath './"+hash+".zip'")
		cmd.Run()

		fmt.Println("File finished compressing.")

		os.RemoveAll(pathToRoblox)
		os.Remove(pathToStudio)
		os.Remove("./RobloxPlayerInstaller.exe")

		fmt.Println("Deployment ready.")
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
