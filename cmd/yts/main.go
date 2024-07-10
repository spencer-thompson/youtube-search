package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/dustin/go-humanize"
)

func getVideoInfo(s *youtube.Service, id string, vids *map[string]youtube.Video, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := s.Videos.List([]string{"snippet,statistics"}).Id(id).Do()
	if err != nil {
		log.Fatalf("Error Getting Video Info: %v", err)
		return
	}
	if len(resp.Items) < 1 {
		return
	} else {
		(*vids)[id] = *resp.Items[0]
	}
}

func main() {
	var query string
	var numResults int
	flag.StringVar(&query, "s", "neovim", "search query")
	flag.IntVar(&numResults, "n", 10, "total number of results to return")
	flag.Parse()

	// Handle flags/command line args or stdin
	// query := ""
	if len(flag.Args()) > 0 {
		query = flag.Arg(0)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		query = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading stdin:", err)
		}
	}

	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if line == "" {
	// 		break
	// 	}
	// 	query = query + line
	// }

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

	response, err := service.Search.List([]string{"id"}).Q(query).MaxResults(int64(numResults)).Order("viewCount").Do()
	if err != nil {
		log.Fatalf("Error Searching Youtube: %v", err)
	}

	// videos := make(map[string]string)
	videos := make(map[string]youtube.Video)

	// longestTitle := 0
	// longestChannelTitle := 0
	var wg sync.WaitGroup

	// fmt.Println("test")
	for _, item := range response.Items {

		wg.Add(1)
		go getVideoInfo(service, item.Id.VideoId, &videos, &wg)

		// videoResponse, err := service.Videos.List([]string{"snippet,statistics"}).Id(item.Id.VideoId).Do()
		// if err != nil {
		// 	log.Fatalf("Error calling Youtube API: %v", err)
		// }
		// video := videoResponse.Items[0]
		//
		// switch item.Id.Kind {
		// case "youtube#video":
		// 	videos[item.Id.VideoId] = *video
		// 	// fmt.Println(item.Snippet.MarshalJSON())
		// case "youtube#channel":
		// 	channels[item.Id.ChannelId] = video.Snippet.Title
		// case "youtube#playlist":
		// 	playlists[item.Id.PlaylistId] = video.Snippet.Title
		// }
	}

	wg.Wait()
	printVideos(videos)
}

func sortVideos() {
}

func formatTime(dateStr string) string {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		fmt.Println("Error parsing date:", err)
	}

	layout := "1.02.2006"

	return date.Format(layout)
}

func truncate(s string, max int) string {
	if max > utf8.RuneCountInString(s) {
		return s
	}
	return fmt.Sprintf("%v...", s[:(max-3)])
}

func printVideos(videos map[string]youtube.Video) {
	length := map[string]int{
		"Title":        0,
		"Channel":      0,
		"titleLimit":   50,
		"channelLimit": 20,
	}

	for _, vid := range videos {
		// find longest of each
		if utf8.RuneCountInString(vid.Snippet.Title) > length["titleLimit"] {
			length["Title"] = length["titleLimit"]
		} else if utf8.RuneCountInString(vid.Snippet.Title) > length["Title"] {
			length["Title"] = utf8.RuneCountInString(vid.Snippet.Title)
		}
		if utf8.RuneCountInString(vid.Snippet.ChannelTitle) > length["channelLimit"] {
			length["Channel"] = length["channelLimit"]
		} else if utf8.RuneCountInString(vid.Snippet.ChannelTitle) > length["Channel"] {
			length["Channel"] = utf8.RuneCountInString(vid.Snippet.ChannelTitle)
		}
	}

	for id, vid := range videos {
		// videoUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", id)
		fmt.Printf(
			"%-*v | %-*v | %*v | %8v | %v\n",
			length["Title"],
			truncate(vid.Snippet.Title, length["Title"]),
			length["Channel"],
			truncate(vid.Snippet.ChannelTitle, length["Channel"]),
			10,
			formatTime(vid.Snippet.PublishedAt),
			humanize.SIWithDigits(float64(vid.Statistics.ViewCount), 2, ""),
			// videoUrl,
			id,
		)
	}
}
