package load

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shkh/lastfm-go/lastfm"
	"github.com/skratchdot/open-golang/open"
)

// Type for the Scrobble map
type TimeToIntMap map[int64]int

// Holds the data for a single Track
type Track struct {
	Artist, ArtistMBID string
	Name               string
	MBID               string
	Album, AlbumMBID   string
	URL                string

	// Value is from last.fm "count", 0-100
	Tags map[string]int
}

// Holds all data and things a Loader needs
type Loader struct {
	apikey, secret    string
	username          string
	filename          string
	api               *lastfm.Api
	requestToken      <-chan bool
	tagLoadingStarted bool

	// list of all Tracks
	Tracks []Track
	// index from url fragment to track index
	TracksMap map[string]int
	// map from unixtime to track index
	Scrobbles TimeToIntMap

	// Newest and Oldest loaded Scrobble
	NewestTrack, OldestTrack time.Time

	// stores indices into Tracks
	tagLoadQueueOld chan int
	tagLoadQueueNew chan int
}

// Create a new Loader, with set username, API key and API secret.
func NewLoader(username, apikey, secret string) *Loader {
	l := &Loader{apikey: apikey, secret: secret,
		username: username,
		filename: "tracks-" + username + ".json",

		TracksMap:   make(map[string]int),
		Scrobbles:   make(map[int64]int),
		OldestTrack: time.Now(),

		tagLoadQueueOld: make(chan int, 1024*64),
		tagLoadQueueNew: make(chan int, 1024*64),
	}

	l.api = lastfm.New(apikey, secret)

	// create channel for rate-limiting
	l.requestToken = time.Tick(time.Second / 5)

	return l
}

// Authenticate against last.fm API.
//
// Only necessary for some functions.
func (l *Loader) Auth() {
	token, err := l.api.GetToken()
	p(err)
	authUrl := l.api.GetAuthTokenUrl(token)
	open.Start(authUrl)

	fmt.Println("Please accept the request in the opened browser window.")
	fmt.Println("Once done, press enter here.")
	fmt.Fscanln(os.Stdin)
	err = l.api.LoginWithToken(token)
	p(err)
}

// Loads all tracks and tags from config.
//
// The file used is determined by the configured username.
func (l *Loader) LoadState() {
	if _, err := os.Stat(l.filename); os.IsNotExist(err) {
		return
	}
	fmt.Print("Loading...")
	b, err := ioutil.ReadFile(l.filename)
	p(err)
	p(json.Unmarshal(b, &l))
	fmt.Println(" ok")
}

// Save current state
func (l *Loader) SaveState() {
	fmt.Print("Saving...")

	b, err := json.MarshalIndent(l, "", "  ")
	p(err)
	p(ioutil.WriteFile(l.filename, b, 0666))
	fmt.Println(" ok")
}

func (m TimeToIntMap) MarshalJSON() ([]byte, error) {
	// convert times to strings
	m2 := make(map[string]int)
	for k, v := range m {
		m2[fmt.Sprint(k)] = v
	}
	return json.Marshal(m2)
}

func (m TimeToIntMap) UnmarshalJSON(data []byte) error {
	// convert times back to int64s
	var m2 map[string]int
	err := json.Unmarshal(data, &m2)
	if err != nil {
		return err
	}
	for k, v := range m2 {
		i, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return err
		}
		m[i] = v
	}
	return nil
}

// Incrementally loads tracks and tags from last.fm.
//
// Returns the number of tracks loaded.
func (l *Loader) LoadTracksAndTags() int {
	fmt.Println("Loading tracks from last.fm")

	var wait chan bool
	if !l.tagLoadingStarted {
		l.tagLoadingStarted = true
		wait = make(chan bool)
		l.startLoadingTags(wait)
	}

	total := 0
	got := 1
	page := 1
	pages := -1
	t := l.OldestTrack
	tries := 5

	// fetch up until the oldest track we have
	for got > 0 {
		got, pages = l.loadTracks(false, t, page, pages)
		if got == -1 && pages == -1 {
			got = 1
			tries--
			if tries == 0 {
				break
			}
			continue //retry
		}
		page++
		total += got
	}

	page = 1
	got = 1
	t = l.NewestTrack
	tries = 5

	// fetch up behind the latest track we have
	for got > 0 {
		got, pages = l.loadTracks(true, t, page, pages)
		if got == -1 && pages == -1 {
			got = 1
			tries--
			if tries == 0 {
				break
			}
			continue //retry
		}
		page++
		total += got
	}

	// print stats
	fmt.Println("----")
	fmt.Println("Total tracks:    ", len(l.Tracks))
	fmt.Println("Total scrobbles: ", len(l.Scrobbles))
	fmt.Println("Newly downloaded:", total)
	fmt.Println("First scrobble:  ", l.OldestTrack)
	fmt.Println("Last scrobble:   ", l.NewestTrack)
	fmt.Println("----")

	close(l.tagLoadQueueNew)

	if wait != nil {
		fmt.Println("Finishing tag loading...")
		<-wait
	}
	fmt.Println("----")
	fmt.Println("Total tags: ", len(l.TagsFloat()))
	fmt.Println("Top 5 tags: ", l.TopTags(5))
	fmt.Println("----")
	return got
}

