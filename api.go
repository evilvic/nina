package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
)

type API struct {
	cfg      Config
	lib      *Library
	store    *Store
	scanning atomic.Bool
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	path := strings.TrimPrefix(r.URL.Path, "/rest/")
	path = strings.TrimSuffix(path, ".view")

	if path != "ping" {
		if err := a.authenticate(r); err != nil {
			writeResponse(w, r, newError(40, "Wrong username or password"))
			return
		}
	}

	log.Printf("%s /rest/%s", r.Method, path)

	switch path {
	case "ping":
		a.ping(w, r)
	case "getLicense":
		a.getLicense(w, r)
	case "getMusicFolders":
		a.getMusicFolders(w, r)
	case "getIndexes":
		a.getIndexes(w, r)
	case "getArtists":
		a.getArtists(w, r)
	case "getArtist":
		a.getArtist(w, r)
	case "getAlbum":
		a.getAlbum(w, r)
	case "getSong":
		a.getSong(w, r)
	case "getMusicDirectory":
		a.getMusicDirectory(w, r)
	case "stream", "download":
		a.stream(w, r)
	case "getCoverArt":
		a.getCoverArt(w, r)
	case "search2":
		a.search2(w, r)
	case "search3":
		a.search3(w, r)
	case "getAlbumList":
		a.getAlbumList(w, r)
	case "getAlbumList2":
		a.getAlbumList2(w, r)
	case "getRandomSongs":
		a.getRandomSongs(w, r)
	case "getUser":
		a.getUser(w, r)
	case "getPlaylists":
		a.getPlaylists(w, r)
	case "getPlaylist":
		a.getPlaylist(w, r)
	case "createPlaylist":
		a.createPlaylist(w, r)
	case "updatePlaylist":
		a.updatePlaylist(w, r)
	case "deletePlaylist":
		a.deletePlaylist(w, r)
	case "star":
		a.star(w, r)
	case "unstar":
		a.unstar(w, r)
	case "getStarred":
		a.getStarred(w, r)
	case "getStarred2":
		a.getStarred2(w, r)
	case "scrobble":
		a.scrobble(w, r)
	case "startScan":
		a.startScan(w, r)
	case "getScanStatus":
		a.getScanStatus(w, r)
	case "setRating":
		writeResponse(w, r, newOK())
	default:
		log.Printf("unhandled: %s", path)
		writeResponse(w, r, newOK())
	}
}

func (a *API) authenticate(r *http.Request) error {
	user := r.FormValue("u")
	if user != a.cfg.User {
		return fmt.Errorf("unknown user")
	}

	if token := r.FormValue("t"); token != "" {
		salt := r.FormValue("s")
		expected := md5Hex(a.cfg.Pass + salt)
		if token != expected {
			return fmt.Errorf("bad token")
		}
		return nil
	}

	pass := r.FormValue("p")
	if strings.HasPrefix(pass, "enc:") {
		decoded, err := hex.DecodeString(strings.TrimPrefix(pass, "enc:"))
		if err != nil {
			return fmt.Errorf("bad encoded password")
		}
		pass = string(decoded)
	}
	if pass != a.cfg.Pass {
		return fmt.Errorf("bad password")
	}
	return nil
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// Basic endpoints

func (a *API) ping(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r, newOK())
}

func (a *API) getLicense(w http.ResponseWriter, r *http.Request) {
	resp := newOK()
	resp.License = &License{Valid: true}
	writeResponse(w, r, resp)
}

func (a *API) getMusicFolders(w http.ResponseWriter, r *http.Request) {
	resp := newOK()
	resp.MusicFolders = &MusicFolders{
		Folders: []MusicFolder{{ID: 1, Name: "Music"}},
	}
	writeResponse(w, r, resp)
}

// Browsing

