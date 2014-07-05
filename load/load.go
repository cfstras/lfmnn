package load

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/shkh/lastfm-go/lastfm"
	"github.com/skratchdot/open-golang/open"
)

type Loader interface {
	Auth()

	// Loads tracks as far as the eye can see.
	// Returns the number of tracks loaded
	LoadTracks() int
	// Loads all tracks from config
	Load()
	// Save current state
	Save()
}

type Track struct {
	Artist, ArtistMBID string
	Name               string
	MBID               string
	Album, AlbumMBID   string
	URL                string
	Date               time.Time
}

type Data struct {
	Tracks       []Track
	NewestTrack  time.Time
	TotalTracks  int
	LoadedTracks int
}

type loader struct {
	apikey, secret string
	username       string
	filename       string
	api            *lastfm.Api

	Data
}

func NewLoader(username, apikey, secret string) Loader {
	l := &loader{apikey: apikey, secret: secret,
		username: username,
		filename: "tracks-" + username + ".json",
	}

	l.api = lastfm.New(apikey, secret)
	return l
}

func (l *loader) Auth() {
	token, err := l.api.GetToken()
	p(err)
	authUrl := l.api.GetAuthTokenUrl(token)
	open.Start(authUrl)

	fmt.Println("Please accept the request in the opened browser window.")
	fmt.Println("Once done, press enter here.")
	fmt.Fscanln(os.Stdin)
	err = l.api.LoginWithToken(token) //discarding error
	p(err)
}

func (l *loader) Load() {
	if _, err := os.Stat(l.filename); os.IsNotExist(err) {
		return
	}
	b, err := ioutil.ReadFile(l.filename)
	p(err)
	p(json.Unmarshal(b, &l.Data))
}

func (l *loader) Save() {
	b, err := json.MarshalIndent(l.Data, "", "  ")
	p(err)
	p(ioutil.WriteFile(l.filename, b, 0666))
}

func (l *loader) LoadTracks() int {
	fmt.Println("Loading tracks")
	res, err := l.api.User.GetRecentTracks(lastfm.P{
		"limit": 200,
		"user":  l.username,
		"from":  fmt.Sprint(l.NewestTrack.Unix()),
	})
	p(err)
	l.TotalTracks = res.Total
	goodTracks := 0
	fmt.Println("Got", len(res.Tracks), "- total", res.Total)
	for _, t := range res.Tracks {
		i, err := strconv.Atoi(t.Date.Uts)
		var trackTime time.Time
		if err != nil {
			// Mon Jan 2 15:04:05 -0700 MST 2006
			trackTime, err = time.Parse("2 Jan 2006, 15:04", t.Date.Date)
			if err != nil {
				fmt.Println("Unknown time, ignoring track:", t)
				continue
			}
		} else {
			trackTime = time.Unix(int64(i), 0)
		}

		l.Tracks = append(l.Tracks, Track{
			Artist:     t.Artist.Name,
			ArtistMBID: t.Artist.Mbid,
			Name:       t.Name,
			MBID:       t.Mbid,
			Album:      t.Album.Name,
			AlbumMBID:  t.Album.Mbid,
			URL:        t.Url,
			Date:       trackTime,
		})
		fmt.Print(".")
		goodTracks++
		if trackTime.After(l.NewestTrack) {
			l.NewestTrack = trackTime
		}
	}
	fmt.Println()
	fmt.Println("Valid entries:", goodTracks)
	return len(res.Tracks)
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
