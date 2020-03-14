package qbt

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"

	"net/url"
	"strconv"
	"strings"

	wrapper "github.com/pkg/errors"

	"golang.org/x/net/publicsuffix"
)

//ErrBadPriority means the priority is not allowd by qbittorrent
var ErrBadPriority = errors.New("priority not available")

//ErrBadResponse means that qbittorrent sent back an unexpected response
var ErrBadResponse = errors.New("received bad response")

//delimit puts list into a combined (single element) map with all items connected separated by the delimiter
//this is how the WEBUI API recognizes multiple items
func delimit(items []string, delimiter string) (delimited string) {
	for i, v := range items {
		if i > 0 {
			delimited += delimiter + v
		} else {
			delimited = v
		}
	}
	return delimited
}

//Client creates a connection to qbittorrent and performs requests
type Client struct {
	http          *http.Client
	URL           string
	Authenticated bool
	Jar           http.CookieJar
}

//NewClient creates a new client connection to qbittorrent
func NewClient(url string) *Client {
	client := &Client{}

	// ensure url ends with "/"
	if url[len(url)-1:] != "/" {
		url += "/"
	}

	client.URL = url

	// create cookie jar
	client.Jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client.http = &http.Client{
		Jar: client.Jar,
	}
	return client
}

//get will perform a GET request with no parameters
func (client *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", client.URL+endpoint, nil)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to build request")
	}

	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	// add optional parameters that the user wants
	if opts != nil {
		query := req.URL.Query()
		for k, v := range opts {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

//post will perform a POST request with no content-type specified
func (client *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", client.URL+endpoint, nil)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to build request")
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	// add optional parameters that the user wants
	if opts != nil {
		form := url.Values{}
		for k, v := range opts {
			form.Add(k, v)
		}
		req.PostForm = form
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil

}

//postMultipart will perform a multiple part POST request
func (client *Client) postMultipart(endpoint string, buffer bytes.Buffer, contentType string) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", client.URL+endpoint, &buffer)
	if err != nil {
		return nil, wrapper.Wrap(err, "error creating request")
	}

	// add the content-type so qbittorrent knows what to expect
	req.Header.Set("Content-Type", contentType)
	// add user-agent header to allow qbittorrent to identify us
	req.Header.Set("User-Agent", "go-qbittorrent v0.1")

	resp, err = client.http.Do(req)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed to perform request")
	}

	return resp, nil
}

//writeOptions will write a map to the buffer through multipart.NewWriter
func writeOptions(writer *multipart.Writer, opts map[string]string) (err error) {
	for key, val := range opts {
		if err := writer.WriteField(key, val); err != nil {
			return err
		}
	}
	return nil
}

