package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/kkdai/youtube/v2"
)

// DownloadVideo downloads a video from YouTube

func DownloadVideo(videoURL string, outputDir string) (string, error) {

	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)

	if err != nil {

		return "", fmt.Errorf("error getting video: %v", err)

	}

	formats := video.Formats.WithAudioChannels()

	stream, _, err := client.GetStream(video, &formats[0])

	if err != nil {

		return "", fmt.Errorf("error getting video stream: %v", err)

	}

	// Clean the video title for use in a file name

	safeTitle := strings.ReplaceAll(video.Title, " ", "_")

	outputPath := filepath.Join(outputDir, safeTitle+".mp4")

	file, err := os.Create(outputPath)

	if err != nil {

		return "", fmt.Errorf("error creating file: %v", err)

	}

	defer file.Close()

	_, err = file.ReadFrom(stream)

	if err != nil {

		return "", fmt.Errorf("error downloading video: %v", err)

	}

	return outputPath, nil

}

// DownloadPlaylist downloads all videos in a YouTube playlist

func DownloadPlaylist(playlistURL string, outputDir string) error {

	client := youtube.Client{}

	playlist, err := client.GetPlaylist(playlistURL)

	if err != nil {

		return fmt.Errorf("error getting playlist: %v", err)

	}

	fmt.Printf("Found %d videos in playlist.\n", len(playlist.Videos))

	for _, video := range playlist.Videos {

		fmt.Printf("Downloading video: %s\n", video.Title)

		_, err := DownloadVideo("https://www.youtube.com/watch?v="+video.ID, outputDir)

		if err != nil {

			return fmt.Errorf("error downloading video %s: %v", video.Title, err)

		}

	}

	return nil

}

func main() {
	// Add this line at the beginning of the main function
	os.Setenv("GOOS", "windows")

	// Create the Fyne app

	myApp := app.New()

	myWindow := myApp.NewWindow("YouTube Playlist Downloader")

	// Input field for YouTube URL

	urlEntry := widget.NewEntry()

	urlEntry.SetPlaceHolder("Enter YouTube Video or Playlist URL")

	// Label to show download status

	statusLabel := widget.NewLabel("")

	// Variable to hold the chosen output directory

	var outputDir string

	// Button to select download folder

	selectFolderBtn := widget.NewButton("Select Download Folder", func() {

		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {

			if uri != nil {

				outputDir = uri.Path()

				statusLabel.SetText("Download folder: " + outputDir)

			} else {

				statusLabel.SetText("No folder selected")

			}

		}, myWindow)

	})

	// Button to download video or playlist

	downloadBtn := widget.NewButton("Download", func() {

		url := urlEntry.Text

		if url == "" {

			dialog.ShowInformation("Error", "Please enter a YouTube URL", myWindow)

			return

		}

		// Ensure that the user has selected a download folder

		if outputDir == "" {

			dialog.ShowInformation("Error", "Please select a download folder", myWindow)

			return

		}

		// Start the download process

		go func() {

			statusLabel.SetText("Downloading...")

			if strings.Contains(url, "playlist") {

				// Download playlist if URL contains "playlist"

				err := DownloadPlaylist(url, outputDir)

				if err != nil {

					statusLabel.SetText("Playlist download failed!")

					dialog.ShowError(err, myWindow)

					return

				}

				statusLabel.SetText("Playlist download complete!")

			} else {

				// Download single video otherwise

				outputPath, err := DownloadVideo(url, outputDir)

				if err != nil {

					statusLabel.SetText("Download failed!")

					dialog.ShowError(err, myWindow)

					return

				}

				statusLabel.SetText("Download complete: " + outputPath)

			}

		}()

	})

	// Layout UI components

	content := container.NewVBox(

		widget.NewLabel("YouTube Video/Playlist Downloader"),

		urlEntry,

		selectFolderBtn,

		downloadBtn,

		statusLabel,
	)

	// Set the content and show the window

	myWindow.SetContent(content)

	myWindow.Resize(fyne.NewSize(400, 250))

	myWindow.ShowAndRun()

}
