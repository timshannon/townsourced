// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"git.townsourced.com/townsourced/townsourced/data/private"
	"git.townsourced.com/townsourced/townsourced/fail"
)

const (
	facebookAppID    = private.FacebookAppID
	facebookDevAppID = private.FacebookDevAppID

	facebookProdClientSecret = private.FacebookProdClientSecret
	facebookDevClientSecret  = private.FacebookDevClientSecret
)

type facebookResponse struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	ExpiresIn   int            `json:"expires_in"`
	Error       *facebookError `json:"error"`
	Data        *facebookData  `json:"data"`
}

type facebookData struct {
	AppID       string         `json:"app_id"`
	Application string         `json:"application"`
	ExpiresAt   int            `json:"expires_at"`
	Valid       bool           `json:"is_valid"`
	IssuedAt    int            `json:"issued_at"`
	Scopes      []string       `json:"scopes"`
	UserID      string         `json:"user_id"`
	Email       string         `json:"email"`
	Name        string         `json:"name"`
	Error       *facebookError `json:"error"`
}

type facebookError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}

func (fe *facebookError) Error() string {
	return fmt.Sprintf("%s Type: %s Code: %d", fe.Message, fe.Type, fe.Code)
}

// facebookSesson is the user, and access tokens needed
// to get data from the facebook api
type facebookSession struct {
	userID    string
	userToken string
	appToken  string
}

// FacebookAppID is the facebook appID for townsourced
func FacebookAppID() string {
	if devMode {
		return facebookDevAppID
	}
	return facebookAppID
}

func facebookClientSecret() string {
	if devMode {
		return facebookDevClientSecret
	}
	return facebookProdClientSecret
}

// FacebookUser gets a user (if possible)
// from the passed in facebook code
func FacebookUser(redirectURI, code string) (*User, error) {
	fbSes, err := facebookGetSession(redirectURI, code)
	if err != nil {
		return nil, err
	}

	//Lookup user
	usr, err := userGetFacebook(fbSes.userID)
	if err == ErrUserNotFound {
		data, err := fbSes.userData()
		if err != nil {
			return nil, err
		}
		usrFail := fail.NewFromErr(ErrUserNeedUsername, map[string]string{
			"username":  urlify(data.Name).make(),
			"email":     data.Email,
			"userToken": fbSes.userToken,
			"appToken":  fbSes.appToken,
			"userID":    fbSes.userID,
		})
		return nil, usrFail
	}

	if err != nil {
		return nil, err
	}

	return usr, nil
}

// FacebookNewUser creates a new facebook user with the passed in username and email,
// using the passed in temporary tokens
func FacebookNewUser(username, email, userID, userToken, appToken string) (*User, error) {
	fbSes := &facebookSession{
		userToken: userToken,
		appToken:  appToken,
		userID:    userID,
	}

	return facebookNewUser(username, email, fbSes)
}

func facebookNewUser(username, email string, fbSes *facebookSession) (*User, error) {
	//Get email address and name
	data, err := fbSes.userData()
	if err != nil {
		return nil, err
	}

	if email == "" {
		email = data.Email
	}

	u, err := userNewFacebook(username, email, data.Name, fbSes.userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func facebookGetSession(redirectURI, code string) (*facebookSession, error) {
	userDone := make(chan bool)
	appDone := make(chan bool)

	var userErr error
	var appErr error

	fbSes := &facebookSession{}

	go func() {
		userValues := url.Values{}
		userValues.Add("redirect_uri", redirectURI)
		userValues.Add("code", code)

		fbSes.userToken, userErr = facebookGetAccessToken(userValues)
		userDone <- true
	}()

	go func() {
		appValues := url.Values{}
		appValues.Add("grant_type", "client_credentials")
		fbSes.appToken, appErr = facebookGetAccessToken(appValues)
		appDone <- true
	}()

	<-userDone
	<-appDone

	if userErr != nil {
		return nil, fmt.Errorf("Error getting facebook user token: %s", userErr)
	}
	if appErr != nil {
		return nil, fmt.Errorf("Error getting facebook app token: %s", appErr)
	}

	err := fbSes.validate()
	if err != nil {
		return nil, err
	}
	return fbSes, nil
}

func facebookGetAccessToken(values url.Values) (token string, err error) {
	values.Add("client_id", FacebookAppID())
	values.Add("client_secret", facebookClientSecret())

	uri := url.URL{
		Scheme:   "https",
		Host:     "graph.facebook.com",
		Path:     "v2.3/oauth/access_token",
		RawQuery: values.Encode(),
	}

	res, err := httpClient.Get(uri.String())
	if err != nil {
		return "", err
	}
	val, err := ioutil.ReadAll(res.Body)
	defer func() {
		if cerr := res.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if err != nil {
		return "", err
	}

	fbResponse := &facebookResponse{}
	err = json.Unmarshal(val, fbResponse)
	if err != nil {
		return "", err
	}

	if fbResponse.Error != nil {
		return "", fbResponse.Error
	}

	return fbResponse.AccessToken, nil
}

func (fbSes *facebookSession) values() url.Values {
	values := url.Values{}
	values.Add("input_token", fbSes.userToken)
	values.Add("access_token", fbSes.appToken)
	return values
}

func (fbSes *facebookSession) validate() (err error) {
	uri := url.URL{
		Scheme:   "https",
		Host:     "graph.facebook.com",
		Path:     "debug_token",
		RawQuery: fbSes.values().Encode(),
	}

	res, err := httpClient.Get(uri.String())
	if err != nil {
		return err
	}
	val, err := ioutil.ReadAll(res.Body)

	defer func() {
		if cerr := res.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if err != nil {
		return err
	}

	fbResponse := &facebookResponse{}
	err = json.Unmarshal(val, fbResponse)
	if err != nil {
		return err
	}

	if fbResponse.Error != nil {
		return fbResponse.Error
	}

	if !fbResponse.Data.Valid {
		return fmt.Errorf("Facebook returned an invalid user access token. Token %s", fbSes.userToken)
	}

	if fbResponse.Data.AppID != FacebookAppID() {
		return fmt.Errorf("Invalid appID returned from facebook.  Wanted %s, got %s", FacebookAppID(), fbResponse.Data.AppID)
	}

	fbSes.userID = fbResponse.Data.UserID
	return nil
}

func (fbSes *facebookSession) userData() (fbData *facebookData, err error) {
	uri := url.URL{
		Scheme:   "https",
		Host:     "graph.facebook.com",
		Path:     fbSes.userID,
		RawQuery: fbSes.values().Encode(),
	}

	res, err := httpClient.Get(uri.String())
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

	fbData = &facebookData{}
	err = json.Unmarshal(val, fbData)
	if err != nil {
		return nil, err
	}

	if fbData.Error != nil {
		return nil, fbData.Error
	}

	return fbData, nil
}