//postMultipartData will perform a multiple part POST request without a file
func (client *Client) postMultipartData(endpoint string, opts map[string]string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// write the options to the buffer
	// will contain the link string
	if err := writeOptions(writer, opts); err != nil {
		return nil, wrapper.Wrap(err, "failed to write options")
	}

	// close the writer before doing request to get closing line on multipart request
	if err := writer.Close(); err != nil {
		return nil, wrapper.Wrap(err, "failed to close writer")
	}

	resp, err := client.postMultipart(endpoint, buffer, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//postMultipartFile will perform a multiple part POST request with a file
func (client *Client) postMultipartFile(endpoint string, fileName string, opts map[string]string) (*http.Response, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		return nil, wrapper.Wrap(err, "error opening file")
	}
	// defer the closing of the file until the end of function
	// so we can still copy its contents
	defer file.Close()

	// create form for writing the file to and give it the filename
	formWriter, err := writer.CreateFormFile("torrents", path.Base(fileName))
	if err != nil {
		return nil, wrapper.Wrap(err, "error adding file")
	}

	// write the options to the buffer
	if err := writeOptions(writer, opts); err != nil {
		return nil, wrapper.Wrap(err, "error writeOptions")
	}

	// copy the file contents into the form
	if _, err = io.Copy(formWriter, file); err != nil {
		return nil, wrapper.Wrap(err, "error copying file")
	}

	// close the writer before doing request to get closing line on multipart request
	if err := writer.Close(); err != nil {
		return nil, wrapper.Wrap(err, "failed to close writer")
	}

	resp, err := client.postMultipart(endpoint, buffer, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Application endpoints

//Login logs you in to the qbittorrent client
//returns the current authentication status
func (client *Client) Login(opts LoginOptions) (err error) {
	params := map[string]string{
		"username": opts.Username,
		"password": opts.Password,
	}

	resp, err := client.post("api/v2/auth/login", params)
	if err != nil {
		return wrapper.Wrap(err, "failed login request")
	} else if resp.StatusCode == 403 {
		return wrapper.Errorf("User's IP is banned for too many failed login attempts")
	}

	// add the cookie to cookie jar to authenticate later requests
	if cookies := resp.Cookies(); len(cookies) > 0 {
		cookieURL, _ := url.Parse("http://localhost:8080")
		client.Jar.SetCookies(cookieURL, cookies)
		// create a new client with the cookie jar and replace the old one
		// so that all our later requests are authenticated
		client.http = &http.Client{
			Jar: client.Jar,
		}
	} else {
		return wrapper.Errorf("Could not get cookie")
	}

	// change authentication status so we know were authenticated in later requests
	client.Authenticated = true

	return nil
}

//Logout logs you out of the qbittorrent client
//returns the current authentication status
func (client *Client) Logout() (err error) {
	_, err = client.get("api/v2/auth/logout", nil)
	if err != nil {
		return wrapper.Wrap(err, "failed logout request")
	}

	// change authentication status so we know were not authenticated in later requests
	client.Authenticated = false

	return nil
}

//ApplicationVersion of the qbittorrent client
func (client *Client) ApplicationVersion() (version string, err error) {
	resp, err := client.get("api/v2/app/version", nil)
	if err != nil {
		return version, wrapper.Wrap(err, "failed version request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return version, wrapper.Wrap(err, "failed decoding version response")
	}

	return version, err
}

//WebAPIVersion of the qbittorrent client
func (client *Client) WebAPIVersion() (version string, err error) {
	resp, err := client.get("api/v2/app/webapiVersion", nil)
	if err != nil {
		return version, wrapper.Wrap(err, "failed webapiVersion request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return version, wrapper.Wrap(err, "failed decoding webapiVersion response")
	}

	return version, err
}

//BuildInfo of the qbittorrent client
func (client *Client) BuildInfo() (buildInfo BuildInfo, err error) {
	resp, err := client.get("api/v2/app/buildInfo", nil)
	if err != nil {
		return buildInfo, wrapper.Wrap(err, "failed buildInfo request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&buildInfo); err != nil {
		return buildInfo, wrapper.Wrap(err, "failed decoding buildInfo response")
	}

	return buildInfo, err
}

//Preferences of the qbittorrent client
func (client *Client) Preferences() (prefs Preferences, err error) {
	resp, err := client.get("api/v2/app/preferences", nil)
	if err != nil {
		return prefs, wrapper.Wrap(err, "failed preferences request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return prefs, wrapper.Wrap(err, "failed decoding preferences response")
	}

	return prefs, err
}

//SetPreferences of the qbittorrent client
func (client *Client) SetPreferences() (prefsSet bool, err error) {
	resp, err := client.post("api/v2/app/setPreferences", nil)
	if err != nil {
		return false, wrapper.Wrap(err, "failed setPreferences request")
	}

	return resp.Status == "200 OK", err
}

//DefaultSavePath of the qbittorrent client
func (client *Client) DefaultSavePath() (path string, err error) {
	resp, err := client.get("api/v2/app/defaultSavePath", nil)
	if err != nil {
		return path, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&path); err != nil {
		return path, wrapper.Wrap(err, "failed decoding defaultSavePath response")
	}
	return path, err
}

//Shutdown shuts down the qbittorrent client
func (client *Client) Shutdown() (shuttingDown bool, err error) {
	resp, err := client.get("api/v2/app/shutdown", nil)
	if err != nil {
		return false, wrapper.Wrap(err, "failed shutdown request")
	}

	// return true if successful
	return resp.Status == "200 OK", err
}

// Log Endpoints

//Logs of the qbittorrent client
func (client *Client) Logs(filters map[string]string) (logs []Log, err error) {
	resp, err := client.get("api/v2/log/main", filters)
	if err != nil {
		return logs, wrapper.Wrap(err, "failed logs request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return logs, wrapper.Wrap(err, "failed decoding logs response")
	}

	return logs, err
}

//PeerLogs of the qbittorrent client
func (client *Client) PeerLogs(filters map[string]string) (logs []PeerLog, err error) {
	resp, err := client.get("api/v2/log/peers", filters)
	if err != nil {
		return logs, wrapper.Wrap(err, "failed peerLogs request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return logs, wrapper.Wrap(err, "failed decoding peerLogs response")
	}

	return logs, err
}

// TODO: Sync Endpoints

// TODO: Transfer Endpoints

//Info returns info you usually see in qBt status bar.
func (client *Client) Info() (info Info, err error) {
	resp, err := client.get("api/v2/transfer/info", nil)
	if err != nil {
		return info, wrapper.Wrap(err, "failed transferInfo request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return info, wrapper.Wrap(err, "failed decoding transferInfo response")
	}

	return info, err
}

//AltSpeedLimitsEnabled returns info you usually see in qBt status bar.
func (client *Client) AltSpeedLimitsEnabled() (mode bool, err error) {
	resp, err := client.get("api/v2/transfer/speedLimitsMode", nil)
	if err != nil {
		return mode, wrapper.Wrap(err, "failed speedLimitsMode request")
	}

	var decoded int
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return mode, wrapper.Wrap(err, "failed decoding speedLimitsMode response")
	}

	mode = decoded == 1
	return mode, err
}

//ToggleAltSpeedLimits returns info you usually see in qBt status bar.
func (client *Client) ToggleAltSpeedLimits() (toggled bool, err error) {
	resp, err := client.get("api/v2/transfer/toggleSpeedLimitsMode", nil)
	if err != nil {
		return toggled, wrapper.Wrap(err, "failed toggleSpeedLimitsMode request")
	}

	return resp.Status == "200 OK", err
}

//DlLimit returns info you usually see in qBt status bar.
func (client *Client) DlLimit() (dlLimit int, err error) {
	resp, err := client.get("api/v2/transfer/downloadLimit", nil)
	if err != nil {
		return dlLimit, wrapper.Wrap(err, "failed downloadLimit request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&dlLimit); err != nil {
		return dlLimit, wrapper.Wrap(err, "failed decoding downloadLimit response")
	}

	return dlLimit, err
}

//SetDlLimit returns info you usually see in qBt status bar.
func (client *Client) SetDlLimit(limit int) (set bool, err error) {
	params := map[string]string{"limit": strconv.Itoa(limit)}
	resp, err := client.get("api/v2/transfer/setDownloadLimit", params)
	if err != nil {
		return set, wrapper.Wrap(err, "failed setDownloadLimit request")
	}

	return resp.Status == "200 OK", err
}

//UlLimit returns info you usually see in qBt status bar.
func (client *Client) UlLimit() (ulLimit int, err error) {
	resp, err := client.get("api/v2/transfer/uploadLimit", nil)
	if err != nil {
		return ulLimit, wrapper.Wrap(err, "failed uploadLimit request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&ulLimit); err != nil {
		return ulLimit, wrapper.Wrap(err, "failed decoding uploadLimit response")
	}

	return ulLimit, err
}

//SetUlLimit returns info you usually see in qBt status bar.
func (client *Client) SetUlLimit(limit int) (set bool, err error) {
	params := map[string]string{"limit": strconv.Itoa(limit)}
	resp, err := client.get("api/v2/transfer/setUploadLimit", params)
	if err != nil {
		return set, wrapper.Wrap(err, "failed setUploadLimit request")
	}

	return resp.Status == "200 OK", err
}

//Torrents returns a list of all torrents in qbittorrent matching your filter
func (client *Client) Torrents(filters map[string]string) (torrentList []BasicTorrent, err error) {
	resp, err := client.get("api/v2/torrents/info", filters)
	if err != nil {
		return torrentList, wrapper.Wrap(err, "failed torrentsInfo request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&torrentList); err != nil {
		return torrentList, wrapper.Wrap(err, "failed decoding torrentsInfo response")
	}

	return torrentList, nil
}

//Torrent returns a specific torrent matching the hash
func (client *Client) Torrent(hash string) (torrent Torrent, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := client.get("api/v2/torrents/properties", opts)
	if err != nil {
		return torrent, wrapper.Wrap(err, "failed torrentsProperties request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&torrent); err != nil {
		return torrent, wrapper.Wrap(err, "failed decoding torrentsProperties response")
	}

	return torrent, nil
}

//TorrentTrackers returns all trackers for a specific torrent matching the hash
func (client *Client) TorrentTrackers(hash string) (trackers []Tracker, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := client.get("api/v2/torrents/trackers", opts)
	if err != nil {
		return trackers, wrapper.Wrap(err, "failed torrentsTrackers request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&trackers); err != nil {
		return trackers, wrapper.Wrap(err, "failed decoding torrentsTrackers response")
	}

	return trackers, nil
}

//TorrentWebSeeds returns seeders for a specific torrent matching the hash
func (client *Client) TorrentWebSeeds(hash string) (webSeeds []WebSeed, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := client.get("api/v2/torrents/webseeds", opts)
	if err != nil {
		return webSeeds, wrapper.Wrap(err, "failed torrentsWebseeds request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&webSeeds); err != nil {
		return webSeeds, wrapper.Wrap(err, "failed decoding torrentsWebseeds response")
	}

	return webSeeds, nil
}

//TorrentFiles from given hash
func (client *Client) TorrentFiles(hash string) (files []TorrentFile, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := client.get("api/v2/torrents/files", opts)
	if err != nil {
		return files, wrapper.Wrap(err, "failed torrentsFiles request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return files, wrapper.Wrap(err, "failed decoding torrentsFiles response")
	}

	return files, nil
}

//TorrentPieceStates for all pieces of torrent
func (client *Client) TorrentPieceStates(hash string) (states []int, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := client.get("api/v2/torrents/pieceStates", opts)
	if err != nil {
		return states, wrapper.Wrap(err, "failed torrentsPieceStates request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return states, wrapper.Wrap(err, "failed decoding torrentsPieceStates response")
	}

	return states, nil
}

//TorrentPieceHashes for all pieces of torrent
func (client *Client) TorrentPieceHashes(hash string) (hashes []string, err error) {
	var opts = map[string]string{"hash": strings.ToLower(hash)}
	resp, err := client.get("api/v2/torrents/pieceHashes", opts)
	if err != nil {
		return hashes, wrapper.Wrap(err, "failed torrentsPieceHashes request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&hashes); err != nil {
		return hashes, wrapper.Wrap(err, "failed decoding torrentsPieceHashes response")
	}

	return hashes, nil
}

//Pause torrents
func (client *Client) Pause(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/pause", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsPause request")
	}

	return resp.StatusCode == 200, nil
}

//Resume torrents
func (client *Client) Resume(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/resume", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsResume request")
	}

	return resp.StatusCode == 200, nil
}

//Delete torrents and optionally delete their files
func (client *Client) Delete(hashes []string, deleteFiles bool) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	opts["deleteFiles"] = strconv.FormatBool(deleteFiles)
	resp, err := client.get("api/v2/torrents/delete", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsDelete request")
	}

	return resp.StatusCode == 200, nil
}

//Recheck torrents
func (client *Client) Recheck(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/recheck", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsRecheck request")
	}

	return resp.StatusCode == 200, nil
}

//Reannounce torrents
func (client *Client) Reannounce(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/reannounce", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsReannounce request")
	}

	return resp.StatusCode == 200, nil
}

//DownloadFromLink starts downloading a torrent from a link
func (client *Client) DownloadFromLink(link string, opts map[string]string) (*http.Response, error) {
	opts["urls"] = link
	resp, err := client.postMultipartData("api/v2/torrents/add", opts)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed torrentsAdd request")
	}

	return resp, nil
}

//DownloadFromFile starts downloading a torrent from a file
func (client *Client) DownloadFromFile(file string, options map[string]string) (bool, error) {
	resp, err := client.postMultipartFile("api/v2/torrents/add", file, options)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsAdd request")
	} else if resp.StatusCode == 415 {
		err = wrapper.Errorf("Torrent file is not valid")
	}

	return resp.StatusCode == 200, nil
}

//AddTrackers to a torrent
func (client *Client) AddTrackers(opts AddTrackersOptions) error {
	params := make(map[string]string)
	params["hash"] = strings.ToLower(opts.Hash)
	params["urls"] = delimit(opts.Trackers, "%0A") // add escaping for ampersand in urls

	resp, err := client.post("api/v2/torrents/addTrackers", params)
	if err != nil {
		return wrapper.Wrap(err, "failed addTrackers request")
	} else if resp != nil && (*resp).StatusCode == 404 {
		return wrapper.New("torrent hash not found")
	}
	return nil
}

//EditTracker on a torrent
func (client *Client) EditTracker(opts EditTrackerOptions) error {
	params := map[string]string{
		"hash":    opts.Hash,
		"origUrl": opts.OrigURL,
		"newUrl":  opts.NewURL,
	}
	resp, err := client.get("api/v2/torrents/editTracker", params)
	if err != nil {
		return wrapper.Wrap(err, "failed editTracker request")
	}
	switch sc := (*resp).StatusCode; sc {
	case 400:
		return wrapper.New("newUrl is not a valid url")
	case 404:
		return wrapper.New("Torrent hash was not found")
	case 409:
		return wrapper.New("newUrl already exists for this torrent or origUrl was not found")
	default:
		return nil
	}
}

//RemoveTrackers from a torrent
func (client *Client) RemoveTrackers(opts RemoveTrackersOptions) (bool, error) {
	params := map[string]string{
		"hash": opts.Hash,
		"urls": delimit(opts.Trackers, "|"),
	}
	resp, err := client.get("api/v2/torrents/removeTrackers", params)
	if err != nil {
		return false, wrapper.Wrap(err, "failed removeTrackers request")
	}

	switch sc := (*resp).StatusCode; sc {
	case 404:
		return false, wrapper.New("torrent hash was not found")
	case 409:
		return false, wrapper.New("all URLs were not found")
	default:
		break
	}

	return true, nil
}

//IncreasePriority of torrents
func (client *Client) IncreasePriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/IncreasePrio", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsIncreasePrio request")
	}

	switch sc := (*resp).StatusCode; sc {
	case 409:
		return false, wrapper.New("torrent queueing not enabled")
	default:
		break
	}

	return true, nil
}

//DecreasePriority of torrents
func (client *Client) DecreasePriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/DecreasePrio", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsDecreasePrio request")
	}

	switch sc := (*resp).StatusCode; sc {
	case 409:
		return false, wrapper.New("torrent queueing not enabled")
	default:
		break
	}

	return true, nil
}

//MaxPriority maximizes the priority of torrents
func (client *Client) MaxPriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/TopPrio", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsTopPrio request")
	}

	switch sc := (*resp).StatusCode; sc {
	case 409:
		return false, wrapper.New("torrent queueing not enabled")
	default:
		break
	}

	return true, nil
}

