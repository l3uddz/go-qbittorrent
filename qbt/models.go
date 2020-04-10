package qbt

//BasicTorrent holds a basic torrent object from qbittorrent
type BasicTorrent struct {
	Category               string  `json:"category"`
	CompletionOn           int64   `json:"completion_on"`
	Dlspeed                int64   `json:"dlspeed"`
	Eta                    int64   `json:"eta"`
	ForceStart             bool    `json:"force_start"`
	Hash                   string  `json:"hash"`
	Name                   string  `json:"name"`
	NumComplete            int64   `json:"num_complete"`
	NumIncomplete          int64   `json:"num_incomplete"`
	NumLeechs              int64   `json:"num_leechs"`
	NumSeeds               int64   `json:"num_seeds"`
	Priority               int64   `json:"priority"`
	Progress               float32 `json:"progress"`
	Ratio                  float32 `json:"ratio"`
	SavePath               string  `json:"save_path"`
	SeqDl                  bool    `json:"seq_dl"`
	Size                   int64   `json:"size"`
	Downloaded             int64   `json:"downloaded"`
	Uploaded               int64   `json:"uploaded"`
	AmountLeft             int64   `json:"amount_left"`
	State                  string  `json:"state"`
	SuperSeeding           bool    `json:"super_seeding"`
	Upspeed                int64   `json:"upspeed"`
	FirstLastPiecePriority bool    `json:"f_l_piece_prio"`
	Tracker                string  `json:"tracker"`
}

//Torrent holds a torrent object from qbittorrent
//with more information than BasicTorrent
type Torrent struct {
	AdditionDate       int64   `json:"addition_date"`
	Comment            string  `json:"comment"`
	CompletionDate     int64   `json:"completion_date"`
	CreatedBy          string  `json:"created_by"`
	CreationDate       int64   `json:"creation_date"`
	DlLimit            int64   `json:"dl_limit"`
	DlSpeed            int64   `json:"dl_speed"`
	DlSpeedAvg         int64   `json:"dl_speed_avg"`
	Eta                int64   `json:"eta"`
	LastSeen           int64   `json:"last_seen"`
	NbConnections      int64   `json:"nb_connections"`
	NbConnectionsLimit int64   `json:"nb_connections_limit"`
	Peers              int64   `json:"peers"`
	PeersTotal         int64   `json:"peers_total"`
	PieceSize          int64   `json:"piece_size"`
	PiecesHave         int64   `json:"pieces_have"`
	PiecesNum          int64   `json:"pieces_num"`
	Reannounce         int64   `json:"reannounce"`
	SavePath           string  `json:"save_path"`
	SeedingTime        int64   `json:"seeding_time"`
	Seeds              int64   `json:"seeds"`
	SeedsTotal         int64   `json:"seeds_total"`
	ShareRatio         float32 `json:"share_ratio"`
	TimeElapsed        int64   `json:"time_elapsed"`
	TotalDl            int64   `json:"total_downloaded"`
	TotalDlSession     int64   `json:"total_downloaded_session"`
	TotalSize          int64   `json:"total_size"`
	TotalUl            int64   `json:"total_uploaded"`
	TotalUlSession     int64   `json:"total_uploaded_session"`
	TotalWasted        int64   `json:"total_wasted"`
	UpLimit            int64   `json:"up_limit"`
	UpSpeed            int64   `json:"up_speed"`
	UpSpeedAvg         int64   `json:"up_speed_avg"`
}

//Tracker holds a tracker object from qbittorrent
type Tracker struct {
	Msg           string      `json:"msg"`
	NumPeers      int64       `json:"num_peers"`
	NumSeeds      int64       `json:"num_seeds"`
	NumLeeches    int64       `json:"num_leeches"`
	NumDownloaded int64       `json:"num_downloaded"`
	Tier          interface{} `json:"tier"`
	Status        int64       `json:"status"`
	URL           string      `json:"url"`
}

//WebSeed holds a webseed object from qbittorrent
type WebSeed struct {
	URL string `json:"url"`
}

//TorrentFile holds a torrent file object from qbittorrent
type TorrentFile struct {
	IsSeed       bool    `json:"is_seed"`
	Name         string  `json:"name"`
	Availability float32 `json:"availability"`
	Priority     int64   `json:"priority"`
	Progress     float32 `json:"progress"`
	Size         int64   `json:"size"`
	PieceRange   []int64 `json:"piece_range"`
}

// Sync holds server state that we are interested in
type Sync struct {
	ServerState struct {
		ConnectionStatus  string `json:"connection_status"`
		DhtNodes          int64  `json:"dht_nodes"`
		DlInfoData        int64  `json:"dl_info_data"`
		DlInfoSpeed       int64  `json:"dl_info_speed"`
		DlRateLimit       int64  `json:"dl_rate_limit"`
		Queueing          bool   `json:"queueing"`
		RefreshInterval   int64  `json:"refresh_interval"`
		UpInfoData        int64  `json:"up_info_data"`
		UpInfoSpeed       int64  `json:"up_info_speed"`
		UpRateLimit       int64  `json:"up_rate_limit"`
		UseAltSpeedLimits bool   `json:"use_alt_speed_limits"`
		FreeSpaceOnDisk   int64  `json:"free_space_on_disk"`
	} `json:"server_state"`
}

