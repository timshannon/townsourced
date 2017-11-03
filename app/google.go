// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/timshannon/townsourced/data/private"
	"github.com/timshannon/townsourced/fail"
)

const (
	googleClientID             = private.GoogleClientID
	googleClientSecret         = private.GoogleClientSecret
	googleDiscoveryDocumentURL = "https://accounts.google.com/.well-known/openid-configuration"
)

type googleDiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

type googleOauthSession struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`

	jwt *googleJWT
}

type googleJWT struct {
	ISS     string `json:"iss"`
	UserID  string `json:"sub"`
	Email   string `json:"email"`
	AppID   string `json:"aud"`
	Expires int64  `json:"exp"`
}

type googlePeople struct {
	DisplayName string `json:"displayName"`

	Error   *googleAPIError `json:"error"`
	session *googleOauthSession
}

type googleAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (ge *googleAPIError) Error() string {
	return fmt.Sprintf("%s  Code: %d", ge.Message, ge.Code)
}

// GoogleOpenIDConfig retrieves the open ID connection information
func GoogleOpenIDConfig() (clientID, authURL string, err error) {
	dDoc, err := googleGetDiscoveryDoc()
	if err != nil {
		return
	}
	authURL = dDoc.AuthorizationEndpoint

	clientID = googleClientID
	return
}

func googleGetDiscoveryDoc() (dDoc *googleDiscoveryDocument, err error) {
	//TODO: respect cache / wrap in mutex w/ timer?
	res, err := httpClient.Get(googleDiscoveryDocumentURL)
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

	dDoc = &googleDiscoveryDocument{}

	err = json.Unmarshal(val, dDoc)
	if err != nil {
		return nil, err
	}

	return dDoc, nil
}

func googleGetOauthSession(code, redirectURI string) (gSes *googleOauthSession, err error) {
	dDoc, err := googleGetDiscoveryDoc()
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Add("code", code)
	values.Add("client_id", googleClientID)
	values.Add("client_secret", googleClientSecret)
	values.Add("redirect_uri", redirectURI)
	values.Add("grant_type", "authorization_code")

	res, err := httpClient.Post(dDoc.TokenEndpoint, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
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

	gSes = &googleOauthSession{}

	err = json.Unmarshal(val, gSes)
	if err != nil {
		return nil, err
	}

	if gSes.Error != "" {
		return nil, fail.New(fmt.Sprintf("Error getting google session - %s : %s", gSes.Error, gSes.ErrorDescription))
	}

	return gSes, nil
}

func (gs *googleOauthSession) userData() (people *googlePeople, err error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/plus/v1/people/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+gs.AccessToken)

	res, err := httpClient.Do(req)
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

	people = &googlePeople{}

	err = json.Unmarshal(val, people)
	if err != nil {
		return nil, err
	}

	if people.Error != nil {
		return nil, people.Error
	}

	people.session = gs

	return people, nil
}

func (gs *googleOauthSession) getJWT() (*googleJWT, error) {
	if gs.jwt != nil {
		return gs.jwt, nil
	}

	if gs.IDToken == "" {
		return nil, fmt.Errorf("Invalid Google JWT")
	}

	parts := strings.Split(gs.IDToken, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Malformed Google JWT: %s", gs.IDToken)
	}

	idToken := parts[1]
	jwt := &googleJWT{}

	// Decode the ID token
	switch len(idToken) % 4 {
	case 2:
		idToken += "=="
	case 3:
		idToken += "="
	}

	b, err := base64.URLEncoding.DecodeString(idToken)
	if err != nil {
		return nil, fmt.Errorf("Malformed Google JWT: %v", err)
	}
	err = json.Unmarshal(b, jwt)
	if err != nil {
		return nil, fmt.Errorf("Malformed Google JWT: %v", err)
	}

	gs.jwt = jwt

	return jwt, nil
}

func (j *googleJWT) validate() error {
	if j.AppID != googleClientID {
		return fmt.Errorf("Invalid Google JWT, AppID mismatch")
	}

	if j.ISS != "accounts.google.com" && j.ISS != "https://accounts.google.com" {
		return fmt.Errorf("Invalid Google JWT, invalid ISS")
	}
	if time.Unix(j.Expires, 0).Before(time.Now()) {
		return fmt.Errorf("Invalid Google JWT, expired")
	}

	//TODO: Check googles cryptographic keys
	return nil
}

// GoogleUser gets a user (if possible)
// from the passed in google code
func GoogleUser(code, redirectURI string) (*User, error) {
	gSes, err := googleGetOauthSession(code, redirectURI)
	if err != nil {
		return nil, err
	}

	jwt, err := gSes.getJWT()
	if err != nil {
		return nil, err
	}

	//Lookup user
	usr, err := userGetGoogle(jwt.UserID)
	if err == ErrUserNotFound {
		data, err := gSes.userData()
		if err != nil {
			return nil, err
		}
		usrFail := fail.NewFromErr(ErrUserNeedUsername, map[string]string{
			"username":    urlify(data.DisplayName).make(),
			"email":       jwt.Email,
			"idToken":     gSes.IDToken,
			"accessToken": gSes.AccessToken,
		})
		return nil, usrFail
	}

	if err != nil {
		return nil, err
	}

	return usr, nil
}

// GoogleNewUser creates a new google user with the passed in username and email,
// using the passed in temporary token
func GoogleNewUser(username, email, googleID, idToken, accessToken string) (*User, error) {
	gSes := &googleOauthSession{
		AccessToken: accessToken,
		IDToken:     idToken,
	}

	jwt, err := gSes.getJWT()
	if err != nil {
		return nil, err
	}

	err = jwt.validate()
	if err != nil {
		return nil, err
	}

	u, err := googleNewUser(username, email, gSes)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func googleNewUser(username, email string, gSes *googleOauthSession) (*User, error) {
	//Get email address and name
	data, err := gSes.userData()
	if err != nil {
		return nil, err
	}

	jwt, err := gSes.getJWT()
	if err != nil {
		return nil, err
	}

	if email == "" {
		email = jwt.Email
	}

	u, err := userNewGoogle(username, email, data.DisplayName, jwt.UserID)
	if err != nil {
		return nil, err
	}

	return u, nil
}
