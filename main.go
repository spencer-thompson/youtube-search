package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"unicode/utf8"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/dustin/go-humanize"
)

func main() {
	// query := ""
	scanner := bufio.NewScanner(os.Stdin)

	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if line == "" {
	// 		break
	// 	}
	// 	query = query + line
	// }
	scanner.Scan()
	query := scanner.Text()

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
	}

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

	// longestTitle := 0
	// longestChannelTitle := 0

	titleLengths := map[string]int{
		"Title":   0,
		"Channel": 0,
		"Date":    0,
		"Views":   0,
	}

	for _, item := range response.Items {
		videoResponse, err := service.Videos.List([]string{"snippet,statistics"}).Id(item.Id.VideoId).Do()
		if err != nil {
			log.Fatalf("Error calling Youtube API: %v", err)
		}
		video := videoResponse.Items[0]
		// fmt.Printf("Title: %v | View Count: %v | Date: %v | Channel: %v\n", video.Snippet.Title, video.Statistics.ViewCount, video.Snippet.PublishedAt, video.Snippet.ChannelTitle)
		// fmt.Printf("%-12v | %v | Date: %d | Channel: %v\n", video.Snippet.ChannelTitle, video.Snippet.Title, video.Statistics.ViewCount, video.Snippet.PublishedAt)

		if utf8.RuneCountInString(video.Snippet.Title) > titleLengths["Title"] {
			titleLengths["Title"] = utf8.RuneCountInString(video.Snippet.Title)
		}

		if utf8.RuneCountInString(video.Snippet.ChannelTitle) > titleLengths["Channel"] {
			titleLengths["Channel"] = utf8.RuneCountInString(video.Snippet.ChannelTitle)
		}

		if utf8.RuneCountInString(video.Snippet.PublishedAt) > titleLengths["Date"] {
			titleLengths["Date"] = utf8.RuneCountInString(video.Snippet.PublishedAt)
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
	printVideos(videos, titleLengths)
}

func sortVideos() {
}

func formatTime(dateStr string) string {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		fmt.Println("Error parsing date:", err)
	}

	layout := "1.2.2006"

	return date.Format(layout)
}

func printVideos(videos map[string]youtube.Video, length map[string]int) {
	for id, vid := range videos {

		videoUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", id)
		fmt.Printf(
			"%v | %-*v | %-*v | %*v | %8v\n",
			videoUrl,
			length["Title"],
			vid.Snippet.Title,
			length["Channel"],
			vid.Snippet.ChannelTitle,
			10,
			formatTime(vid.Snippet.PublishedAt),
			humanize.SIWithDigits(float64(vid.Statistics.ViewCount), 2, ""),
		)
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
