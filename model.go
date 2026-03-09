package main

import (
	"sort"
	"strings"
	"sync"
	"unicode"
)

type Artist struct {
	ID       string
	Name     string
	AlbumIDs []string
}

type Album struct {
	ID         string
	Name       string
	ArtistID   string
	ArtistName string
	Year       int
	SongIDs    []string
	CoverPath  string
	Duration   int
}

type Song struct {
	ID          string
	Title       string
	AlbumID     string
	AlbumName   string
	ArtistID    string
	ArtistName  string
	Track       int
	Disc        int
	Year        int
	Duration    int
	Size        int64
	Path        string
	Suffix      string
	ContentType string
	BitRate     int
}

type Library struct {
	mu            sync.RWMutex
	Artists       map[string]*Artist
	Albums        map[string]*Album
	Songs         map[string]*Song
	SortedArtists []*Artist
}

func NewLibrary() *Library {
	return &Library{
		Artists: make(map[string]*Artist),
		Albums:  make(map[string]*Album),
		Songs:   make(map[string]*Song),
	}
}

func (l *Library) SortArtists() {
	l.SortedArtists = make([]*Artist, 0, len(l.Artists))
	for _, a := range l.Artists {
		l.SortedArtists = append(l.SortedArtists, a)
	}
	sort.Slice(l.SortedArtists, func(i, j int) bool {
		return strings.ToLower(l.SortedArtists[i].Name) < strings.ToLower(l.SortedArtists[j].Name)
	})
}

func (l *Library) GetAlbumsForArtist(artistID string) []*Album {
	artist, ok := l.Artists[artistID]
	if !ok {
		return nil
	}
	albums := make([]*Album, 0, len(artist.AlbumIDs))
	for _, id := range artist.AlbumIDs {
		if a, ok := l.Albums[id]; ok {
			albums = append(albums, a)
		}
	}
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].Year != albums[j].Year {
			return albums[i].Year < albums[j].Year
		}
		return albums[i].Name < albums[j].Name
	})
	return albums
}

func (l *Library) GetSongsForAlbum(albumID string) []*Song {
	album, ok := l.Albums[albumID]
	if !ok {
		return nil
	}
	songs := make([]*Song, 0, len(album.SongIDs))
	for _, id := range album.SongIDs {
		if s, ok := l.Songs[id]; ok {
			songs = append(songs, s)
		}
	}
	sort.Slice(songs, func(i, j int) bool {
		if songs[i].Disc != songs[j].Disc {
			return songs[i].Disc < songs[j].Disc
		}
		return songs[i].Track < songs[j].Track
	})
	return songs
}

func (l *Library) Search(query string, artistCount, albumCount, songCount int) ([]*Artist, []*Album, []*Song) {
	q := strings.ToLower(query)
	var artists []*Artist
	var albums []*Album
	var songs []*Song

	for _, a := range l.SortedArtists {
		if len(artists) >= artistCount {
			break
		}
		if strings.Contains(strings.ToLower(a.Name), q) {
			artists = append(artists, a)
		}
	}
	for _, a := range l.Albums {
		if len(albums) >= albumCount {
			break
		}
		if strings.Contains(strings.ToLower(a.Name), q) {
			albums = append(albums, a)
		}
	}
	for _, s := range l.Songs {
		if len(songs) >= songCount {
			break
		}
		if strings.Contains(strings.ToLower(s.Title), q) {
			songs = append(songs, s)
		}
	}
	return artists, albums, songs
}

func indexChar(name string) string {
	if len(name) == 0 {
		return "#"
	}
	r := rune(strings.ToUpper(name)[0])
	if unicode.IsLetter(r) {
		return string(r)
	}
	return "#"
}
