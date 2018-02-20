package radiko

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
)

func getAuthToken2(client *http.Client) (string, string, string, error) {
	/* Get authtoken from auth1*/
	u, err := url.Parse(RADIKO_AUTH1_URL)
	if err != nil {
		return "", "", "", err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("X-Radiko-App", RADIKO_AUTH_HEADER_APP)
	req.Header.Set("X-Radiko-App-Version", RADIKO_AUTH_HEADER_APP_VERSION)
	req.Header.Set("X-Radiko-User", RADIKO_AUTH_HEADER_USER)
	req.Header.Set("X-Radiko-Device", RADIKO_AUTH_HEADER_DEVICE)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return "", "", "", err
	}
	token := res.Header.Get(`X-Radiko-Authtoken`)
	length := res.Header.Get(`X-Radiko-KeyLength`)
	offset := res.Header.Get(`X-Radiko-KeyOffset`)
	return token, length, offset, nil
}
func getPartialKey2(client *http.Client, offset, length string) (string, error) {
	echo := exec.Command(`echo`, RADIKO_AUTHKEY)
	stdout1, err := echo.StdoutPipe()
	if err != nil {
		return "", err
	}

	dd := exec.Command(`dd`, `bs=1`, `skip=`+offset, `count=`+length)
	dd.Stdin = stdout1
	stdout2, err := dd.StdoutPipe()
	if err != nil {
		return "", err
	}

	base64 := exec.Command(`base64`)
	base64.Stdin = stdout2
	go func() {
		defer stdout1.Close()
		echo.Start()
	}()
	go func() {
		defer stdout2.Close()
		dd.Start()
	}()
	result, err := base64.CombinedOutput()
	if err != nil {
		return "", err
	}
	partialkey := string(result)
	partialkey = partialkey[0 : len(partialkey)-1]
	return partialkey, nil
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
	partialkey, err := getPartialKey2(client, offset, length)
	if err != nil {
		return client, token, err
	}

	/* Activate authkey by sending token and partial key to auth2 */
	err = activateAuthToken2(client, token, partialkey)
	if err != nil {
		return client, token, err
	}

	return client, token, nil
}
