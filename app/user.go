// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"fmt"
	"math"
	"net/mail"
	"path"
	"regexp"
	"strings"
	"time"

	"git.townsourced.com/townsourced/townsourced/app/email"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
	"golang.org/x/crypto/bcrypt"
)

// User is a townsourced user
type User struct {
	Username       data.Key        `json:"username,omitempty" gorethink:",omitempty"`
	Email          string          `json:"email,omitempty" gorethink:",omitempty"`
	EmailSearch    string          `json:"-" gorethink:",omitempty"` // for case insensitive searching, sending emails should use the email field
	EmailValidated bool            `json:"emailValidated,omitempty"`
	GoogleID       string          `json:"googleID,omitempty"`
	FacebookID     string          `json:"facebookID,omitempty"`
	TwitterID      string          `json:"twitterID,omitempty"`
	Name           string          `json:"name,omitempty"`
	Password       []byte          `json:"-" gorethink:",omitempty"`
	HasPassword    bool            `json:"hasPassword,omitempty"`
	TownKeys       []data.KeyWhen  `json:"townKeys,omitempty"`
	Badges         []int           `json:"badges,omitempty" gorethink:",omitempty"` // Awards given by townsourced (i.e. 5 year badge)
	Stamps         []data.Key      `json:"stamps,omitempty" gorethink:",omitempty"`
	ProfileImage   data.UUID       `json:"profileImage,omitempty" gorethink:",omitempty"`
	ProfileIcon    data.UUID       `json:"profileIcon,omitempty" gorethink:",omitempty"`
	Admin          bool            `json:"admin,omitempty"` // Only set directly in DB currently
	SavedPosts     []data.UUIDWhen `json:"savedPosts,omitempty"`

	NotifyPost    bool `json:"notifyPost,omitempty"`
	NotifyComment bool `json:"notifyComment,omitempty"`

	EmailPrivateMsg     bool `json:"emailPrivateMsg,omitempty"`
	EmailPostMention    bool `json:"emailPostMention,omitempty"`
	EmailCommentMention bool `json:"emailCommentMention,omitempty"`
	EmailCommentReply   bool `json:"emailCommentReply,omitempty"`
	EmailPostComment    bool `json:"emailPostComment,omitempty"`

	data.Version

	privateCleared bool // set when the private data is cleared, prevents updates from happening if private data is missing
}

const (
	passwordBcryptCost = 12
	passwordMinLength  = 10

	userImageWidth  = 300
	userImageHeight = 300
	userIconWidth   = 32
	userIconHeight  = 32

	//UsernameSelf is the username that is reserved for self access
	UsernameSelf = "me"

	userMatchMax    = 50
	userMatchMinLen = 2

	userEmailTokenExpire = 24 * time.Hour

	//UserForgotPasswordPath is the url path used in building the forgot password url
	UserForgotPasswordPath = "forgotpassword"
	//UserConfirmEmailPath is the url path used in building the confirm email url
	UserConfirmEmailPath = "confirmemail"
)

var (
	// ErrUserLogonFailure is when a user fails a login attempt
	ErrUserLogonFailure = fail.New("Invalid user and / or password")
	// ErrUserExists is when a user already exists
	ErrUserExists = fail.New("User already exists")
	//ErrUserEmailExists is when a particular email address is already in use
	ErrUserEmailExists = fail.New("Email already in use")
	// ErrUserNotFound is when a user is not found
	ErrUserNotFound = fail.New("User not found")
	// ErrUserEmailNotFound is when a user is not found by the given email
	ErrUserEmailNotFound = fail.New("No user is was found with this email address")
	// ErrUserPassShort is when a users password is too short
	ErrUserPassShort = fail.New(fmt.Sprintf("Invalid password.  Your password must be at least %d characters long",
		passwordMinLength))

	// ErrUserNeedUsername is when a townsourced user doesn't exist for this 3rdparty credential, and a new
	//  username is needed to create the user
	ErrUserNeedUsername = fail.New("Username needed")

	// ErrUserNoDisconnect is when a user can't disconnect their credentials from a 3rd party like twitter because they have
	// no other authentication method
	ErrUserNoDisconnect = fail.New("Can't disconnect user because there is no other way to log in")

	// ErrUserPrivatePosts is returned when trying to view posts that a user does not have permissions to see
	ErrUserPrivatePosts = fail.New("You do not have permissions to view these posts")
	// ErrUserInvalidEmailToken is returned when a user tries to reset a password with an invalid or expired reset token
	ErrUserInvalidEmailToken = fail.New("This email token is invalid or has expired")
)

