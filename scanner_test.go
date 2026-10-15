package main

import "testing"

func TestResolveAlbumArtist(t *testing.T) {
	cases := []struct {
		name                  string
		tagged, track, folder string
		compilation           bool
		want                  string
	}{
		{"album artist tag wins over a featured track artist", "Mumford & Sons", "Mumford & Sons & Chris Stapleton", "Mumford & Sons", false, "Mumford & Sons"},
		{"tag is trimmed", "  Nirvana ", "Nirvana", "Nirvana", false, "Nirvana"},
		{"compilation flag without tag", "", "Bad Bunny", "Compilations", true, variousArtists},
		{"Compilations folder without flag", "", "Bad Bunny", "Compilations", false, variousArtists},
		{"no tag: the artist folder groups the album", "", "Kendrick Lamar feat. Drake", "Kendrick Lamar", false, "Kendrick Lamar"},
		{"loose file at the root: track artist", "", "Unknown Artist", "", false, "Unknown Artist"},
		{"explicit tag beats the Compilations folder", "Various Artists", "Bad Bunny", "Compilations", true, "Various Artists"},
	}
	for _, c := range cases {
		if got := resolveAlbumArtist(c.tagged, c.track, c.folder, c.compilation); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestIsCompilation(t *testing.T) {
	cases := []struct {
		raw  map[string]interface{}
		want bool
	}{
		{map[string]interface{}{"compilation": 1}, true},  // MP4 cpil
		{map[string]interface{}{"compilation": 0}, false}, // MP4 cpil off
		{map[string]interface{}{"TCMP": "1"}, true},       // ID3
		{map[string]interface{}{"COMPILATION": "true"}, true},
		{map[string]interface{}{"compilation": true}, true},
		{map[string]interface{}{}, false},
		{nil, false},
	}
	for i, c := range cases {
		if got := isCompilation(c.raw); got != c.want {
			t.Errorf("case %d (%v): got %v, want %v", i, c.raw, got, c.want)
		}
	}
}