// Adds a scrobble manually to the local scrobble list.
//
// Does not submit to last.fm.
func (l *Loader) AddScrobble(track Track, date time.Time) bool {
	// ignore duplicates by time
	_, dup := l.Scrobbles[date.Unix()]
	if dup {
		return false
	}

	// find track
	// cut "http://www.last.fm/music/", "+noredirect/"
	fragment := strings.TrimPrefix(track.URL, "http://www.last.fm/music/")
	fragment = strings.TrimPrefix(fragment, "+noredirect/")
	ind, ok := l.TracksMap[fragment]
	if !ok {
		ind = len(l.Tracks)
		l.Tracks = append(l.Tracks, track)
		l.TracksMap[fragment] = ind
		l.tagLoadQueueNew <- ind
	}

	// add to index
	l.Scrobbles[date.Unix()] = ind

	// keep newest&oldest up to date
	if date.After(l.NewestTrack) {
		l.NewestTrack = date
	}
	if date.Before(l.OldestTrack) {
		l.OldestTrack = date
	}
	return true
}

// returns (gotThisTime, numPages)
// totalPages is just for display
func (l *Loader) loadTracks(doFrom bool, timeFromTo time.Time, page int, totalPages int) (int, int) {
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
	<-l.requestToken // wait for request limit
	res, err := l.api.User.GetRecentTracks(props)
	if err != nil {
		fmt.Println(err)
		return -1, -1
	}
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

		track := Track{
			Artist:     t.Artist.Name,
			ArtistMBID: t.Artist.Mbid,
			Name:       t.Name,
			MBID:       t.Mbid,
			Album:      t.Album.Name,
			AlbumMBID:  t.Album.Mbid,
			URL:        t.Url,
		}
		if l.AddScrobble(track, trackTime) {
			goodTracks++
			fmt.Print(".")
		}
	}
	fmt.Println()
	fmt.Println("valid entries:", goodTracks)
	return goodTracks, res.TotalPages
}

func (l *Loader) startLoadingTags(fin chan<- bool) {
	go func(in chan<- int) {
		for i, v := range l.Tracks {
			if v.Tags == nil {
				in <- i
			}
		}
		close(in)
	}(l.tagLoadQueueOld)

	done := 0
	do := func(i int) {
		// load current
		t := l.Tracks[i]
		if t.Tags != nil {
			fmt.Println("dummy in tagQueue:", t)
			return
		}

		m := l.LoadTags(t)
		// store into current
		l.Tracks[i].Tags = m

		done++

		if done%25 == 0 {
			total := len(l.tagLoadQueueNew) + len(l.tagLoadQueueOld)
			fmt.Print("[tags] got ", done, " / ", total, "left")
			if len(l.tagLoadQueueNew)+1 >= cap(l.tagLoadQueueNew) ||
				len(l.tagLoadQueueOld)+1 >= cap(l.tagLoadQueueOld) {
				fmt.Print("+")
			}
			fmt.Println()
		}
		if done%100 == 0 {
			l.SaveState()
		}
	}

	go func() {
		for i := range l.tagLoadQueueOld {
			do(i)
		}
		for i := range l.tagLoadQueueNew {
			do(i)
		}
		fin <- true
	}()
}

// Loads tags for a single track from last.fm
//
// If t.MBID is set, only the MBID is used. Otherwise, t.Name and t.Artist are
// used.
func (l *Loader) LoadTags(t Track) map[string]int {
	for tries := 0; tries < 5; tries++ {
		props := lastfm.P{"autocorrect": 1}
		if t.MBID != "" {
			props["mbid"] = t.MBID
		} else {
			props["track"] = t.Name
			props["artist"] = t.Artist
		}
		<-l.requestToken // wait for request limit

		res, err := l.api.Track.GetTopTags(props)
		if err != nil {
			fmt.Println("[tags]", err, "request properties:", props)
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			continue
		}
		m := map[string]int{}
		for _, v := range res.Tags {
			i, err := strconv.Atoi(v.Count)
			if err != nil {
				i = 1
			}
			m[v.Name] = i
		}
		return m
	}
	return nil
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