type BuildInfo struct {
	QTVersion         string `json:"qt"`
	LibtorrentVersion string `json:"libtorrent"`
	BoostVersion      string `json:"boost"`
	OpenSSLVersion    string `json:"openssl"`
	AppBitness        string `json:"bitness"`
}

type Preferences struct {
	Locale                             string                 `json:"locale"`
	CreateSubfolderEnabled             bool                   `json:"create_subfolder_enabled"`
	StartPausedEnabled                 bool                   `json:"start_paused_enabled"`
	AutoDeleteMode                     int64                  `json:"auto_delete_mode"`
	PreallocateAll                     bool                   `json:"preallocate_all"`
	IncompleteFilesExt                 bool                   `json:"incomplete_files_ext"`
	AutoTMMEnabled                     bool                   `json:"auto_tmm_enabled"`
	TorrentChangedTMMEnabled           bool                   `json:"torrent_changed_tmm_enabled"`
	SavePathChangedTMMEnabled          bool                   `json:"save_path_changed_tmm_enabled"`
	CategoryChangedTMMEnabled          bool                   `json:"category_changed_tmm_enabled"`
	SavePath                           string                 `json:"save_path"`
	TempPathEnabled                    bool                   `json:"temp_path_enabled"`
	TempPath                           string                 `json:"temp_path"`
	ScanDirs                           map[string]interface{} `json:"scan_dirs"`
	ExportDir                          string                 `json:"export_dir"`
	ExportDirFin                       string                 `json:"export_dir_fin"`
	MailNotificationEnabled            string                 `json:"mail_notification_enabled"`
	MailNotificationSender             string                 `json:"mail_notification_sender"`
	MailNotificationEmail              string                 `json:"mail_notification_email"`
	MailNotificationSMPTP              string                 `json:"mail_notification_smtp"`
	MailNotificationSSLEnabled         bool                   `json:"mail_notification_ssl_enabled"`
	MailNotificationAuthEnabled        bool                   `json:"mail_notification_auth_enabled"`
	MailNotificationUsername           string                 `json:"mail_notification_username"`
	MailNotificationPassword           string                 `json:"mail_notification_password"`
	AutorunEnabled                     bool                   `json:"autorun_enabled"`
	AutorunProgram                     string                 `json:"autorun_program"`
	QueueingEnabled                    bool                   `json:"queueing_enabled"`
	MaxActiveDls                       int64                  `json:"max_active_downloads"`
	MaxActiveTorrents                  int64                  `json:"max_active_torrents"`
	MaxActiveUls                       int64                  `json:"max_active_uploads"`
	DontCountSlowTorrents              bool                   `json:"dont_count_slow_torrents"`
	SlowTorrentDlRateThreshold         int64                  `json:"slow_torrent_dl_rate_threshold"`
	SlowTorrentUlRateThreshold         int64                  `json:"slow_torrent_ul_rate_threshold"`
	SlowTorrentInactiveTimer           int64                  `json:"slow_torrent_inactive_timer"`
	MaxRatioEnabled                    bool                   `json:"max_ratio_enabled"`
	MaxRatio                           float64                `json:"max_ratio"`
	MaxRatioAct                        bool                   `json:"max_ratio_act"`
	ListenPort                         int64                  `json:"listen_port"`
	UPNP                               bool                   `json:"upnp"`
	RandomPort                         bool                   `json:"random_port"`
	DlLimit                            int64                  `json:"dl_limit"`
	UlLimit                            int64                  `json:"up_limit"`
	MaxConnections                     int64                  `json:"max_connec"`
	MaxConnectionsPerTorrent           int64                  `json:"max_connec_per_torrent"`
	MaxUls                             int64                  `json:"max_uploads"`
	MaxUlsPerTorrent                   int64                  `json:"max_uploads_per_torrent"`
	UTPEnabled                         bool                   `json:"enable_utp"`
	LimitUTPRate                       bool                   `json:"limit_utp_rate"`
	LimitTCPOverhead                   bool                   `json:"limit_tcp_overhead"`
	LimitLANPeers                      bool                   `json:"limit_lan_peers"`
	AltDlLimit                         int64                  `json:"alt_dl_limit"`
	AltUlLimit                         int64                  `json:"alt_up_limit"`
	SchedulerEnabled                   bool                   `json:"scheduler_enabled"`
	ScheduleFromHour                   int64                  `json:"schedule_from_hour"`
	ScheduleFromMin                    int64                  `json:"schedule_from_min"`
	ScheduleToHour                     int64                  `json:"schedule_to_hour"`
	ScheduleToMin                      int64                  `json:"schedule_to_min"`
	SchedulerDays                      int64                  `json:"scheduler_days"`
	DHTEnabled                         bool                   `json:"dht"`
	DHTSameAsBT                        bool                   `json:"dhtSameAsBT"`
	DHTPort                            int64                  `json:"dht_port"`
	PexEnabled                         bool                   `json:"pex"`
	LSDEnabled                         bool                   `json:"lsd"`
	Encryption                         int64                  `json:"encryption"`
	AnonymousMode                      bool                   `json:"anonymous_mode"`
	ProxyType                          int64                  `json:"proxy_type"`
	ProxyIP                            string                 `json:"proxy_ip"`
	ProxyPort                          int64                  `json:"proxy_port"`
	ProxyPeerConnections               bool                   `json:"proxy_peer_connections"`
	ForceProxy                         bool                   `json:"force_proxy"`
	ProxyAuthEnabled                   bool                   `json:"proxy_auth_enabled"`
	ProxyUsername                      string                 `json:"proxy_username"`
	ProxyPassword                      string                 `json:"proxy_password"`
	IPFilterEnabled                    bool                   `json:"ip_filter_enabled"`
	IPFilterPath                       string                 `json:"ip_filter_path"`
	IPFilterTrackers                   string                 `json:"ip_filter_trackers"`
	WebUIDomainList                    string                 `json:"web_ui_domain_list"`
	WebUIAddress                       string                 `json:"web_ui_address"`
	WebUIPort                          int64                  `json:"web_ui_port"`
	WebUIUPNPEnabled                   bool                   `json:"web_ui_upnp"`
	WebUIUsername                      string                 `json:"web_ui_username"`
	WebUIPassword                      string                 `json:"web_ui_password"`
	WebUICSRFProtectionEnabled         bool                   `json:"web_ui_csrf_protection_enabled"`
	WebUIClickjackingProtectionEnabled bool                   `json:"web_ui_clickjacking_protection_enabled"`
	BypassLocalAuth                    bool                   `json:"bypass_local_auth"`
	BypassAuthSubnetWhitelistEnabled   bool                   `json:"bypass_auth_subnet_whitelist_enabled"`
	BypassAuthSubnetWhitelist          string                 `json:"bypass_auth_subnet_whitelist"`
	AltWebUIEnabled                    bool                   `json:"alternative_webui_enabled"`
	AltWebUIPath                       string                 `json:"alternative_webui_path"`
	UseHTTPS                           bool                   `json:"use_https"`
	SSLKey                             string                 `json:"ssl_key"`
	SSLCert                            string                 `json:"ssl_cert"`
	DynDNSEnabled                      bool                   `json:"dyndns_enabled"`
	DynDNSService                      int64                  `json:"dyndns_service"`
	DynDNSUsername                     string                 `json:"dyndns_username"`
	DynDNSPassword                     string                 `json:"dyndns_password"`
	DynDNSDomain                       string                 `json:"dyndns_domain"`
	RSSRefreshInterval                 int64                  `json:"rss_refresh_interval"`
	RSSMaxArtPerFeed                   int64                  `json:"rss_max_articles_per_feed"`
	RSSProcessingEnabled               bool                   `json:"rss_processing_enabled"`
	RSSAutoDlEnabled                   bool                   `json:"rss_auto_downloading_enabled"`
}