func (a *API) getIndexes(w http.ResponseWriter, r *http.Request) {
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	indexMap := make(map[string][]Child)
	for _, artist := range a.lib.SortedArtists {
		c := indexChar(artist.Name)
		indexMap[c] = append(indexMap[c], Child{
			ID:      artist.ID,
			IsDir:   true,
			Title:   artist.Name,
			Starred: a.store.StarredTime(artist.ID),
		})
	}

	var indexes []Index
	for _, c := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ#" {
		key := string(c)
		if artists, ok := indexMap[key]; ok {
			indexes = append(indexes, Index{Name: key, Artists: artists})
		}
	}

	resp := newOK()
	resp.Indexes = &Indexes{Index: indexes}
	writeResponse(w, r, resp)
}

func (a *API) getArtists(w http.ResponseWriter, r *http.Request) {
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	indexMap := make(map[string][]ArtistID3)
	for _, artist := range a.lib.SortedArtists {
		c := indexChar(artist.Name)
		indexMap[c] = append(indexMap[c], ArtistID3{
			ID:         artist.ID,
			Name:       artist.Name,
			AlbumCount: len(artist.AlbumIDs),
			Starred:    a.store.StarredTime(artist.ID),
		})
	}

	var indexes []IndexID3
	for _, c := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ#" {
		key := string(c)
		if artists, ok := indexMap[key]; ok {
			indexes = append(indexes, IndexID3{Name: key, Artists: artists})
		}
	}

	resp := newOK()
	resp.Artists = &ArtistsID3{Index: indexes}
	writeResponse(w, r, resp)
}

func (a *API) getArtist(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	artist, ok := a.lib.Artists[id]
	if !ok {
		writeResponse(w, r, newError(70, "Artist not found"))
		return
	}

	albums := a.lib.GetAlbumsForArtist(id)
	var albumsID3 []AlbumID3
	for _, al := range albums {
		albumsID3 = append(albumsID3, a.albumToID3(al))
	}

	resp := newOK()
	resp.ArtistResponse = &ArtistWithAlbumsID3{
		ID:         artist.ID,
		Name:       artist.Name,
		AlbumCount: len(artist.AlbumIDs),
		Starred:    a.store.StarredTime(artist.ID),
		Albums:     albumsID3,
	}
	writeResponse(w, r, resp)
}

func (a *API) getAlbum(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	album, ok := a.lib.Albums[id]
	if !ok {
		writeResponse(w, r, newError(70, "Album not found"))
		return
	}

	songs := a.lib.GetSongsForAlbum(id)
	var children []Child
	for _, s := range songs {
		children = append(children, a.songToChild(s))
	}

	coverArt := ""
	if album.CoverPath != "" {
		coverArt = album.ID
	}

	resp := newOK()
	resp.AlbumResponse = &AlbumWithSongsID3{
		ID:        album.ID,
		Name:      album.Name,
		Artist:    album.ArtistName,
		ArtistID:  album.ArtistID,
		CoverArt:  coverArt,
		SongCount: len(album.SongIDs),
		Duration:  album.Duration,
		Year:      album.Year,
		Starred:   a.store.StarredTime(album.ID),
		Songs:     children,
	}
	writeResponse(w, r, resp)
}

func (a *API) getSong(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	song, ok := a.lib.Songs[id]
	if !ok {
		writeResponse(w, r, newError(70, "Song not found"))
		return
	}

	child := a.songToChild(song)
	resp := newOK()
	resp.SongResponse = &child
	writeResponse(w, r, resp)
}

func (a *API) getMusicDirectory(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	if artist, ok := a.lib.Artists[id]; ok {
		albums := a.lib.GetAlbumsForArtist(id)
		var children []Child
		for _, al := range albums {
			coverArt := ""
			if al.CoverPath != "" {
				coverArt = al.ID
			}
			children = append(children, Child{
				ID:       al.ID,
				Parent:   artist.ID,
				IsDir:    true,
				Title:    al.Name,
				Artist:   al.ArtistName,
				CoverArt: coverArt,
				Year:     al.Year,
				Starred:  a.store.StarredTime(al.ID),
			})
		}
		resp := newOK()
		resp.Directory = &Directory{
			ID:       artist.ID,
			Name:     artist.Name,
			Children: children,
		}
		writeResponse(w, r, resp)
		return
	}

	if album, ok := a.lib.Albums[id]; ok {
		songs := a.lib.GetSongsForAlbum(id)
		var children []Child
		for _, s := range songs {
			children = append(children, a.songToChild(s))
		}
		resp := newOK()
		resp.Directory = &Directory{
			ID:       album.ID,
			Parent:   album.ArtistID,
			Name:     album.Name,
			Children: children,
		}
		writeResponse(w, r, resp)
		return
	}

	writeResponse(w, r, newError(70, "Directory not found"))
}