//MinPriority maximizes the priority of torrents
func (client *Client) MinPriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/BottomPrio", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsBottomPrio request")
	}

	switch sc := (*resp).StatusCode; sc {
	case 409:
		return false, wrapper.New("torrent queueing not enabled")
	default:
		break
	}

	return true, nil
}

//FilePriority for a torrent
func (client *Client) FilePriority(hash string, ids []string, priority int) (bool, error) {
	opts := map[string]string{
		"hashes":   hash,
		"id":       delimit(ids, "|"),
		"priority": strconv.Itoa(priority),
	}
	resp, err := client.get("api/v2/torrents/filePrio", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsFilePrio request")
	}

	switch sc := (*resp).StatusCode; sc {
	case 400:
		return false, wrapper.New("priority is invalid")
	case 404:
		return false, wrapper.New("torrent hash not found")
	case 409:
		return false, wrapper.New("torrent metadata hasnt downloaded yet")
	default:
		break
	}

	return true, nil
}

//GetTorrentDownloadLimit for a list of torrents
func (client *Client) GetTorrentDownloadLimit(hashes []string) (limits map[string]string, err error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.post("api/v2/torrents/downloadLimit", opts)
	if err != nil {
		return limits, wrapper.Wrap(err, "failed torrentsDownloadLimit request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return limits, wrapper.Wrap(err, "failed decoding torrentsDownloadLimit response")
	}

	return limits, nil
}

//SetTorrentDownloadLimit for a list of torrents
func (client *Client) SetTorrentDownloadLimit(hashes []string, limit string) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"limit":  limit,
	}
	resp, err := client.post("api/v2/torrents/setDownloadLimit", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetDownloadLimit request")
	}

	return resp.StatusCode == 200, nil
}