var emailTest = regexp.MustCompile(".+@.+\\..+")

//TODO: Financials - payments and payouts
//TODO: Ratings, separate table?

// UserNew creates a new user
func UserNew(username data.Key, email, password string) (*User, error) {
	u := makeUser(username, email)
	err := u.setPassword(password)
	if err != nil {
		return nil, err
	}
	err = u.insert()
	if err != nil {
		return nil, err
	}

	return u, nil
}

// makse a new user with defaults
func makeUser(username data.Key, email string) *User {
	return &User{
		Username:            username,
		Email:               email,
		EmailSearch:         strings.ToLower(email),
		TownKeys:            []data.KeyWhen{data.NewKey(AnnouncementTown).KeyWhen()},
		NotifyPost:          true,
		NotifyComment:       true,
		EmailPrivateMsg:     true,
		EmailPostMention:    false,
		EmailCommentMention: false,
		EmailCommentReply:   true,
		EmailPostComment:    true,
	}
}

func (u *User) insert() error {
	err := data.UserGet(&User{}, u.Username)
	if err == nil {
		return ErrUserExists
	}

	if err != data.ErrNotFound {
		return err
	}

	//no user found
	err = u.validate()
	if err != nil {
		return err
	}

	u.Rev() // generate new version

	err = data.UserInsert(u)
	if err != nil {
		return err
	}

	err = u.sendEmailConfirmation(true)
	if err != nil {
		return err
	}

	sub, msg, err := messages.use("msgUserWelcome").Execute(u)
	if err != nil {
		return err
	}

	return notificationNew(data.EmptyKey, u.Username, sub, msg)
}

// UserEmailExists checks if the passed in email is already in use
func UserEmailExists(email string) error {
	if email == "" {
		return fail.New("An email is required")
	}

	if !emailTest.MatchString(email) {
		return fail.New("Invalid email address format.", email)
	}
	err := data.UserGetEmail(&User{}, email)
	if err == data.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	return fail.NewFromErr(ErrUserEmailExists, email)
}

func (u *User) validate() error {
	if u.Username == "" {
		return fail.New("A username is required")
	}

	if u.Username == UsernameSelf {
		return fail.New("Invalid username")
	}

	if !urlify(string(u.Username)).is() {
		return fail.New("A username can only contain letters, numbers and dashes")
	}

	if len(u.Password) == 0 && u.GoogleID == "" && u.TwitterID == "" && u.FacebookID == "" {
		return fail.New("User must have a password, or 3rd party credential")
	}

	return UserEmailExists(u.Email)
}

func (u *User) setPassword(newPassword string) error {
	if newPassword == "" {
		if u.GoogleID != "" || u.TwitterID != "" || u.FacebookID != "" {
			return nil
		}
	}

	if len(newPassword) < passwordMinLength {
		return ErrUserPassShort
	}

	var err error

	u.Password, err = bcrypt.GenerateFromPassword([]byte(newPassword), passwordBcryptCost)
	if err != nil {
		return err
	}

	u.HasPassword = true
	return nil
}

// ClearPrivate clears the private information from
// the user record
func (u *User) ClearPrivate() {
	u.Email = ""
	u.EmailSearch = ""
	u.GoogleID = ""
	u.FacebookID = ""
	u.TwitterID = ""
	u.Password = nil
	u.HasPassword = false
	u.NotifyPost = false
	u.NotifyComment = false
	u.EmailPrivateMsg = false
	u.EmailPostMention = false
	u.EmailCommentMention = false
	u.EmailCommentReply = false
	u.EmailPostComment = false
	u.EmailValidated = false

	u.SavedPosts = nil
	u.privateCleared = true
}