// Streaming & transcoding

func (a *API) stream(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.lib.mu.RLock()
	song, ok := a.lib.Songs[id]
	a.lib.mu.RUnlock()

	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	maxBitRate := intParam(r, "maxBitRate", 0)
	format := r.FormValue("format")

	if maxBitRate > 0 && maxBitRate < 320 {
		a.transcode(w, song, maxBitRate, format)
		return
	}

	f, err := os.Open(song.Path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", song.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(song.Size, 10))
	io.Copy(w, f)
}

func (a *API) transcode(w http.ResponseWriter, song *Song, maxBitRate int, format string) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		f, err := os.Open(song.Path)
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", song.ContentType)
		io.Copy(w, f)
		return
	}

	if format == "" || format == "raw" {
		format = "mp3"
	}
	if maxBitRate == 0 {
		maxBitRate = 192
	}

	var args []string
	args = append(args, "-i", song.Path, "-v", "quiet", "-map", "a:0")

	var contentType string
	var outputFmt string

	switch format {
	case "mp3":
		contentType = "audio/mpeg"
		outputFmt = "mp3"
		args = append(args, "-c:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", maxBitRate))
	case "opus":
		contentType = "audio/ogg"
		outputFmt = "ogg"
		args = append(args, "-c:a", "libopus", "-b:a", fmt.Sprintf("%dk", maxBitRate))
	case "aac":
		contentType = "audio/aac"
		outputFmt = "adts"
		args = append(args, "-c:a", "aac", "-b:a", fmt.Sprintf("%dk", maxBitRate))
	case "flac":
		contentType = "audio/flac"
		outputFmt = "flac"
		args = append(args, "-c:a", "flac")
	default:
		contentType = "audio/mpeg"
		outputFmt = "mp3"
		args = append(args, "-c:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", maxBitRate))
	}

	args = append(args, "-f", outputFmt, "pipe:1")

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = w

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Transfer-Encoding", "chunked")

	if err := cmd.Run(); err != nil {
		log.Printf("transcode error: %v", err)
	}
}

