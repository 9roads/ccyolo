package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ccyolo to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", Version)
		fmt.Println("Checking for updates...")

		latest, err := getLatestRelease()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		latestVersion := strings.TrimPrefix(latest.TagName, "v")
		if latestVersion == Version {
			fmt.Println("Already up to date!")
			return
		}

		fmt.Printf("New version available: %s\n", latestVersion)

		// Find the right asset for this platform
		assetName := getAssetName()
		var downloadURL string
		for _, asset := range latest.Assets {
			if asset.Name == assetName {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			fmt.Printf("No binary available for %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Println("Please update manually or use your package manager.")
			return
		}

		fmt.Printf("Downloading %s...\n", assetName)

		// Get current executable path
		execPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Download to temp file
		resp, err := http.Get(downloadURL)
		if err != nil {
			fmt.Printf("Download error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		tmpFile, err := os.CreateTemp("", "ccyolo-*")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		tmpFile.Close()

		// Make executable
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Replace current binary
		if err := os.Rename(tmpFile.Name(), execPath); err != nil {
			// Try with sudo on Unix
			if runtime.GOOS != "windows" {
				fmt.Println("Need elevated permissions to update. Trying sudo...")
				cmd := exec.Command("sudo", "mv", tmpFile.Name(), execPath)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("Update failed: %v\n", err)
					return
				}
			} else {
				fmt.Printf("Update failed: %v\n", err)
				return
			}
		}

		fmt.Printf("Updated to version %s!\n", latestVersion)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show ccyolo version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ccyolo version %s\n", Version)
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func getLatestRelease() (*Release, error) {
	resp, err := http.Get("https://api.github.com/repos/9roads/ccyolo/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch arch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	}

	ext := ""
	if os == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("ccyolo-%s-%s%s", os, arch, ext)
}
