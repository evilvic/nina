package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

const (
	apiVersion    = "1.16.1"
	serverType    = "nina"
	serverVersion = "0.2.0"
	xmlns         = "http://subsonic.org/restapi"
)

type SubsonicResponse struct {
	XMLName       xml.Name `xml:"subsonic-response" json:"-"`
	XMLNS         string   `xml:"xmlns,attr" json:"-"`
	Status        string   `xml:"status,attr" json:"status"`
	Version       string   `xml:"version,attr" json:"version"`
	Type          string   `xml:"type,attr" json:"type"`
	ServerVersion string   `xml:"serverVersion,attr" json:"serverVersion"`

	Error            *APIError            `xml:"error,omitempty" json:"error,omitempty"`
	License          *License             `xml:"license,omitempty" json:"license,omitempty"`
	MusicFolders     *MusicFolders        `xml:"musicFolders,omitempty" json:"musicFolders,omitempty"`
	Indexes          *Indexes             `xml:"indexes,omitempty" json:"indexes,omitempty"`
	Artists          *ArtistsID3          `xml:"artists,omitempty" json:"artists,omitempty"`
	ArtistResponse   *ArtistWithAlbumsID3 `xml:"artist,omitempty" json:"artist,omitempty"`
	AlbumResponse    *AlbumWithSongsID3   `xml:"album,omitempty" json:"album,omitempty"`
	SongResponse     *Child               `xml:"song,omitempty" json:"song,omitempty"`
	SearchResult2    *SearchResult2       `xml:"searchResult2,omitempty" json:"searchResult2,omitempty"`
	SearchResult3    *SearchResult3       `xml:"searchResult3,omitempty" json:"searchResult3,omitempty"`
	AlbumList        *AlbumListResponse   `xml:"albumList,omitempty" json:"albumList,omitempty"`
	AlbumList2       *AlbumList2Response  `xml:"albumList2,omitempty" json:"albumList2,omitempty"`
	RandomSongs      *SongsResponse       `xml:"randomSongs,omitempty" json:"randomSongs,omitempty"`
	User             *UserResponse        `xml:"user,omitempty" json:"user,omitempty"`
	Playlists        *PlaylistsResponse   `xml:"playlists,omitempty" json:"playlists,omitempty"`
	PlaylistResponse *PlaylistWithSongs   `xml:"playlist,omitempty" json:"playlist,omitempty"`
	Starred          *StarredResponse     `xml:"starred,omitempty" json:"starred,omitempty"`
	Starred2         *Starred2Response    `xml:"starred2,omitempty" json:"starred2,omitempty"`
	Directory        *Directory           `xml:"directory,omitempty" json:"directory,omitempty"`
	ScanStatus       *ScanStatus          `xml:"scanStatus,omitempty" json:"scanStatus,omitempty"`
}

type APIError struct {
	Code    int    `xml:"code,attr" json:"code"`
	Message string `xml:"message,attr" json:"message"`
}

type License struct {
	Valid bool `xml:"valid,attr" json:"valid"`
}

type MusicFolders struct {
	Folders []MusicFolder `xml:"musicFolder" json:"musicFolder"`
}

type MusicFolder struct {
	ID   int    `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr" json:"name"`
}

type Indexes struct {
	LastModified int64   `xml:"lastModified,attr" json:"lastModified"`
	Index        []Index `xml:"index" json:"index"`
}

type Index struct {
	Name    string  `xml:"name,attr" json:"name"`
	Artists []Child `xml:"artist" json:"artist"`
}

type ArtistsID3 struct {
	Index []IndexID3 `xml:"index" json:"index"`
}

type IndexID3 struct {
	Name    string      `xml:"name,attr" json:"name"`
	Artists []ArtistID3 `xml:"artist" json:"artist"`
}

type ArtistID3 struct {
	ID         string `xml:"id,attr" json:"id"`
	Name       string `xml:"name,attr" json:"name"`
	AlbumCount int    `xml:"albumCount,attr" json:"albumCount"`
	Starred    string `xml:"starred,attr,omitempty" json:"starred,omitempty"`
}