func (a *API) getCoverArt(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	if album, ok := a.lib.Albums[id]; ok && album.CoverPath != "" {
		http.ServeFile(w, r, album.CoverPath)
		return
	}

	if song, ok := a.lib.Songs[id]; ok {
		if album, ok := a.lib.Albums[song.AlbumID]; ok && album.CoverPath != "" {
			http.ServeFile(w, r, album.CoverPath)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

// Search

func (a *API) search2(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	artistCount := intParam(r, "artistCount", 20)
	albumCount := intParam(r, "albumCount", 20)
	songCount := intParam(r, "songCount", 20)

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	artists, albums, songs := a.lib.Search(query, artistCount, albumCount, songCount)

	var rArtists []Child
	for _, ar := range artists {
		rArtists = append(rArtists, Child{ID: ar.ID, IsDir: true, Title: ar.Name, Starred: a.store.StarredTime(ar.ID)})
	}
	var rAlbums []Child
	for _, al := range albums {
		coverArt := ""
		if al.CoverPath != "" {
			coverArt = al.ID
		}
		rAlbums = append(rAlbums, Child{
			ID: al.ID, Parent: al.ArtistID, IsDir: true,
			Title: al.Name, Artist: al.ArtistName, CoverArt: coverArt, Year: al.Year,
			Starred: a.store.StarredTime(al.ID),
		})
	}
	var rSongs []Child
	for _, s := range songs {
		rSongs = append(rSongs, a.songToChild(s))
	}

	resp := newOK()
	resp.SearchResult2 = &SearchResult2{Artists: rArtists, Albums: rAlbums, Songs: rSongs}
	writeResponse(w, r, resp)
}

func (a *API) search3(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	artistCount := intParam(r, "artistCount", 20)
	albumCount := intParam(r, "albumCount", 20)
	songCount := intParam(r, "songCount", 20)

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	artists, albums, songs := a.lib.Search(query, artistCount, albumCount, songCount)

	var rArtists []ArtistID3
	for _, ar := range artists {
		rArtists = append(rArtists, ArtistID3{ID: ar.ID, Name: ar.Name, AlbumCount: len(ar.AlbumIDs), Starred: a.store.StarredTime(ar.ID)})
	}
	var rAlbums []AlbumID3
	for _, al := range albums {
		rAlbums = append(rAlbums, a.albumToID3(al))
	}
	var rSongs []Child
	for _, s := range songs {
		rSongs = append(rSongs, a.songToChild(s))
	}

	resp := newOK()
	resp.SearchResult3 = &SearchResult3{Artists: rArtists, Albums: rAlbums, Songs: rSongs}
	writeResponse(w, r, resp)
}

// Album lists

func (a *API) getAlbumList(w http.ResponseWriter, r *http.Request) {
	listType := r.FormValue("type")
	size := intParam(r, "size", 10)
	offset := intParam(r, "offset", 0)

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	albums := a.collectAlbums(listType, size, offset)
	var children []Child
	for _, al := range albums {
		coverArt := ""
		if al.CoverPath != "" {
			coverArt = al.ID
		}
		children = append(children, Child{
			ID: al.ID, Parent: al.ArtistID, IsDir: true,
			Title: al.Name, Artist: al.ArtistName, CoverArt: coverArt, Year: al.Year,
			Starred: a.store.StarredTime(al.ID),
		})
	}

	resp := newOK()
	resp.AlbumList = &AlbumListResponse{Albums: children}
	writeResponse(w, r, resp)
}

func (a *API) getAlbumList2(w http.ResponseWriter, r *http.Request) {
	listType := r.FormValue("type")
	size := intParam(r, "size", 10)
	offset := intParam(r, "offset", 0)

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	albums := a.collectAlbums(listType, size, offset)
	var albumsID3 []AlbumID3
	for _, al := range albums {
		albumsID3 = append(albumsID3, a.albumToID3(al))
	}

	resp := newOK()
	resp.AlbumList2 = &AlbumList2Response{Albums: albumsID3}
	writeResponse(w, r, resp)
}

func (a *API) collectAlbums(listType string, size, offset int) []*Album {
	all := make([]*Album, 0, len(a.lib.Albums))
	for _, al := range a.lib.Albums {
		all = append(all, al)
	}

	switch listType {
	case "random":
		rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	case "alphabeticalByName":
		sort.Slice(all, func(i, j int) bool {
			return strings.ToLower(all[i].Name) < strings.ToLower(all[j].Name)
		})
	case "alphabeticalByArtist":
		sort.Slice(all, func(i, j int) bool {
			return strings.ToLower(all[i].ArtistName) < strings.ToLower(all[j].ArtistName)
		})
	case "starred":
		starred := a.store.GetStarred()
		var starredAlbums []*Album
		for _, al := range all {
			if _, ok := starred.Albums[al.ID]; ok {
				starredAlbums = append(starredAlbums, al)
			}
		}
		all = starredAlbums
	default:
		sort.Slice(all, func(i, j int) bool {
			return all[i].Name < all[j].Name
		})
	}

	if offset >= len(all) {
		return nil
	}
	end := offset + size
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end]
}

func (a *API) getRandomSongs(w http.ResponseWriter, r *http.Request) {
	size := intParam(r, "size", 10)

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	all := make([]*Song, 0, len(a.lib.Songs))
	for _, s := range a.lib.Songs {
		all = append(all, s)
	}
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	if size > len(all) {
		size = len(all)
	}

	var children []Child
	for _, s := range all[:size] {
		children = append(children, a.songToChild(s))
	}

	resp := newOK()
	resp.RandomSongs = &SongsResponse{Songs: children}
	writeResponse(w, r, resp)
}

// User

func (a *API) getUser(w http.ResponseWriter, r *http.Request) {
	resp := newOK()
	resp.User = &UserResponse{
		Username:     a.cfg.User,
		AdminRole:    true,
		StreamRole:   true,
		DownloadRole: true,
		ScrobbleRole: true,
		ShareRole:    true,
		CoverArtRole: true,
		CommentRole:  true,
		PodcastRole:  true,
		UploadRole:   true,
		PlaylistRole: true,
		JukeboxRole:  true,
		SettingsRole: true,
	}
	writeResponse(w, r, resp)
}

// Playlists

func (a *API) getPlaylists(w http.ResponseWriter, r *http.Request) {
	playlists := a.store.GetPlaylists()

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	var entries []PlaylistEntry
	for _, p := range playlists {
		dur := 0
		for _, sid := range p.SongIDs {
			if s, ok := a.lib.Songs[sid]; ok {
				dur += s.Duration
			}
		}
		entries = append(entries, PlaylistEntry{
			ID:        p.ID,
			Name:      p.Name,
			Comment:   p.Comment,
			SongCount: len(p.SongIDs),
			Duration:  dur,
			Owner:     p.Owner,
			Public:    p.Public,
			Created:   p.Created,
			Changed:   p.Changed,
		})
	}

	resp := newOK()
	resp.Playlists = &PlaylistsResponse{Playlists: entries}
	writeResponse(w, r, resp)
}

func (a *API) getPlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	p := a.store.GetPlaylist(id)
	if p == nil {
		writeResponse(w, r, newError(70, "Playlist not found"))
		return
	}

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	var children []Child
	dur := 0
	for _, sid := range p.SongIDs {
		if s, ok := a.lib.Songs[sid]; ok {
			children = append(children, a.songToChild(s))
			dur += s.Duration
		}
	}

	resp := newOK()
	resp.PlaylistResponse = &PlaylistWithSongs{
		ID:        p.ID,
		Name:      p.Name,
		Comment:   p.Comment,
		SongCount: len(p.SongIDs),
		Duration:  dur,
		Owner:     p.Owner,
		Public:    p.Public,
		Created:   p.Created,
		Changed:   p.Changed,
		Entries:   children,
	}
	writeResponse(w, r, resp)
}

func (a *API) createPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID := r.FormValue("playlistId")
	name := r.FormValue("name")
	songIDs := r.Form["songId"]

	if playlistID != "" {
		p := a.store.GetPlaylist(playlistID)
		if p != nil {
			if name != "" {
				p.Name = name
			}
			if len(songIDs) > 0 {
				p.SongIDs = songIDs
			}
			a.store.SavePlaylist(p)
			a.getPlaylist2(w, r, p.ID)
			return
		}
	}

	if name == "" {
		name = "New Playlist"
	}
	p := a.store.CreatePlaylist(name, a.cfg.User, songIDs)
	a.getPlaylist2(w, r, p.ID)
}

