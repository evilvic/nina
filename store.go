package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Store struct {
	mu      sync.RWMutex
	dataDir string
	data    StoreData
}

type StoreData struct {
	Playlists  map[string]*Playlist `json:"playlists"`
	Starred    StarredData          `json:"starred"`
	PlayCounts map[string]int       `json:"playCounts"`
}

type Playlist struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Comment string   `json:"comment,omitempty"`
	Owner   string   `json:"owner"`
	Public  bool     `json:"public"`
	SongIDs []string `json:"songIds"`
	Created string   `json:"created"`
	Changed string   `json:"changed"`
}

type StarredData struct {
	Artists map[string]string `json:"artists"`
	Albums  map[string]string `json:"albums"`
	Songs   map[string]string `json:"songs"`
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	s := &Store{
		dataDir: dataDir,
		data: StoreData{
			Playlists:  make(map[string]*Playlist),
			Starred:    newStarredData(),
			PlayCounts: make(map[string]int),
		},
	}
	s.load()
	return s, nil
}

func newStarredData() StarredData {
	return StarredData{
		Artists: make(map[string]string),
		Albums:  make(map[string]string),
		Songs:   make(map[string]string),
	}
}

func (s *Store) load() {
	path := filepath.Join(s.dataDir, "nina.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.data)
	if s.data.Playlists == nil {
		s.data.Playlists = make(map[string]*Playlist)
	}
	if s.data.Starred.Artists == nil {
		s.data.Starred.Artists = make(map[string]string)
	}
	if s.data.Starred.Albums == nil {
		s.data.Starred.Albums = make(map[string]string)
	}
	if s.data.Starred.Songs == nil {
		s.data.Starred.Songs = make(map[string]string)
	}
	if s.data.PlayCounts == nil {
		s.data.PlayCounts = make(map[string]int)
	}
}

func (s *Store) save() {
	path := filepath.Join(s.dataDir, "nina.json")
	data, _ := json.MarshalIndent(s.data, "", "  ")
	os.WriteFile(path, data, 0644)
}

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05")
}

// Playlists

func (s *Store) GetPlaylists() []*Playlist {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Playlist, 0, len(s.data.Playlists))
	for _, p := range s.data.Playlists {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (s *Store) GetPlaylist(id string) *Playlist {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Playlists[id]
}

func (s *Store) CreatePlaylist(name, owner string, songIDs []string) *Playlist {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := makeID("pl", fmt.Sprintf("%s-%d", name, time.Now().UnixNano()))
	now := nowISO()
	p := &Playlist{
		ID:      id,
		Name:    name,
		Owner:   owner,
		SongIDs: songIDs,
		Created: now,
		Changed: now,
	}
	s.data.Playlists[id] = p
	s.save()
	return p
}

func (s *Store) SavePlaylist(p *Playlist) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.Changed = nowISO()
	s.data.Playlists[p.ID] = p
	s.save()
}

func (s *Store) DeletePlaylist(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data.Playlists, id)
	s.save()
}

// Starred

func (s *Store) Star(songIDs, albumIDs, artistIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := nowISO()
	for _, id := range artistIDs {
		s.data.Starred.Artists[id] = now
	}
	for _, id := range albumIDs {
		s.data.Starred.Albums[id] = now
	}
	for _, id := range songIDs {
		s.data.Starred.Songs[id] = now
	}
	s.save()
}

func (s *Store) Unstar(songIDs, albumIDs, artistIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range artistIDs {
		delete(s.data.Starred.Artists, id)
	}
	for _, id := range albumIDs {
		delete(s.data.Starred.Albums, id)
	}
	for _, id := range songIDs {
		delete(s.data.Starred.Songs, id)
	}
	s.save()
}

func (s *Store) StarredTime(id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.data.Starred.Songs[id]; ok {
		return t
	}
	if t, ok := s.data.Starred.Albums[id]; ok {
		return t
	}
	if t, ok := s.data.Starred.Artists[id]; ok {
		return t
	}
	return ""
}

func (s *Store) GetStarred() StarredData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Starred
}

// Scrobble

func (s *Store) Scrobble(songID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.PlayCounts[songID]++
	s.save()
}