// UserGet returns a user including private information
func UserGet(username data.Key) (*User, error) {
	u := &User{}
	err := data.UserGet(u, username)
	if err == data.ErrNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UserLogin logs in a user via their password and either their username or email
func UserLogin(usernameOrEmail, password string) (*User, error) {
	u := &User{}

	if strings.Contains(usernameOrEmail, "@") {
		//login with email
		err := data.UserGetEmail(u, usernameOrEmail)
		if err == data.ErrNotFound {
			return nil, ErrUserEmailNotFound // can test for email, but not username
		}
		if err != nil {
			return nil, err
		}
	} else {
		// login with username
		err := data.UserGet(u, data.NewKey(usernameOrEmail))
		if err == data.ErrNotFound {
			return nil, ErrUserLogonFailure
		}
		if err != nil {
			return nil, err
		}
	}

	if u.HasPassword == false || len(u.Password) == 0 {
		//User isn't a password based user, and can't login with a password
		return nil, ErrUserLogonFailure
	}

	err := u.login(password)
	if err != nil {
		return nil, err
	}
	return u, nil

}

func (u *User) login(password string) error {
	if !u.HasPassword || len(u.Password) == 0 {
		if u.GoogleID == "" && u.FacebookID == "" && u.TwitterID == "" {
			return ErrUserLogonFailure
		}
		// Can't auth password for non-password based users
		return nil
	}

	//compare password
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return ErrUserLogonFailure
		}
		return err
	}

	return nil
}

// Towns returns the towns a user is a member of
func (u *User) Towns() ([]Town, error) {
	towns, err := TownsGet(data.KeyWhenSlice(u.TownKeys).Keys()...)
	if err != nil {
		return nil, err
	}

	return towns, nil
}

// Update updates the user
func (u *User) Update() error {
	if u.privateCleared {
		panic("Cannot update user after private data has been cleared!")
	}
	return data.UserUpdate(u, u.Username)
}

//Facebook
func userNewFacebook(username, email, name, facebookID string) (*User, error) {
	u := makeUser(data.NewKey(username), email)
	u.Name = name
	u.FacebookID = facebookID

	err := data.UserGetFacebook(&User{}, facebookID)
	if err == nil {
		return nil, ErrUserExists
	}

	if err != data.ErrNotFound {
		return nil, err
	}
	err = u.insert()
	if err != nil {
		return nil, err
	}

	return u, nil

}

