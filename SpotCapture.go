package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/pat"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

// Used to persist token and playlist id
type SpotCaptureConfig struct {
	Token      *oauth2.Token
	PlaylistId spotify.ID
	UserId     string
}

func loadConfig() *SpotCaptureConfig {
	raw, err := ioutil.ReadFile("/Users/idavies/.spotcapture")
	if err != nil {
		fmt.Println("Failed to open /Users/idavies/.spotcapture")
		return nil
	}

	var config *SpotCaptureConfig
	err = json.Unmarshal(raw, &config)
	if err != nil {
		fmt.Println("Failed to parse /Users/idavies/.spotcapture")
		removeConfig()
		return nil
	}
	fmt.Printf("Config: %s\n", config)
	return config
}

func saveConfig(config *SpotCaptureConfig) {
	raw, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("Failed to marshal config. %s\n", err.Error())
	}

	err = ioutil.WriteFile("/Users/idavies/.spotcapture", raw, 0644)
	if err != nil {
		fmt.Printf("Failed to persist config: %s\n", err.Error())
	}
}

func removeConfig() {
	err := os.Remove("/Users/idavies/.spotcapture")
	if err != nil {
		fmt.Println("Unable to remove invalid configuration file!")
		os.Exit(1)
	}
}

func handleAuth(done chan *oauth2.Token, state string, auth spotify.Authenticator) {
	p := pat.New()
	p.Get("/auth/spotify/callback", func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("Handling auth callback\n")
		// use the same state string here that you used to generate the URL
		token, err := auth.Token(state, req)
		if err != nil {
			http.Error(res, "Couldn't get token", http.StatusNotFound)
			return
		}

		done <- token
		res.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(res, "Ok")
	})
	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(res, "Welcome to SpotCapture!")
	})
	http.ListenAndServe(":3000", p)
}

func createPlaylist(client spotify.Client, userId string, playlistName string) spotify.ID {
	playlist, err := client.CreatePlaylistForUser(userId, playlistName, false)
	if err != nil {
		fmt.Printf("Unable to create playlist for current user: %s\n", err.Error())
		os.Exit(1)
	}
	return playlist.ID
}

func currentTrack(client spotify.Client) *spotify.CurrentlyPlaying {
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		fmt.Printf("Couldn't get Current Track. %s", err.Error())
		return nil
	}

	return currentlyPlaying
}

func currentUserId(client spotify.Client) string {
	user, err := client.CurrentUser()
	if err != nil {
		fmt.Printf("Couldn't get current user: %s\n", err.Error())
		os.Exit(1)
	}
	return user.ID
}

func main() {
	removeFlag := flag.Bool("remove", false, "Remove currently playing track rather than adding")
	flag.Parse()

	auth := spotify.NewAuthenticator("http://localhost:3000/auth/spotify/callback", spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopePlaylistReadCollaborative, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistModifyPrivate, spotify.ScopeUserLibraryModify)
	auth.SetAuthInfo("fill in client id", "fill in secret key")
	done := make(chan *oauth2.Token)

	loaded := loadConfig()

	if loaded == nil {
		// Need to perform our auth dance
		state := "cheesecake"

		go handleAuth(done, state, auth)

		// Prompt to login
		err := exec.Command("open", auth.AuthURL(state)).Run()
		if err != nil {
			fmt.Printf("Couldn't launch browser to authenticate. %s", err.Error())
			os.Exit(1)
		}

		fmt.Println("Waiting for you to login in the opened browser window!")

		token := <-done

		fmt.Println("Logged in successfully")

		client := auth.NewClient(token)

		config := new(SpotCaptureConfig)
		config.Token = token
		config.UserId = currentUserId(client)
		config.PlaylistId = createPlaylist(client, config.UserId, "SpotCapture")

		saveConfig(config)

		fmt.Println("Configuration created!")
	}

	loaded = loadConfig()
	client := auth.NewClient(loaded.Token)

	track := currentTrack(client)
	if track != nil && track.Playing && track.Item != nil {
		if *removeFlag {
			client.RemoveTracksFromPlaylist(loaded.UserId, loaded.PlaylistId, track.Item.ID)
			fmt.Printf("Sucessfully removed '%s - %s' to your playlist", track.Item.Name, track.Item.Artists)
		} else {
			client.AddTracksToPlaylist(loaded.UserId, loaded.PlaylistId, track.Item.ID)
			fmt.Printf("Sucessfully added '%s - %s' to your playlist", track.Item.Name, track.Item.Artists)
		}
	}
}