func (a *API) getPlaylist2(w http.ResponseWriter, r *http.Request, id string) {
	p := a.store.GetPlaylist(id)
	if p == nil {
		writeResponse(w, r, newOK())
		return
	}

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	var children []Child
	dur := 0
	for _, sid := range p.SongIDs {
		if s, ok := a.lib.Songs[sid]; ok {
			children = append(children, a.songToChild(s))
			dur += s.Duration
		}
	}

	resp := newOK()
	resp.PlaylistResponse = &PlaylistWithSongs{
		ID:        p.ID,
		Name:      p.Name,
		Comment:   p.Comment,
		SongCount: len(p.SongIDs),
		Duration:  dur,
		Owner:     p.Owner,
		Public:    p.Public,
		Created:   p.Created,
		Changed:   p.Changed,
		Entries:   children,
	}
	writeResponse(w, r, resp)
}

func (a *API) updatePlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("playlistId")
	p := a.store.GetPlaylist(id)
	if p == nil {
		writeResponse(w, r, newError(70, "Playlist not found"))
		return
	}

	if name := r.FormValue("name"); name != "" {
		p.Name = name
	}
	if comment := r.FormValue("comment"); comment != "" {
		p.Comment = comment
	}

	// Remove songs by index (process in reverse to preserve indices)
	removeIndices := r.Form["songIndexToRemove"]
	if len(removeIndices) > 0 {
		indices := make([]int, 0, len(removeIndices))
		for _, s := range removeIndices {
			if idx, err := strconv.Atoi(s); err == nil {
				indices = append(indices, idx)
			}
		}
		sort.Sort(sort.Reverse(sort.IntSlice(indices)))
		for _, idx := range indices {
			if idx >= 0 && idx < len(p.SongIDs) {
				p.SongIDs = append(p.SongIDs[:idx], p.SongIDs[idx+1:]...)
			}
		}
	}

	// Add songs
	addIDs := r.Form["songIdToAdd"]
	p.SongIDs = append(p.SongIDs, addIDs...)

	a.store.SavePlaylist(p)
	writeResponse(w, r, newOK())
}

