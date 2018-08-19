package melican

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/LincolnEthan/oauth2"
	//"io/ioutil"
	"github.com/pkg/errors"
)

var (
	host     string // ex. "10.1.51.8"
	domain   string // ex. "https://liveoa.melican.cn"
	hostname string // ex. "liveoa.melican.cn"
)

const (
	accessTokenPath = "/o/token/"
	authorizePath   = "/o/authorize/"
	userPath        = "/api/v1/shangmu/user/"
	logoutPath      = "/api/v1/shangmu/logout/"
)

// Endpoint is melican online-oa endpoint.
var Endpoint oauth2.Endpoint

// User defines oa response of user information.
type User struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
}

// Init initializes oa domain host and endpoints.
func Init(d, h string) {
	domain, host = d, h
	u, err := url.Parse(domain)
	if err != nil {
		fmt.Printf("oa domain parse fail, %s\n", err.Error())
		panic(err)
	}
	hostname = u.Host

	tokenURL := fmt.Sprintf("%s%s", domain, accessTokenPath)
	if host != "" {
		oauth2.WrapRequest(func(req *http.Request) {
			req.Host = hostname
		})
		tokenURL = fmt.Sprintf("http://%s%s", host, accessTokenPath)
	}
	oauth2.RegisterBrokenAuthHeaderProvider(tokenURL)
	Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s%s", domain, authorizePath),
		TokenURL: tokenURL,
	}
}

// GetUser gives you oa login user.
func GetUser(client *http.Client, token *oauth2.Token) (User, error) {

	var req *http.Request
	var err error

	if host != "" {
		getUserURL := fmt.Sprintf("http://%s%s?token=%s", host, userPath, token.AccessToken)
		req, err = http.NewRequest("GET", getUserURL, nil)
		if err != nil {
			err = errors.New(fmt.Sprintf("build request fail, %s\n", err.Error()))
			return User{}, err
		}
		req.Host = hostname
	} else {
		getUserURL := fmt.Sprintf("%s%s?token=%s", domain, userPath, token.AccessToken)
		req, err = http.NewRequest("GET", getUserURL, nil)
		fmt.Println("resp.Body", getUserURL)

		if err != nil {
			err = errors.New(fmt.Sprintf("build request fail, %s\n", err.Error()))
			return User{}, err
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()

	var response struct {
		Error    uint32 `json:"error"`
		Msg      string `json:"msg"`
		ID       string `json:"ID"`
		UserName string `json:"username"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return User{}, err
	}

	if response.Error != 0 {
		err = errors.New(fmt.Sprintf("oauth2 response code is %d, error msg is %s", response.Error, response.Msg))
		return User{}, err
	}

	return User{
		ID:       response.ID,
		UserName: response.UserName,
	}, nil
}

// GetLogoutURL gives you logout URL.
func GetLogoutURL(service string) string {
	fmt.Println(fmt.Sprintf("%s%s?service=%s", domain, logoutPath, service))
	return fmt.Sprintf("%s%s?service=%s", domain, logoutPath, service)
}