type ArtistWithAlbumsID3 struct {
	ID         string     `xml:"id,attr" json:"id"`
	Name       string     `xml:"name,attr" json:"name"`
	AlbumCount int        `xml:"albumCount,attr" json:"albumCount"`
	Starred    string     `xml:"starred,attr,omitempty" json:"starred,omitempty"`
	Albums     []AlbumID3 `xml:"album" json:"album"`
}

type AlbumID3 struct {
	ID        string `xml:"id,attr" json:"id"`
	Name      string `xml:"name,attr" json:"name"`
	Artist    string `xml:"artist,attr" json:"artist"`
	ArtistID  string `xml:"artistId,attr" json:"artistId"`
	CoverArt  string `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	SongCount int    `xml:"songCount,attr" json:"songCount"`
	Duration  int    `xml:"duration,attr" json:"duration"`
	Year      int    `xml:"year,attr,omitempty" json:"year,omitempty"`
	Starred   string `xml:"starred,attr,omitempty" json:"starred,omitempty"`
}

type AlbumWithSongsID3 struct {
	ID        string  `xml:"id,attr" json:"id"`
	Name      string  `xml:"name,attr" json:"name"`
	Artist    string  `xml:"artist,attr" json:"artist"`
	ArtistID  string  `xml:"artistId,attr" json:"artistId"`
	CoverArt  string  `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	SongCount int     `xml:"songCount,attr" json:"songCount"`
	Duration  int     `xml:"duration,attr" json:"duration"`
	Year      int     `xml:"year,attr,omitempty" json:"year,omitempty"`
	Starred   string  `xml:"starred,attr,omitempty" json:"starred,omitempty"`
	Songs     []Child `xml:"song" json:"song"`
}