func (a *API) deletePlaylist(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	a.store.DeletePlaylist(id)
	writeResponse(w, r, newOK())
}

// Star/Unstar

func (a *API) star(w http.ResponseWriter, r *http.Request) {
	songIDs := r.Form["id"]
	albumIDs := r.Form["albumId"]
	artistIDs := r.Form["artistId"]
	a.store.Star(songIDs, albumIDs, artistIDs)
	writeResponse(w, r, newOK())
}

func (a *API) unstar(w http.ResponseWriter, r *http.Request) {
	songIDs := r.Form["id"]
	albumIDs := r.Form["albumId"]
	artistIDs := r.Form["artistId"]
	a.store.Unstar(songIDs, albumIDs, artistIDs)
	writeResponse(w, r, newOK())
}

func (a *API) getStarred(w http.ResponseWriter, r *http.Request) {
	starred := a.store.GetStarred()

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	var artists []Child
	for id, t := range starred.Artists {
		if ar, ok := a.lib.Artists[id]; ok {
			artists = append(artists, Child{ID: ar.ID, IsDir: true, Title: ar.Name, Starred: t})
		}
	}
	var albums []Child
	for id, t := range starred.Albums {
		if al, ok := a.lib.Albums[id]; ok {
			coverArt := ""
			if al.CoverPath != "" {
				coverArt = al.ID
			}
			albums = append(albums, Child{
				ID: al.ID, Parent: al.ArtistID, IsDir: true,
				Title: al.Name, Artist: al.ArtistName, CoverArt: coverArt, Year: al.Year, Starred: t,
			})
		}
	}
	var songs []Child
	for id, t := range starred.Songs {
		if s, ok := a.lib.Songs[id]; ok {
			c := a.songToChild(s)
			c.Starred = t
			songs = append(songs, c)
		}
	}

	resp := newOK()
	resp.Starred = &StarredResponse{Artists: artists, Albums: albums, Songs: songs}
	writeResponse(w, r, resp)
}

func (a *API) getStarred2(w http.ResponseWriter, r *http.Request) {
	starred := a.store.GetStarred()

	a.lib.mu.RLock()
	defer a.lib.mu.RUnlock()

	var artists []ArtistID3
	for id, t := range starred.Artists {
		if ar, ok := a.lib.Artists[id]; ok {
			artists = append(artists, ArtistID3{ID: ar.ID, Name: ar.Name, AlbumCount: len(ar.AlbumIDs), Starred: t})
		}
	}
	var albums []AlbumID3
	for id, t := range starred.Albums {
		if al, ok := a.lib.Albums[id]; ok {
			a3 := a.albumToID3(al)
			a3.Starred = t
			albums = append(albums, a3)
		}
	}
	var songs []Child
	for id, t := range starred.Songs {
		if s, ok := a.lib.Songs[id]; ok {
			c := a.songToChild(s)
			c.Starred = t
			songs = append(songs, c)
		}
	}

	resp := newOK()
	resp.Starred2 = &Starred2Response{Artists: artists, Albums: albums, Songs: songs}
	writeResponse(w, r, resp)
}