func userGetFacebook(facebookID string) (*User, error) {
	u := &User{}
	err := data.UserGetFacebook(u, facebookID)
	if err == data.ErrNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

//Google
func userNewGoogle(username, email, name, googleID string) (*User, error) {
	u := makeUser(data.NewKey(username), email)
	u.Name = name
	u.GoogleID = googleID

	err := data.UserGetGoogle(&User{}, googleID)
	if err == nil {
		return nil, ErrUserExists
	}

	if err != data.ErrNotFound {
		return nil, err
	}
	err = u.insert()
	if err != nil {
		return nil, err
	}
	return u, nil

}

func userGetGoogle(googleID string) (*User, error) {
	u := &User{}
	err := data.UserGetGoogle(u, googleID)
	if err == data.ErrNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

//Twitter
func userNewTwitter(username, email, name, twitterID string) (*User, error) {
	u := makeUser(data.NewKey(username), email)
	u.Name = name
	u.TwitterID = twitterID

	err := data.UserGetTwitter(&User{}, twitterID)
	if err == nil {
		return nil, ErrUserExists
	}

	if err != data.ErrNotFound {
		return nil, err
	}

	err = u.insert()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func userGetTwitter(twitterID string) (*User, error) {
	u := &User{}
	err := data.UserGetTwitter(u, twitterID)
	if err == data.ErrNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UserGetMatch returns a all users who's username starts with match
func UserGetMatch(match string, limit int) ([]User, error) {
	var users []User

	if len(match) < userMatchMinLen {
		return users, nil
	}

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(userMatchMax)))

	if len(match) > data.MaxKeyLength {
		match = match[:data.MaxKeyLength]
	}

	err := data.UserGetMatching(&users, match, limit)
	if err == data.ErrNotFound {
		return users, nil
	}

	if err != nil {
		return nil, err
	}

	return users, nil
}

// Etag returns an appropriate string to use as an HTTP etag
func (u *User) Etag() string {
	return u.Version.Ver()
}

// DisplayName returns the users display name.  If they haven't specified a name, it'll return the username
func (u *User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return string(u.Username)
}

// SetProfileImage sets the user's profile image to the specified dimensions
// resizes it to the max profile image size, and creates an icon
func (u *User) SetProfileImage(image *Image, x0, y0, x1, y1 float64) error {
	if u.Username != image.OwnerKey {
		return ErrImageNotOwner
	}

	err := image.decode()
	if err != nil {
		return err
	}

	if x0 < 0 || y0 < 0 || x1 < 0 || y1 < 0 {
		x0 = 0
		x1 = float64(image.image.Bounds().Dx())
		y0 = 0
		y1 = float64(image.image.Bounds().Dy())
	}

	if (x1 - x0) < userIconWidth {
		x0 = 0
		x1 = float64(image.image.Bounds().Dx())
	}

	if (y1 - y0) < userIconHeight {
		y0 = 0
		y1 = float64(image.image.Bounds().Dy())
	}

	//delete any existing profile image
	if u.ProfileImage != data.EmptyUUID {
		err = data.ImageDelete(u.ProfileImage)
		if err != nil {
			return err
		}
	}

	if u.ProfileIcon != data.EmptyUUID {
		err = data.ImageDelete(u.ProfileIcon)
		if err != nil {
			return err
		}
	}

	ratio := (x1 - x0) / (y1 - y0)
	//give a little wiggle room for rounding
	if ratio < 0.95 || ratio > 1.05 {
		min := math.Min(float64(image.image.Bounds().Dx()), float64(image.image.Bounds().Dy()))

		err = image.cropCenter(u, int(min), int(min))
		if err != nil {
			return err
		}
	} else {
		err = image.crop(u, round(x0), round(y0), round(x1), round(y1))
		if err != nil {
			return err
		}
	}

	err = image.resize(u, userImageWidth, userImageHeight)
	if err != nil {
		return err
	}

	image.InUse = true
	err = image.update()
	if err != nil {
		return err
	}

	u.ProfileImage = image.Key

	err = image.resize(u, userIconWidth, userIconHeight)
	if err != nil {
		return err
	}

	err = image.encode()
	if err != nil {
		return err
	}

	icon, err := imageNew(u, image.ContentType, image.Data, true)
	if err != nil {
		return err
	}

	u.ProfileIcon = icon.Key
	return nil
}

// SetVer  prepares the user data for an update
// based on the passed in vertag, if the vertag doesn't
// match the current record, then the update won't complete
func (u *User) SetVer(verTag string) {
	u.VerTag = verTag
}

// SetName sets the users name
func (u *User) SetName(name string) {
	u.Name = name
}

// SetEmail updates the users email address, requires their current password
// to be passed in
func (u *User) SetEmail(email string, password string) error {
	err := UserEmailExists(email)
	if err != nil {
		return err
	}

	err = u.login(password)
	if err != nil {
		return err
	}

	u.Email = email
	u.EmailSearch = strings.ToLower(email)
	u.EmailValidated = false

	return u.SendEmailConfirmation()
}

// SetPassword updates a users password
func (u *User) SetPassword(currentPass, newPass string) error {
	err := u.login(currentPass)
	if err != nil {
		return err
	}

	return u.setPassword(newPass)
}

// JoinTown joins a user a to a town
func (u *User) JoinTown(town *Town) error {
	if town.IsMember(u) {
		// already a member
		return nil
	}

	if town.Private {
		//check if invited
		if !town.Invited(u) {
			return ErrTownNoInvite
		}
	}

	u.TownKeys = append(u.TownKeys, town.Key.KeyWhen())
	return nil
}

// LeaveTown is when a user leaves a town
func (u *User) LeaveTown(town *Town) {
	for i := range u.TownKeys {
		if u.TownKeys[i].Key == town.Key {
			u.TownKeys = append(u.TownKeys[:i], u.TownKeys[i+1:]...)
			return
		}
	}
}

// DisconnectFacebook disconnects this user from Facebook credentials
func (u *User) DisconnectFacebook() error {
	if u.TwitterID == "" && u.GoogleID == "" && !u.HasPassword {
		return ErrUserNoDisconnect
	}

	u.FacebookID = ""
	return nil
}

// DisconnectGoogle disconnects this user from Google credentials
func (u *User) DisconnectGoogle() error {
	if u.TwitterID == "" && u.FacebookID == "" && !u.HasPassword {
		return ErrUserNoDisconnect
	}

	u.GoogleID = ""
	return nil
}

// DisconnectTwitter disconnects this user from Twitter credentials
func (u *User) DisconnectTwitter() error {
	if u.GoogleID == "" && u.FacebookID == "" && !u.HasPassword {
		return ErrUserNoDisconnect
	}

	u.TwitterID = ""
	return nil
}

// LinkFacebook links a facebook account to an existing user
func (u *User) LinkFacebook(redirectURI, code string) error {
	if u.FacebookID != "" {
		return fail.New("User is already linked to a facebook account")
	}
	fbSes, err := facebookGetSession(redirectURI, code)
	if err != nil {
		return err
	}

	//Lookup user
	usr, err := userGetFacebook(fbSes.userID)
	if err != nil && err != ErrUserNotFound {
		return err
	}

	if err == nil {
		if usr.Username == u.Username {
			return nil // nothing to do
		}

		return fail.New("These facebook credentials are already associated to a townsourced user")
	}

	u.FacebookID = fbSes.userID
	return nil
}

// LinkGoogle links a google account to an existing user
func (u *User) LinkGoogle(redirectURI, code string) error {
	if u.GoogleID != "" {
		return fail.New("User is already linked to a google account")
	}

	gSes, err := googleGetOauthSession(code, redirectURI)
	if err != nil {
		return err
	}

	jwt, err := gSes.getJWT()
	if err != nil {
		return err
	}

	//Lookup user
	usr, err := userGetGoogle(jwt.UserID)
	if err != nil && err != ErrUserNotFound {
		return err
	}

	if err == nil {
		if usr.Username == u.Username {
			return nil // nothing to do
		}
		return fail.New("These google credentials are already associated to a townsourced user")
	}

	u.GoogleID = jwt.UserID
	return nil
}

// LinkTwitter links a twitter account to an existing user
func (u *User) LinkTwitter(stateToken, verificationCode string) error {
	if u.TwitterID != "" {
		return fail.New("User is already linked to a twitter account")
	}

	session := &twitterSession{}

	err := data.TempTokenGet(session, stateToken)
	if err != nil {
		return err
	}

	accessToken, err := twitterConsumer.AuthorizeToken(session.OauthToken, verificationCode)
	if err != nil {
		return err
	}

	session.OauthToken = nil
	session.AccessToken = accessToken

	userData, err := session.userData()
	if err != nil {
		return err
	}

	//Lookup user
	usr, err := userGetTwitter(userData.ID)
	if err != nil && err != ErrUserNotFound {
		return err
	}

	if err == nil {
		if usr.Username == u.Username {
			return nil // nothing to do
		}
		return fail.New("These twitter credentials are already associated to a townsourced user")
	}

	u.TwitterID = userData.ID

	return nil
}

// SendMessage sends a private message to another user
func (u *User) SendMessage(to *User, subject, message string) error {
	//TODO: Blocking users?
	err := notificationNew(u.Username, to.Username, subject, message)
	if err != nil {
		return err
	}

	if to.EmailPrivateMsg {
		sub, msg, err := messages.use("emailUserPrivateMessage").Execute(struct {
			From    *User
			To      *User
			Subject string
		}{
			From:    u,
			To:      to,
			Subject: subject,
		})

		if err != nil {
			return err
		}
		err = to.sendEmail(sub, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) sendEmail(subject, body string) error {
	if !u.EmailValidated {
		return nil
	}

	return email.Send(email.DefaultFrom, &mail.Address{
		Name:    u.DisplayName(),
		Address: u.Email,
	}, subject, body)
}

func (u *User) mentionPost(p *Post) error {
	if !u.NotifyPost {
		return nil
	}

	visible, err := p.Visible(u)
	if err != nil {
		return err
	}

	if !visible {
		return nil
	}

	sub, msg, err := messages.use("msgPostUserMention").Execute(p)
	if err != nil {
		return err
	}

	err = notificationNew(data.EmptyKey, u.Username, sub, msg)
	if err != nil {
		return err
	}

	if u.EmailPostMention {
		sub, msg, err = messages.use("emailPostMention").Execute(struct {
			To   *User
			Post *Post
		}{
			To:   u,
			Post: p,
		})
		if err != nil {
			return err
		}

		err = u.sendEmail(sub, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *User) mentionComment(c *Comment) error {
	if !u.NotifyComment {
		return nil
	}

	post, err := c.Post()
	if err != nil {
		return err
	}

	visible, err := post.Visible(u)
	if err != nil {
		return err
	}

	if !visible {
		return nil
	}

	msgData := struct {
		To      *User
		Post    *Post
		Comment *Comment
	}{
		To:      u,
		Post:    post,
		Comment: c,
	}

	sub, msg, err := messages.use("msgCommentUserMention").Execute(msgData)
	if err != nil {
		return err
	}

	err = notificationNew(data.EmptyKey, u.Username, sub, msg)
	if err != nil {
		return err
	}

	if u.EmailCommentMention {
		sub, msg, err = messages.use("emailCommentMention").Execute(msgData)
		if err != nil {
			return err
		}
		err = u.sendEmail(sub, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// Posts returns a users posts since the passed in time
func (u *User) Posts(who *User, status string, since time.Time, limit int) ([]*Post, error) {
	var posts []*Post

	if status != "" { //blank is all statuses
		err := postStatusCheck(status)
		if err != nil {
			return nil, err
		}
	}

	public := who == nil || u.Username != who.Username

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(postMaxRetrieve)))

	err := data.PostGetByUser(&posts, u.Username, status, public, since, limit)
	if err == data.ErrNotFound {
		return posts, nil
	}
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetSavedPosts returns a users saved posts
func (u *User) GetSavedPosts(who *User, status string, from, limit int) ([]*Post, error) {
	var posts []*Post

	if status != "" { //blank is all statuses
		err := postStatusCheck(status)
		if err != nil {
			return nil, err
		}
	}

	if who == nil || u.Username != who.Username {
		return nil, ErrUserPrivatePosts
	}

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(postMaxRetrieve)))

	if len(u.SavedPosts) == 0 {
		return posts, nil
	}

	err := data.PostGetUserSaved(&posts, u.Username, status, from, limit)
	if err != nil {
		return nil, err
	}
	return posts, nil

}

// savedPostIndex returns the index of the saved post, and -1 if it's not saved
func (u *User) savedPostIndex(post *Post) int {
	for i := range u.SavedPosts {
		if u.SavedPosts[i].Key == post.Key {
			return i
		}
	}
	return -1
}

// SavePost saves / favorites a post to a user
func (u *User) SavePost(post *Post) {
	if u.savedPostIndex(post) != -1 {
		return
	}
	u.SavedPosts = append(u.SavedPosts, post.Key.UUIDWhen())
}

// RemoveSavedPost removes a post from a users saved posts list
func (u *User) RemoveSavedPost(post *Post) {
	i := u.savedPostIndex(post)
	if i == -1 {
		return
	}

	u.SavedPosts = append(u.SavedPosts[:i], u.SavedPosts[i+1:]...)
}

// Comments retrieves the comments made by this user
func (u *User) Comments(who *User, since time.Time, limit int) ([]Comment, error) {
	var comments []Comment

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(commentMaxRetrieve)))

	public := who == nil || who.Username != u.Username

	err := data.CommentsGetByUser(&comments, u.Username, public, since, limit)
	if err == data.ErrNotFound {
		return comments, nil
	}
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// ForgotPassword is when a user forgets their password and needs to be sent a reset password email
func ForgotPassword(usernameOrEmail string) error {
	u := &User{}

	if strings.Contains(usernameOrEmail, "@") {
		err := data.UserGetEmail(u, usernameOrEmail)
		if err == data.ErrNotFound {
			return ErrUserEmailNotFound // can test for email, but not username
		}
		if err != nil {
			return err
		}
	} else {
		// login with username
		err := data.UserGet(u, data.NewKey(usernameOrEmail))
		if err == data.ErrNotFound {
			// do nothing
			return nil
		}
		if err != nil {
			return err
		}
	}

	tokenURL := ""
	if u.HasPassword {
		token := Random(256)
		tokenURL = baseURL + "/" + path.Join(UserForgotPasswordPath, token)

		err := data.TempTokenSet(u.Username, token, userEmailTokenExpire)
		if err != nil {
			return err
		}
	}

	sub, body, err := messages.use("emailUserForgotPassword").Execute(struct {
		URL  string
		User *User
	}{
		URL:  tokenURL,
		User: u,
	})

	err = email.Send(email.DefaultFrom, &mail.Address{
		Name:    u.DisplayName(),
		Address: u.Email,
	}, sub, body)

	if err != nil {
		return fmt.Errorf("Error sending forgot password email: %s", err)
	}

	return nil
}

// RetrievePasswordToken retrieves a previously requested password token
func RetrievePasswordToken(token string) (data.Key, error) {
	username := data.EmptyKey
	err := data.TempTokenGet(&username, token)

	if err != nil {
		return data.EmptyKey, err
	}

	return username, nil
}

// ResetPassword resets a users password from a forgot password request
func ResetPassword(resetToken, newPass string) (*User, error) {
	username, err := RetrievePasswordToken(resetToken)
	if err != nil {
		return nil, ErrUserInvalidEmailToken
	}

	u, err := UserGet(username)
	if err != nil {
		return nil, ErrUserInvalidEmailToken
	}

	err = u.setPassword(newPass)
	if err != nil {
		return nil, err
	}

	err = u.Update()
	if err != nil {
		return nil, err
	}

	// expire token
	err = data.TempTokenExpire(resetToken)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// SendEmailConfirmation sends an email to validate a users email address
func (u *User) SendEmailConfirmation() error {
	return u.sendEmailConfirmation(false)
}

func (u *User) sendEmailConfirmation(showWelcome bool) error {
	token := Random(256)
	tokenURL := baseURL + "/" + path.Join(UserConfirmEmailPath, token)

	err := data.TempTokenSet(u.Username, token, userEmailTokenExpire)
	if err != nil {
		return err
	}

	sub, body, err := messages.use("emailUserConfirmEmail").Execute(struct {
		Name    string
		URL     string
		Welcome bool
	}{
		Name:    u.DisplayName(),
		URL:     tokenURL,
		Welcome: showWelcome,
	})

	if err != nil {
		return err
	}

	err = email.Send(email.DefaultFrom, &mail.Address{
		Name:    u.DisplayName(),
		Address: u.Email,
	}, sub, body)
	if err != nil {
		return fmt.Errorf("Error sending email confirmation: %s", err)
	}

	return nil
}

// ConfirmEmail confirms via the passed in token that a user's email address is working, in that they
// must have recieved the email containing the token URL.
func ConfirmEmail(token string) error {
	username := data.EmptyKey
	err := data.TempTokenGet(&username, token)

	if err != nil {
		return ErrUserInvalidEmailToken
	}

	u, err := UserGet(username)
	if err != nil {
		return ErrUserInvalidEmailToken
	}

	if !u.EmailValidated {
		u.EmailValidated = true
		err = u.Update()
		if err != nil {
			return err
		}
	}

	// expire token
	err = data.TempTokenExpire(token)
	if err != nil {
		return err
	}

	return nil
}
