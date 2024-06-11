package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"unicode/utf8"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {
	query := "python"

	youtubeApiKey := os.Getenv("YOUTUBE_DATA_API_KEY")

	if youtubeApiKey == "" {
		fmt.Println("YOUTUBE_DATA_API_KEY environment variable is not set.")
		return
	}

	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(youtubeApiKey))
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	call := service.Search.List([]string{"id"}).Q(query).MaxResults(5).Order("viewCount") //.Q(*query).MaxResults(25)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error calling Youtube API: %v", err)
	}

	// videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)
	videos := make(map[string]youtube.Video)

	longestChannelTitle := 0

	for _, item := range response.Items {
		videoResponse, err := service.Videos.List([]string{"snippet,statistics"}).Id(item.Id.VideoId).Do()
		if err != nil {
			log.Fatalf("Error calling Youtube API: %v", err)
		}
		video := videoResponse.Items[0]
		// fmt.Printf("Title: %v | View Count: %v | Date: %v | Channel: %v\n", video.Snippet.Title, video.Statistics.ViewCount, video.Snippet.PublishedAt, video.Snippet.ChannelTitle)
		fmt.Printf("%-12v | %v | Date: %d | Channel: %v\n", video.Snippet.ChannelTitle, video.Snippet.Title, video.Statistics.ViewCount, video.Snippet.PublishedAt)

		if utf8.RuneCountInString(video.Snippet.ChannelTitle) > longestChannelTitle {
			longestChannelTitle = utf8.RuneCountInString(video.Snippet.ChannelTitle)
		}

		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = *video
			// fmt.Println(item.Snippet.MarshalJSON())
		case "youtube#channel":
			channels[item.Id.ChannelId] = video.Snippet.Title
		case "youtube#playlist":
			playlists[item.Id.PlaylistId] = video.Snippet.Title
		}
	}

	// printIDs("Videos", videos)
	// printIDs("Channels", channels)
	// printIDs("Playlists", playlists)
	printVideos(videos, longestChannelTitle)
}

func sortVideos() {
}

func printVideo(title string, channel string) {
	utf8.RuneCountInString(channel)
	fmt.Printf(title)
}

func printVideos(videos map[string]youtube.Video, length int) {
	for _, vid := range videos {
		fmt.Printf("%-*v | %v\n", length, vid.Snippet.ChannelTitle, vid.Snippet.Title)
	}
}

// Print the ID and title of each result in a list as well as a name that
// identifies the list. For example, print the word section name "Videos"
// above a list of video search results, followed by the video ID and title
// of each matching video.
func printIDs(sectionName string, matches map[string]string) {
	fmt.Printf("%v:\n", sectionName)
	for id, title := range matches {
		fmt.Printf("[%v] %v\n", id, title)
	}
	fmt.Printf("\n\n")
}
