package radiko

import (
	"encoding/base64"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

func getAuthToken2(client *http.Client) (string, int, int, error) {
	/* Get authtoken from auth1*/
	u, err := url.Parse(RADIKO_AUTH1_URL)
	if err != nil {
		return "", 0, 0, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", 0, 0, err
	}
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("X-Radiko-App", RADIKO_AUTH_HEADER_APP)
	req.Header.Set("X-Radiko-App-Version", RADIKO_AUTH_HEADER_APP_VERSION)
	req.Header.Set("X-Radiko-User", RADIKO_AUTH_HEADER_USER)
	req.Header.Set("X-Radiko-Device", RADIKO_AUTH_HEADER_DEVICE)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return "", 0, 0, err
	}
	token := res.Header.Get(`X-Radiko-Authtoken`)
	length := res.Header.Get(`X-Radiko-KeyLength`)
	l, err := strconv.Atoi(length)
	if err != nil {
		return token, 0, 0, err
	}
	offset := res.Header.Get(`X-Radiko-KeyOffset`)
	o, err := strconv.Atoi(offset)
	if err != nil {
		return token, l, 0, err
	}
	return token, l, o, nil
}

func getPartialKey2(client *http.Client, offset, length int) string {
	// comparable to following command in unix
	// `echo RADIKO_AUTHKEY | dd bs=1 skip=offset count=length | base64`
	partialkey := base64.StdEncoding.EncodeToString([]byte(RADIKO_AUTHKEY[offset : offset+length]))
	return partialkey
}

func activateAuthToken2(client *http.Client, token, partialkey string) error {
	/* Acitivate authtoken using partial key from auth2*/
	u, err := url.Parse(RADIKO_AUTH2_URL)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("pragma", `no-cache`)
	req.Header.Set("X-Radiko-App", RADIKO_AUTH_HEADER_APP)
	req.Header.Set("X-Radiko-App-Version", RADIKO_AUTH_HEADER_APP_VERSION)
	req.Header.Set("X-Radiko-User", RADIKO_AUTH_HEADER_USER)
	req.Header.Set("X-Radiko-Device", RADIKO_AUTH_HEADER_DEVICE)
	req.Header.Set("X-Radiko-AuthToken", token)
	req.Header.Set("X-Radiko-Partialkey", partialkey)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func login2() (*http.Client, string, error) {
	/* Save cookies */
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{Jar: jar}

	/* Get authtoken from auth1 response header */
	token, length, offset, err := getAuthToken2(client)
	if err != nil {
		return client, "", err
	}

	/* Get partialkey from myplayer.swf */
	partialkey := getPartialKey2(client, offset, length)

	/* Activate authkey by sending token and partial key to auth2 */
	if err := activateAuthToken2(client, token, partialkey); err != nil {
		return client, token, err
	}

	return client, token, nil
}