//SetTorrentShareLimit for a list of torrents
func (client *Client) SetTorrentShareLimit(hashes []string, ratioLimit string, seedingTimeLimit string) (bool, error) {
	opts := map[string]string{
		"hashes":           delimit(hashes, "|"),
		"ratioLimit":       ratioLimit,
		"seedingTimeLimit": seedingTimeLimit,
	}
	resp, err := client.post("api/v2/torrents/setShareLimits", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetShareLimits request")
	}

	return resp.StatusCode == 200, nil
}

//GetTorrentUploadLimit for a list of torrents
func (client *Client) GetTorrentUploadLimit(hashes []string) (limits map[string]string, err error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.post("api/v2/torrents/uploadLimit", opts)
	if err != nil {
		return limits, wrapper.Wrap(err, "failed torrentsUploadLimit request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return limits, wrapper.Wrap(err, "failed decoding torrentsUploadLimit response")
	}
	return limits, nil
}

//SetTorrentUploadLimit for a list of torrents
func (client *Client) SetTorrentUploadLimit(hashes []string, limit string) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"limit":  limit,
	}
	resp, err := client.post("api/v2/torrents/setUploadLimit", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetUploadLimit request")
	}

	return resp.StatusCode == 200, nil
}

//SetTorrentLocation for a list of torrents
func (client *Client) SetTorrentLocation(hashes []string, location string) (bool, error) {
	opts := map[string]string{
		"hashes":   delimit(hashes, "|"),
		"location": location,
	}
	resp, err := client.post("api/v2/torrents/setLocation", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetLocation request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//SetTorrentName for a torrent
func (client *Client) SetTorrentName(hash string, name string) (bool, error) {
	opts := map[string]string{
		"hash": hash,
		"name": name,
	}
	resp, err := client.post("api/v2/torrents/rename", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsRename request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//SetTorrentCategory for a list of torrents
func (client *Client) SetTorrentCategory(hashes []string, category string) (bool, error) {
	opts := map[string]string{
		"hashes":   delimit(hashes, "|"),
		"category": category,
	}
	resp, err := client.post("api/v2/torrents/setCategory", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetCategory request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//GetCategories used by client
func (client *Client) GetCategories() (categories Categories, err error) {
	resp, err := client.get("api/v2/torrents/categories", nil)
	if err != nil {
		return categories, wrapper.Wrap(err, "failed torrentsCategories request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return categories, wrapper.Wrap(err, "failed decoding torrentsCategories response")
	}

	return categories, nil
}

//CreateCategory for use by client
func (client *Client) CreateCategory(category string, savePath string) (bool, error) {
	opts := map[string]string{
		"category": category,
		"savePath": savePath,
	}
	resp, err := client.post("api/v2/torrents/createCategory", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsCreateCategory request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//UpdateCategory used by client
func (client *Client) UpdateCategory(category string, savePath string) (bool, error) {
	opts := map[string]string{
		"category": category,
		"savePath": savePath,
	}
	resp, err := client.post("api/v2/torrents/editCategory", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsEditCategory request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//DeleteCategories used by client
func (client *Client) DeleteCategories(categories []string) (bool, error) {
	opts := map[string]string{"categories": delimit(categories, "\n")}
	resp, err := client.post("api/v2/torrents/removeCategories", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsRemoveCategories request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//AddTorrentTags to a list of torrents
func (client *Client) AddTorrentTags(hashes []string, tags []string) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"tags":   delimit(tags, ","),
	}
	resp, err := client.post("api/v2/torrents/addTags", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsAddTags request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//RemoveTorrentTags from a list of torrents (empty list removes all tags)
func (client *Client) RemoveTorrentTags(hashes []string, tags []string) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"tags":   delimit(tags, ","),
	}
	resp, err := client.post("api/v2/torrents/removeTags", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsRemoveTags request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//GetTorrentTags from a list of torrents (empty list removes all tags)
func (client *Client) GetTorrentTags() (tags []string, err error) {
	resp, err := client.get("api/v2/torrents/tags", nil)
	if err != nil {
		return nil, wrapper.Wrap(err, "failed torrentsTags request")
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return tags, wrapper.Wrap(err, "failed decoding torrentsTags response")
	}

	return tags, nil
}

//CreateTags for use by client
func (client *Client) CreateTags(tags []string) (bool, error) {
	opts := map[string]string{"tags": delimit(tags, ",")}
	resp, err := client.post("api/v2/torrents/createTags", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsCreateTags request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//DeleteTags used by client
func (client *Client) DeleteTags(tags []string) (bool, error) {
	opts := map[string]string{"tags": delimit(tags, ",")}
	resp, err := client.post("api/v2/torrents/deleteTags", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsDeleteTags request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//SetAutoManagement for a list of torrents
func (client *Client) SetAutoManagement(hashes []string, enable bool) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"enable": strconv.FormatBool(enable),
	}
	resp, err := client.post("api/v2/torrents/setAutoManagement", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetAutoManagement request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//ToggleSequentialDownload for a list of torrents
func (client *Client) ToggleSequentialDownload(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/toggleSequentialDownload", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsToggleSequentialDownload request")
	}
	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//ToggleFirstLastPiecePriority for a list of torrents
func (client *Client) ToggleFirstLastPiecePriority(hashes []string) (bool, error) {
	opts := map[string]string{"hashes": delimit(hashes, "|")}
	resp, err := client.get("api/v2/torrents/toggleFirstLastPiecePrio", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsToggleFirstLastPiecePrio request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//SetForceStart for a list of torrents
func (client *Client) SetForceStart(hashes []string, value bool) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"value":  strconv.FormatBool(value),
	}
	resp, err := client.post("api/v2/torrents/setForceStart", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetForceStart request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}

//SetSuperSeeding for a list of torrents
func (client *Client) SetSuperSeeding(hashes []string, value bool) (bool, error) {
	opts := map[string]string{
		"hashes": delimit(hashes, "|"),
		"value":  strconv.FormatBool(value),
	}
	resp, err := client.post("api/v2/torrents/setSuperSeeding", opts)
	if err != nil {
		return false, wrapper.Wrap(err, "failed torrentsSetSuperSeeding request")
	}

	return resp.StatusCode == 200, nil //TODO: look into other statuses
}
