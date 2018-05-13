package radiko

const (
	// API to get programs informations (latest version is v3)
	RADIKO_API_URL string = `https://radiko.jp/v3`

	//RADIKO_API_URL   string = `https://radiko.jp/v2/api`
	// URL to get .m3u8
	RADIKO_PLAYLIST_URL string = `https://radiko.jp/v2/api/ts/playlist.m3u8`

	// URL to get auth token
	RADIKO_AUTH1_URL string = `https://radiko.jp/v2/api/auth1`

	// URL to get activate auth token
	RADIKO_AUTH2_URL string = `https://radiko.jp/v2/api/auth2`

	// authkey necessary to get partial key if using "login2()"
	RADIKO_AUTHKEY string = `bcd151073c03b352e1ef2fd66c32209da9ca0afa`
	// Header values used to connect auth1 and auth2.
	RADIKO_AUTH_HEADER_APP         string = `pc_html5`
	RADIKO_AUTH_HEADER_APP_VERSION string = `0.0.1`
	RADIKO_AUTH_HEADER_USER        string = `dummy_user`
	RADIKO_AUTH_HEADER_DEVICE      string = `pc`

	// Index used to extract partial key from swf.
	// It is used in the old auth process (i.e. auth.go. It is not used in authnew.go)
	// This number is sometimes changed by Radiko
	RADIKO_SWF_EXTRACT_INDEX string = `12`
	// URL to get swf, which is necessary to get partial key if using "login()"
	RADIKO_SWF_URL string = `https://radiko.jp/apps/js/flash/myplayer-release.swf`
	// Header values used to connect auth1 and auth2 with the old method.
	RADIKO_AUTH_HEADER_APP_OLD         string = `pc_ts`
	RADIKO_AUTH_HEADER_APP_VERSION_OLD string = `4.0.0`
	RADIKO_AUTH_HEADER_USER_OLD        string = `test-stream`
	RADIKO_AUTH_HEADER_DEVICE_OLD      string = `pc`
)
