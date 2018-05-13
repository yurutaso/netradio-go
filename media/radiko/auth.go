package radiko

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
)

/*
NOTE: It is recommended to use authnew.go (i.e. function login2() in authnew.go),
      since that is newer and follows Radiko's default authorization process.
*/

func getAuthToken(client *http.Client) (token string, length string, offset string, err error) {
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
	req.Header.Set("X-Radiko-App", RADIKO_AUTH_HEADER_APP_OLD)
	req.Header.Set("X-Radiko-App-Version", RADIKO_AUTH_HEADER_APP_VERSION_OLD)
	req.Header.Set("X-Radiko-User", RADIKO_AUTH_HEADER_USER_OLD)
	req.Header.Set("X-Radiko-Device", RADIKO_AUTH_HEADER_DEVICE_OLD)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return "", "", "", err
	}
	token = res.Header.Get(`X-Radiko-Authtoken`)
	length = res.Header.Get(`X-Radiko-KeyLength`)
	offset = res.Header.Get(`X-Radiko-KeyOffset`)
	return token, length, offset, nil
}

func getPartialKey(client *http.Client, offset, length string) (string, error) {
	/* Download palyer.swf */
	tmpswf, err := ioutil.TempFile("", "tmp_swf")
	if err != nil {
		return "", err
	}
	defer tmpswf.Close()
	req, err := http.NewRequest("GET", RADIKO_SWF_URL, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	_, err = io.Copy(tmpswf, res.Body)
	if err != nil {
		return "", err
	}

	/* Get partial key from swf*/
	tmpext, err := ioutil.TempFile("", "tmp_swf_extracted")
	if err != nil {
		return "", err
	}
	defer tmpext.Close()
	swfextract := exec.Command(`swfextract`, `-b`, RADIKO_SWF_EXTRACT_INDEX, tmpswf.Name(), `-o`, tmpext.Name())
	err = swfextract.Run()
	if err != nil {
		return "", err
	}
	dd := exec.Command(`dd`, `if=`+tmpext.Name(), `bs=1`, `skip=`+offset, `count=`+length)
	base64 := exec.Command(`base64`)
	stdout, err := dd.StdoutPipe()
	if err != nil {
		return "", err
	}
	base64.Stdin = stdout
	go func() {
		defer stdout.Close()
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

func activateAuthToken(client *http.Client, token, partialkey string) error {
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
	req.Header.Set("X-Radiko-App", RADIKO_AUTH_HEADER_APP_OLD)
	req.Header.Set("X-Radiko-App-Version", RADIKO_AUTH_HEADER_APP_VERSION_OLD)
	req.Header.Set("X-Radiko-User", RADIKO_AUTH_HEADER_USER_OLD)
	req.Header.Set("X-Radiko-Device", RADIKO_AUTH_HEADER_DEVICE_OLD)
	req.Header.Set("X-Radiko-AuthToken", token)
	req.Header.Set("X-Radiko-Partialkey", partialkey)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func login() (client *http.Client, token string, err error) {
	/* Save cookies */
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "", err
	}
	client = &http.Client{Jar: jar}

	/* Get authtoken from auth1 response header */
	token, length, offset, err := getAuthToken(client)
	if err != nil {
		return client, "", err
	}

	/* Get partialkey from myplayer.swf */
	partialkey, err := getPartialKey(client, offset, length)
	if err != nil {
		return client, token, err
	}

	/* Activate authkey by sending token and partial key to auth2 */
	err = activateAuthToken(client, token, partialkey)
	if err != nil {
		return client, token, err
	}

	return client, token, nil
}
