// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"git.townsourced.com/townsourced/oauth"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/data/private"
	"git.townsourced.com/townsourced/townsourced/fail"
)

func init() {
	twitterConsumer = oauth.NewCustomHttpClientConsumer(twitterAPIKey, twitterAPISecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authenticate",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		}, httpClient)
}

const (
	twitterAPIKey    = private.TwitterAPIKey
	twitterAPISecret = private.TwitterAPISecret
)

var twitterConsumer *oauth.Consumer

type twitterSession struct {
	*ThirdPartyState
	OauthToken  *oauth.RequestToken
	AccessToken *oauth.AccessToken
}

type twitterData struct {
	Name       string `json:"name"`
	ID         string `json:"id_str"`
	ScreenName string `json:"screen_name"`
	Email      string `json:"email"`
}

// TwitterGetLoginURL requests a new oauth token for use with sign in with twitter
func TwitterGetLoginURL(redirectURL, stateToken string) (string, error) {
	//get user state
	state, err := ThirdPartyStateGet(stateToken)
	if err != nil {
		return "", err
	}

	//Build state token into redirect url
	uri, err := url.Parse(redirectURL)
	if err != nil {
		return "", fmt.Errorf("Error parsing twitter redirectURL: %v", err)
	}

	values := url.Values{}
	values.Add("state", stateToken)
	uri.RawQuery = values.Encode()

	oauthToken, loginURL, err := twitterConsumer.GetRequestTokenAndUrl(uri.String())
	if err != nil {
		return "", err
	}

	//store oauthToken with user state for use later
	err = data.TempTokenSet(&twitterSession{
		ThirdPartyState: state,
		OauthToken:      oauthToken,
	}, stateToken, 15*time.Minute)

	if err != nil {
		return "", err
	}

	return loginURL, nil
}

// TwitterGetUser returns a townsourced user (if one exists) associated with
// the passed in oauth verification code.  If a user doesn't yet exist
// it'll pass back username and email hints for creating a new user
func TwitterGetUser(stateToken, verificationCode string) (*User, error) {
	session := &twitterSession{}

	err := data.TempTokenGet(session, stateToken)
	if err != nil {
		return nil, err
	}

	accessToken, err := twitterConsumer.AuthorizeToken(session.OauthToken, verificationCode)
	if err != nil {
		return nil, err
	}

	session.OauthToken = nil
	session.AccessToken = accessToken

	userData, err := session.userData()
	if err != nil {
		return nil, err
	}

	//Lookup user
	usr, err := userGetTwitter(userData.ID)
	if err == ErrUserNotFound {
		//save session for later
		err = data.TempTokenSet(session, session.Token, 15*time.Minute)

		if err != nil {
			return nil, err
		}

		usrFail := fail.NewFromErr(ErrUserNeedUsername, map[string]string{
			"username": urlify(userData.ScreenName).make(),
			"email":    userData.Email,
			"token":    stateToken,
		})
		return nil, usrFail
	}

	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (t *twitterSession) userData() (userData *twitterData, err error) {
	res, err := twitterConsumer.Get("https://api.twitter.com/1.1/account/verify_credentials.json", map[string]string{
		"include_email":    "true",
		"skip_status":      "true",
		"include_entities": "false",
	}, t.AccessToken)
	if err != nil {
		return nil, err
	}

	val, err := ioutil.ReadAll(res.Body)
	defer func() {
		if cerr := res.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if err != nil {
		return nil, err
	}

	userData = &twitterData{}
	err = json.Unmarshal(val, userData)
	if err != nil {
		return nil, err
	}
	return userData, nil
}

// TwitterNewUser creates a new twitter user with the passed in username and email,
// using the passed in temporary token
func TwitterNewUser(username, email, stateToken string) (*User, error) {
	session := &twitterSession{}

	err := data.TempTokenGet(session, stateToken)
	if err != nil {
		return nil, err
	}

	userData, err := session.userData()
	if err != nil {
		return nil, err
	}

	if email == "" {
		email = userData.Email
	}

	u, err := userNewTwitter(username, email, userData.Name, userData.ID)
	if err != nil {
		return nil, err
	}

	return u, nil
}
