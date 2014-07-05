package load

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
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

type Tracks []Track

func (t Tracks) Len() int           { return len(t) }
func (t Tracks) Less(i, j int) bool { return t[i].Date.Before(t[j].Date) }
func (t Tracks) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

type Data struct {
	Tracks                   []Track
	times                    map[time.Time]int
	NewestTrack, OldestTrack time.Time
	TotalTracks              int
	LoadedTracks             int
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
		Data: Data{
			times:       make(map[time.Time]int),
			OldestTrack: time.Now(),
		},
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

	// build date index
	for i, t := range l.Tracks {
		l.times[t.Date] = i
	}
}

func (l *loader) Save() {
	b, err := json.MarshalIndent(l.Data, "", "  ")
	p(err)
	p(ioutil.WriteFile(l.filename, b, 0666))
}

func (l *loader) LoadTracks() int {
	fmt.Println("Loading tracks")

	total := 0
	got := 1
	page := 1
	pages := -1
	t := l.OldestTrack

	// fetch up until the oldest track we have
	for got > 0 {
		got, pages = l.loadTracks(false, t, page, pages)
		page++
		total += got
	}

	page = 1
	got = 1
	t = l.NewestTrack

	// fetch up behind the latest track we have
	for got > 0 {
		got, pages = l.loadTracks(true, t, page, pages)
		total += got
	}

	sort.Sort(sort.Reverse(Tracks(l.Tracks)))

	// print stats
	fmt.Println("----")
	fmt.Println("Total tracks:    ", len(l.Tracks))
	fmt.Println("Newly downloaded:", total)
	fmt.Println("First track:     ", l.OldestTrack)
	fmt.Println("Last track:      ", l.NewestTrack)
	fmt.Println("----")

	return got
}

// returns (gotThisTime, numPages)
// totalPages is just for display
func (l *loader) loadTracks(doFrom bool, timeFromTo time.Time, page int, totalPages int) (int, int) {
	props := lastfm.P{
		"limit": 200,
		"user":  l.username,
		"page":  page,
	}
	if doFrom {
		fmt.Println("batch after", timeFromTo, " page", page, "/", totalPages)
		props["from"] = fmt.Sprint(timeFromTo.Unix())
	} else {
		fmt.Println("batch until", timeFromTo, " page", page, "/", totalPages)
		props["to"] = fmt.Sprint(timeFromTo.Unix())
	}
	res, err := l.api.User.GetRecentTracks(props)
	if err != nil {
		fmt.Println(err)
		return -1, -1
	}
	l.TotalTracks = res.Total
	goodTracks := 0
	// mitigate some insanity
	if len(res.Tracks) > res.Total && res.Total <= 2 {
		return 0, 0
	}
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

		// ignore duplicates by time
		_, dup := l.times[trackTime]
		if dup {
			continue
		}
		// add to index
		l.times[trackTime] = len(l.Tracks)

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
		goodTracks++
		fmt.Print(".")

		// keep newest&oldest up to date
		if trackTime.After(l.NewestTrack) {
			l.NewestTrack = trackTime
		}
		if trackTime.Before(l.OldestTrack) {
			l.OldestTrack = trackTime
		}
	}
	fmt.Println()
	fmt.Println("valid entries:", goodTracks)
	return goodTracks, res.TotalPages
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
