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

type Track struct {
	Artist, ArtistMBID string
	Name               string
	MBID               string
	Album, AlbumMBID   string
	URL                string

	// value is last.fm "count", 0-100
	Tags map[string]int
}

type Loader struct {
	apikey, secret string
	username       string
	filename       string
	api            *lastfm.Api
	requestToken   <-chan bool

	Tracks []Track
	// index (url fragment) -> tracks[index]
	TracksMap map[string]int
	// index unixtime -> tracks[index]
	Scrobbles map[int64]int `json:"-"`

	NewestTrack, OldestTrack time.Time

	// for json
	ScrobblesJSON map[string]int `json:"Scrobbles"`

	// stores indices into Tracks
	tagLoadQueueOld chan int
	tagLoadQueueNew chan int
}

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

	// limit outgoing requests
	tokens := make(chan bool, 20)
	go func(in <-chan time.Time, out chan<- bool) {
		// every second,
		for _ = range in {
			// try to add 5 tokens
			for i := 0; i < 5; i++ {
				select {
				case tokens <- true:
				default:
				}
			}
		}
	}(time.Tick(time.Second), tokens)
	l.requestToken = tokens

	return l
}

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

// Loads all tracks from config
func (l *Loader) Load() {
	if _, err := os.Stat(l.filename); os.IsNotExist(err) {
		return
	}
	fmt.Print("Loading...")
	b, err := ioutil.ReadFile(l.filename)
	p(err)
	p(json.Unmarshal(b, &l))

	// convert times back to int64s
	l.Scrobbles = make(map[int64]int)
	for k, v := range l.ScrobblesJSON {
		i, err := strconv.ParseInt(k, 10, 64)
		p(err)
		l.Scrobbles[i] = v
	}
	fmt.Println(" ok")
}

// Save current state
func (l *Loader) Save() {
	fmt.Print("Saving...")
	// convert times to strings
	l.ScrobblesJSON = make(map[string]int)
	for k, v := range l.Scrobbles {
		l.ScrobblesJSON[fmt.Sprint(k)] = v
	}

	b, err := json.MarshalIndent(l, "", "  ")
	p(err)
	p(ioutil.WriteFile(l.filename, b, 0666))
	fmt.Println(" ok")
}

// Loads tracks as far as the eye can see.
// Returns the number of tracks loaded
func (l *Loader) LoadTracks() int {
	fmt.Println("Loading tracks from last.fm")

	total := 0
	got := 1
	page := 1
	pages := -1
	t := l.OldestTrack

	// fetch up until the oldest track we have
	for got > 0 {
		got, pages = l.loadTracks(false, t, page, pages)
		if got == -1 && pages == -1 {
			got = 1
			continue //retry
		}
		page++
		total += got
	}

	page = 1
	got = 1
	t = l.NewestTrack

	// fetch up behind the latest track we have
	for got > 0 {
		got, pages = l.loadTracks(true, t, page, pages)
		if got == -1 && pages == -1 {
			got = 1
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
	return got
}

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

func (l *Loader) StartLoadingTags(fin chan<- bool) {
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
			fmt.Print("[tags] got ", done, " / ", total)
			if len(l.tagLoadQueueNew)+1 >= cap(l.tagLoadQueueNew) ||
				len(l.tagLoadQueueOld)+1 >= cap(l.tagLoadQueueOld) {
				fmt.Print("+")
			}
			fmt.Println()
		}
		if done%100 == 0 {
			l.Save()
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

func (l *Loader) LoadTags(t Track) map[string]int {
	for {
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
			fmt.Println("[tags]", err)
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
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}