// Scrobble

func (a *API) scrobble(w http.ResponseWriter, r *http.Request) {
	ids := r.Form["id"]
	for _, id := range ids {
		a.store.Scrobble(id)
	}
	writeResponse(w, r, newOK())
}

// Scan

func (a *API) startScan(w http.ResponseWriter, r *http.Request) {
	if a.scanning.Load() {
		a.lib.mu.RLock()
		count := len(a.lib.Songs)
		a.lib.mu.RUnlock()
		resp := newOK()
		resp.ScanStatus = &ScanStatus{Scanning: true, Count: count}
		writeResponse(w, r, resp)
		return
	}

	a.scanning.Store(true)
	go func() {
		defer a.scanning.Store(false)
		log.Printf("rescan starting...")
		newLib, err := fullScan(a.cfg.MusicDir)
		if err != nil {
			log.Printf("rescan failed: %v", err)
			return
		}

		a.lib.mu.Lock()
		a.lib.Artists = newLib.Artists
		a.lib.Albums = newLib.Albums
		a.lib.Songs = newLib.Songs
		a.lib.SortedArtists = newLib.SortedArtists
		a.lib.mu.Unlock()

		log.Printf("rescan complete: %d artists, %d albums, %d songs",
			len(newLib.Artists), len(newLib.Albums), len(newLib.Songs))
	}()

	a.lib.mu.RLock()
	count := len(a.lib.Songs)
	a.lib.mu.RUnlock()
	resp := newOK()
	resp.ScanStatus = &ScanStatus{Scanning: true, Count: count}
	writeResponse(w, r, resp)
}

func (a *API) getScanStatus(w http.ResponseWriter, r *http.Request) {
	a.lib.mu.RLock()
	count := len(a.lib.Songs)
	a.lib.mu.RUnlock()
	resp := newOK()
	resp.ScanStatus = &ScanStatus{Scanning: a.scanning.Load(), Count: count}
	writeResponse(w, r, resp)
}

// Helpers

func (a *API) songToChild(s *Song) Child {
	coverArt := ""
	if al, ok := a.lib.Albums[s.AlbumID]; ok && al.CoverPath != "" {
		coverArt = al.ID
	}
	relPath, _ := filepath.Rel(a.cfg.MusicDir, s.Path)
	return Child{
		ID:          s.ID,
		Parent:      s.AlbumID,
		IsDir:       false,
		Title:       s.Title,
		Album:       s.AlbumName,
		Artist:      s.ArtistName,
		Track:       s.Track,
		Year:        s.Year,
		CoverArt:    coverArt,
		Size:        s.Size,
		ContentType: s.ContentType,
		Suffix:      s.Suffix,
		Duration:    s.Duration,
		BitRate:     s.BitRate,
		Path:        relPath,
		Type:        "music",
		AlbumID:     s.AlbumID,
		ArtistID:    s.ArtistID,
		Starred:     a.store.StarredTime(s.ID),
	}
}

func (a *API) albumToID3(al *Album) AlbumID3 {
	coverArt := ""
	if al.CoverPath != "" {
		coverArt = al.ID
	}
	return AlbumID3{
		ID:        al.ID,
		Name:      al.Name,
		Artist:    al.ArtistName,
		ArtistID:  al.ArtistID,
		CoverArt:  coverArt,
		SongCount: len(al.SongIDs),
		Duration:  al.Duration,
		Year:      al.Year,
		Starred:   a.store.StarredTime(al.ID),
	}
}

func intParam(r *http.Request, name string, def int) int {
	v := r.FormValue(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