type Child struct {
	ID          string `xml:"id,attr" json:"id"`
	Parent      string `xml:"parent,attr,omitempty" json:"parent,omitempty"`
	IsDir       bool   `xml:"isDir,attr" json:"isDir"`
	Title       string `xml:"title,attr" json:"title"`
	Album       string `xml:"album,attr,omitempty" json:"album,omitempty"`
	Artist      string `xml:"artist,attr,omitempty" json:"artist,omitempty"`
	Track       int    `xml:"track,attr,omitempty" json:"track,omitempty"`
	Year        int    `xml:"year,attr,omitempty" json:"year,omitempty"`
	CoverArt    string `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	Size        int64  `xml:"size,attr,omitempty" json:"size,omitempty"`
	ContentType string `xml:"contentType,attr,omitempty" json:"contentType,omitempty"`
	Suffix      string `xml:"suffix,attr,omitempty" json:"suffix,omitempty"`
	Duration    int    `xml:"duration,attr,omitempty" json:"duration,omitempty"`
	BitRate     int    `xml:"bitRate,attr,omitempty" json:"bitRate,omitempty"`
	Path        string `xml:"path,attr,omitempty" json:"path,omitempty"`
	Type        string `xml:"type,attr,omitempty" json:"type,omitempty"`
	AlbumID     string `xml:"albumId,attr,omitempty" json:"albumId,omitempty"`
	ArtistID    string `xml:"artistId,attr,omitempty" json:"artistId,omitempty"`
	Starred     string `xml:"starred,attr,omitempty" json:"starred,omitempty"`
}

type SearchResult2 struct {
	Artists []Child `xml:"artist,omitempty" json:"artist,omitempty"`
	Albums  []Child `xml:"album,omitempty" json:"album,omitempty"`
	Songs   []Child `xml:"song,omitempty" json:"song,omitempty"`
}

type SearchResult3 struct {
	Artists []ArtistID3 `xml:"artist,omitempty" json:"artist,omitempty"`
	Albums  []AlbumID3  `xml:"album,omitempty" json:"album,omitempty"`
	Songs   []Child     `xml:"song,omitempty" json:"song,omitempty"`
}

type AlbumListResponse struct {
	Albums []Child `xml:"album" json:"album"`
}

type AlbumList2Response struct {
	Albums []AlbumID3 `xml:"album" json:"album"`
}

type SongsResponse struct {
	Songs []Child `xml:"song" json:"song"`
}

type UserResponse struct {
	Username     string `xml:"username,attr" json:"username"`
	AdminRole    bool   `xml:"adminRole,attr" json:"adminRole"`
	StreamRole   bool   `xml:"streamRole,attr" json:"streamRole"`
	DownloadRole bool   `xml:"downloadRole,attr" json:"downloadRole"`
	ScrobbleRole bool   `xml:"scrobbleRole,attr" json:"scrobbleRole"`
	ShareRole    bool   `xml:"shareRole,attr" json:"shareRole"`
	CoverArtRole bool   `xml:"coverArtRole,attr" json:"coverArtRole"`
	CommentRole  bool   `xml:"commentRole,attr" json:"commentRole"`
	PodcastRole  bool   `xml:"podcastRole,attr" json:"podcastRole"`
	UploadRole   bool   `xml:"uploadRole,attr" json:"uploadRole"`
	PlaylistRole bool   `xml:"playlistRole,attr" json:"playlistRole"`
	JukeboxRole  bool   `xml:"jukeboxRole,attr" json:"jukeboxRole"`
	SettingsRole bool   `xml:"settingsRole,attr" json:"settingsRole"`
}

type PlaylistsResponse struct {
	Playlists []PlaylistEntry `xml:"playlist,omitempty" json:"playlist,omitempty"`
}

type PlaylistEntry struct {
	ID        string `xml:"id,attr" json:"id"`
	Name      string `xml:"name,attr" json:"name"`
	Comment   string `xml:"comment,attr,omitempty" json:"comment,omitempty"`
	SongCount int    `xml:"songCount,attr" json:"songCount"`
	Duration  int    `xml:"duration,attr" json:"duration"`
	Owner     string `xml:"owner,attr" json:"owner"`
	Public    bool   `xml:"public,attr" json:"public"`
	Created   string `xml:"created,attr" json:"created"`
	Changed   string `xml:"changed,attr" json:"changed"`
}

type PlaylistWithSongs struct {
	ID        string  `xml:"id,attr" json:"id"`
	Name      string  `xml:"name,attr" json:"name"`
	Comment   string  `xml:"comment,attr,omitempty" json:"comment,omitempty"`
	SongCount int     `xml:"songCount,attr" json:"songCount"`
	Duration  int     `xml:"duration,attr" json:"duration"`
	Owner     string  `xml:"owner,attr" json:"owner"`
	Public    bool    `xml:"public,attr" json:"public"`
	Created   string  `xml:"created,attr" json:"created"`
	Changed   string  `xml:"changed,attr" json:"changed"`
	Entries   []Child `xml:"entry" json:"entry"`
}

type StarredResponse struct {
	Artists []Child `xml:"artist,omitempty" json:"artist,omitempty"`
	Albums  []Child `xml:"album,omitempty" json:"album,omitempty"`
	Songs   []Child `xml:"song,omitempty" json:"song,omitempty"`
}

type Starred2Response struct {
	Artists []ArtistID3 `xml:"artist,omitempty" json:"artist,omitempty"`
	Albums  []AlbumID3  `xml:"album,omitempty" json:"album,omitempty"`
	Songs   []Child     `xml:"song,omitempty" json:"song,omitempty"`
}

type ScanStatus struct {
	Scanning bool `xml:"scanning,attr" json:"scanning"`
	Count    int  `xml:"count,attr" json:"count"`
}

type Directory struct {
	ID       string  `xml:"id,attr" json:"id"`
	Parent   string  `xml:"parent,attr,omitempty" json:"parent,omitempty"`
	Name     string  `xml:"name,attr" json:"name"`
	Children []Child `xml:"child,omitempty" json:"child,omitempty"`
}

func newOK() *SubsonicResponse {
	return &SubsonicResponse{
		XMLNS:         xmlns,
		Status:        "ok",
		Version:       apiVersion,
		Type:          serverType,
		ServerVersion: serverVersion,
	}
}

func newError(code int, message string) *SubsonicResponse {
	resp := newOK()
	resp.Status = "failed"
	resp.Error = &APIError{Code: code, Message: message}
	return resp
}

func writeResponse(w http.ResponseWriter, r *http.Request, resp *SubsonicResponse) {
	format := r.FormValue("f")
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		wrapper := map[string]*SubsonicResponse{
			"subsonic-response": resp,
		}
		json.NewEncoder(w).Encode(wrapper)
	default:
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(xml.Header))
		xml.NewEncoder(w).Encode(resp)
	}
}