//Log
type Log struct {
	ID        int64  `json:"id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      int64  `json:"type"`
}

//PeerLog
type PeerLog struct {
	ID        int64  `json:"id"`
	IP        string `json:"ip"`
	Blocked   bool   `json:"blocked"`
	Timestamp int64  `json:"timestamp"`
	reason    string `json:"reason"`
}

//Info
type Info struct {
	ConnectionStatus  string `json:"connection_status"`
	DHTNodes          int64  `json:"dht_nodes"`
	DlInfoData        int64  `json:"dl_info_data"`
	DlInfoSpeed       int64  `json:"dl_info_speed"`
	DlRateLimit       int64  `json:"dl_rate_limit"`
	UlInfoData        int64  `json:"up_info_data"`
	UlInfoSpeed       int64  `json:"up_info_speed"`
	UlRateLimit       int64  `json:"up_rate_limit"`
	Queueing          bool   `json:"queueing"`
	UseAltSpeedLimits bool   `json:"use_alt_speed_limits"`
	RefreshInterval   int64  `json:"refresh_interval"`
}

type TorrentInfo struct {
	Filter   string // all, downloading, completed, paused, active, inactive => optional
	Category string // => optional
	Sort     string // => optional
	Reverse  bool   // => optional
	Limit    int64  // => optional (no negatives)
	Offset   int64  // => optional (negatives allowed)
	Hashes   string // separated by | => optional
}

//Category of torrent
type Category struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}

//Categories mapping
type Categories struct {
	Category map[string]Category
}

//LoginOptions contains all options for /login endpoint
type LoginOptions struct {
	Username string
	Password string
}

//AddTrackersOptions contains all options for /addTrackers endpoint
type AddTrackersOptions struct {
	Hash     string
	Trackers []string
}

//EditTrackerOptions contains all options for /editTracker endpoint
type EditTrackerOptions struct {
	Hash    string
	OrigURL string
	NewURL  string
}

//RemoveTrackersOptions contains all options for /removeTrackers endpoint
type RemoveTrackersOptions struct {
	Hash     string
	Trackers []string
}
