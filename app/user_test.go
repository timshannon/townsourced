// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package app_test

import (
	"strings"
	"time"

	. "git.townsourced.com/townsourced/check"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
)

func usersEqual(c *C, user, other *app.User) {
	c.Assert(user.Name, Equals, other.Name)
	c.Assert(user.Email, Equals, other.Email)
	c.Assert(user.VerTag, Equals, other.VerTag)
}

//User Test Suite
type UserSuite struct {
	*testData
}

var _ = Suite(&UserSuite{testData: &testData{}})

func (s *UserSuite) SetUpTest(c *C) {
	// add test user
	s.testData.setup(c)
}

func (s *UserSuite) TearDownTest(c *C) {
	s.testData.teardown(c)
}

func (s *UserSuite) TestUserNew(c *C) {
	var username = data.NewKey("testOtherUser")

	// bad password
	_, err := app.UserNew(username, "test@test.townsourced.com", "short")
	c.Assert(err, ErrorMatches, app.ErrUserPassShort.Error())

	// bad email
	_, err = app.UserNew(username, "bad@email", s.userPassword)
	c.Assert(err, ErrorMatches, "Invalid email address format.")

	// existing user
	_, err = app.UserNew(s.user.Username, "test@test.townsourced.com", s.userPassword)
	c.Assert(err, ErrorMatches, app.ErrUserExists.Error())

	// existing email
	_, err = app.UserNew(username, s.user.Email, s.userPassword)
	c.Assert(err, ErrorMatches, app.ErrUserEmailExists.Error())

	// success
	u, err := app.UserNew(username, "test@test.townsourced.com", s.userPassword)
	defer s.testData.deleteUser(c, u)
	c.Assert(err, Equals, nil)

	other, err := app.UserGet(username)
	c.Assert(err, Equals, nil)

	usersEqual(c, u, other)
}

func (s *UserSuite) TestUserEmailExists(c *C) {
	// empty
	c.Assert(app.UserEmailExists(""), Not(Equals), nil)

	// invalid email

	c.Assert(app.UserEmailExists("invalid email"), Not(Equals), nil)

	// non-existing email
	c.Assert(app.UserEmailExists("invalid.email@invalidemail.com"), Equals, nil)

	// existing email
	c.Assert(app.UserEmailExists(s.user.Email), ErrorMatches, app.ErrUserEmailExists.Error())
}

func (s *UserSuite) TestUserClearPrivate(c *C) {
	u := s.user
	u.ClearPrivate()

	c.Assert(u.Email, Equals, "")
	c.Assert(u.EmailSearch, Equals, "")
	c.Assert(u.GoogleID, Equals, "")
	c.Assert(u.FacebookID, Equals, "")
	c.Assert(u.TwitterID, Equals, "")
	c.Assert(u.Password, HasLen, 0)
	c.Assert(u.HasPassword, Equals, false)
	c.Assert(u.NotifyPost, Equals, false)
	c.Assert(u.NotifyComment, Equals, false)
	c.Assert(u.EmailPrivateMsg, Equals, false)
	c.Assert(u.EmailPostMention, Equals, false)
	c.Assert(u.EmailCommentMention, Equals, false)
	c.Assert(u.EmailCommentReply, Equals, false)
	c.Assert(u.EmailPostComment, Equals, false)
	c.Assert(u.EmailValidated, Equals, false)

	c.Assert(u.SavedPosts, HasLen, 0)
}

func (s *UserSuite) TestUserGet(c *C) {
	other, err := app.UserGet(s.user.Username)
	c.Assert(err, Equals, nil)
	usersEqual(c, s.user, other)
}

func (s *UserSuite) TestUserLogin(c *C) {
	// bad password
	_, err := app.UserLogin(string(s.user.Username), "badpassword")
	c.Assert(err, ErrorMatches, app.ErrUserLogonFailure.Error())

	// bad username
	_, err = app.UserLogin("badusername", "badpassword")
	c.Assert(err, ErrorMatches, app.ErrUserLogonFailure.Error())

	//username login
	other, err := app.UserLogin(string(s.user.Username), s.userPassword)
	c.Assert(err, Equals, nil)
	usersEqual(c, s.user, other)

	//email login
	other = nil
	other, err = app.UserLogin(s.user.Email, s.userPassword)
	c.Assert(err, Equals, nil)
	usersEqual(c, s.user, other)

	// case insensitive login
	other = nil
	other, err = app.UserLogin(strings.ToUpper(s.user.Email), s.userPassword)
	c.Assert(err, Equals, nil)
	usersEqual(c, s.user, other)

	other = nil
	other, err = app.UserLogin(strings.ToUpper(string(s.user.Username)), s.userPassword)
	c.Assert(err, Equals, nil)
	usersEqual(c, s.user, other)
}

func (s *UserSuite) TestUserTowns(c *C) {
	//test for Annoucement town
	towns, err := s.user.Towns()
	c.Assert(err, Equals, nil)
	// brand new user should only be a member in the Annoucements town
	c.Assert(towns, HasLen, 1)

	//Add and join town
	// test to make sure it exists

	c.Assert(s.user.JoinTown(s.town1), Equals, nil)

	towns, err = s.user.Towns()
	c.Assert(err, Equals, nil)
	c.Assert(towns, HasLen, 2)

	found := false
	for i := range towns {
		if towns[i].Key == s.town1.Key {
			found = true
			break
		}
	}

	c.Assert(found, Equals, true)
}

func (s *UserSuite) TestUserUpdate(c *C) {
	var newName = "updated name"

	s.user.Name = newName
	c.Assert(s.user.Update(), Equals, nil)

	other, err := app.UserGet(s.user.Username)
	c.Assert(err, Equals, nil)

	usersEqual(c, s.user, other)
}

func (s *UserSuite) TestUserUpdateAfterClear(c *C) {
	s.user.ClearPrivate()

	s.user.Name = "new name"
	c.Assert(func() { s.user.Update() }, PanicMatches, "Cannot update user after private data has been cleared!")
}

func (s *UserSuite) TestUserGetMatch(c *C) {
	users, err := app.UserGetMatch("", 10)
	c.Assert(err, Equals, nil)
	c.Assert(users, HasLen, 0)

	users, err = app.UserGetMatch(string(s.user.Username), 10)
	c.Assert(err, Equals, nil)
	c.Assert(users, HasLen, 1)
	c.Assert(users[0].Username, Equals, s.user.Username)

	// multiple matches

	other, err := app.UserNew(data.NewKey(string(s.user.Username)+"other"), "unittestuserother@townsourced.com",
		s.userPassword)
	defer s.testData.deleteUser(c, other)
	c.Assert(err, Equals, nil)

	users, err = app.UserGetMatch(string(s.user.Username), 10)
	c.Assert(err, Equals, nil)
	c.Assert(users, HasLen, 2)

	found := false
	for i := range users {
		if users[i].Username == other.Username {
			found = true
			break
		}
	}

	c.Assert(found, Equals, true)
}

func (s *UserSuite) TestUserEtag(c *C) {
	c.Assert(s.user.VerTag, Equals, s.user.Etag())
}

func (s *UserSuite) TestUserDisplayName(c *C) {
	var newName = "New Name"

	c.Assert(string(s.user.Username), Equals, s.user.DisplayName())

	s.user.Name = newName
	c.Assert(s.user.Name, Equals, s.user.DisplayName())
}

func (s *UserSuite) TestUserSetProfileImage(c *C) {
	//bad owner
	img := s.addImage(c, s.user)
	defer s.deleteImage(c, img)

	c.Assert(s.other.SetProfileImage(img, 0, 0, 100, 100), ErrorMatches, app.ErrImageNotOwner.Error())

	//coords too small
	c.Assert(s.user.SetProfileImage(img, 0, 0, 10, 10), Equals, nil)

	//coords negative
	img = s.addImage(c, s.user)
	defer s.deleteImage(c, img)

	c.Assert(s.user.SetProfileImage(img, 0, 0, -10, 10), Equals, nil)

	//invalid coord ratio
	img = s.addImage(c, s.user)
	defer s.deleteImage(c, img)

	c.Assert(s.user.SetProfileImage(img, 0, 0, 500, 100), Equals, nil)

	var checkInUse = func(username data.Key) {
		// confirm images exist and are in use
		cUser, err := app.UserGet(username)
		c.Assert(err, Equals, nil)

		//icon
		cImg, err := app.ImageGet(cUser.ProfileIcon)
		c.Assert(err, Equals, nil)
		c.Assert(cImg.InUse, Equals, true)
		c.Assert(cImg.OwnerKey, Equals, username)

		// image
		cImg, err = app.ImageGet(cUser.ProfileImage)
		c.Assert(err, Equals, nil)
		c.Assert(cImg.InUse, Equals, true)
		c.Assert(cImg.OwnerKey, Equals, username)
	}

	// successful add
	img = s.addImage(c, s.user)
	defer s.deleteImage(c, img)

	c.Assert(s.user.SetProfileImage(img, 0, 0, 100, 100), Equals, nil)
	c.Assert(s.user.Update(), Equals, nil)
	checkInUse(s.user.Username)

	// successful replace existing

	rImg := s.addImage(c, s.user)
	defer s.deleteImage(c, rImg)

	c.Assert(s.user.SetProfileImage(rImg, 0, 0, 100, 100), Equals, nil)
	c.Assert(s.user.Update(), Equals, nil)
	checkInUse(s.user.Username)
}

func (s *UserSuite) TestUserVersion(c *C) {
	changed, err := app.UserGet(s.user.Username)
	c.Assert(err, Equals, nil)

	changed.SetName("New Name")
	c.Assert(changed.Update(), Equals, nil)

	c.Assert(changed.Ver(), Not(Equals), s.user.Ver())

	s.user.SetName("New Name")
	c.Assert(s.user.Update(), ErrorMatches, data.ErrVersionStale.Error())

	s.user.SetVer(changed.Ver())

	s.user.SetName("New Name")
	c.Assert(s.user.Update(), Equals, nil)

}

func (s *UserSuite) TestUserSetName(c *C) {
	var newName = "New Name"
	s.user.SetName(newName)
	c.Assert(s.user.Name, Equals, newName)
}

func (s *UserSuite) TestUserSetEmail(c *C) {
	// test existing email
	c.Assert(s.user.SetEmail(s.other.Email, s.userPassword), ErrorMatches, app.ErrUserEmailExists.Error())

	var newEmail = "newEmail@test.townsourced.com"
	// test invalid password
	c.Assert(s.user.SetEmail(newEmail, "badpassword"), ErrorMatches, app.ErrUserLogonFailure.Error())

	// test successful
	c.Assert(s.user.SetEmail(newEmail, s.userPassword), Equals, nil)
	c.Assert(s.user.Email, Equals, newEmail)

	// test validated email
	c.Assert(s.user.EmailValidated, Equals, false)
	c.Assert(s.user.EmailSearch, Equals, strings.ToLower(newEmail))
}

func (s *UserSuite) TestUserSetPassword(c *C) {
	// Test bad password
	c.Assert(s.user.SetPassword("badpassword", "newpassword"), ErrorMatches, app.ErrUserLogonFailure.Error())

	// test invalid password
	c.Assert(s.user.SetPassword(s.userPassword, "short"), ErrorMatches, app.ErrUserPassShort.Error())

	// success
	c.Assert(s.user.SetPassword(s.userPassword, "newpassword"), Equals, nil)

	// check has password
	c.Assert(s.user.HasPassword, Equals, true)
}

func (s *UserSuite) TestUserJoinTown(c *C) {
	c.Assert(s.user.JoinTown(s.town1), Equals, nil)

	towns, err := s.user.Towns()
	c.Assert(err, Equals, nil)
	c.Assert(towns, HasLen, 2)

	found := false
	for i := range towns {
		if towns[i].Key == s.town1.Key {
			found = true
			break
		}
	}

	c.Assert(found, Equals, true)
}

func (s *UserSuite) TestUserJoinPrivateTown(c *C) {
	c.Assert(s.user.JoinTown(s.townPrivate), ErrorMatches, app.ErrTownNoInvite.Error())

	c.Assert(s.townPrivate.AddInvite(s.moderator, s.user), Equals, nil)

	c.Assert(s.user.JoinTown(s.townPrivate), Equals, nil)
}

func (s *UserSuite) TestUserLeaveTown(c *C) {
	// test leaving before join
	s.user.LeaveTown(s.town1)
	c.Assert(s.user.Update(), Equals, nil)

	s.user.JoinTown(s.town1)
	c.Assert(s.user.Update(), Equals, nil)

	s.user.LeaveTown(s.town1)
	c.Assert(s.user.Update(), Equals, nil)

	towns, err := s.user.Towns()
	c.Assert(err, Equals, nil)
	c.Assert(towns, HasLen, 1)

	found := false
	for i := range towns {
		if towns[i].Key == s.town1.Key {
			found = true
			break
		}
	}

	c.Assert(found, Equals, false)
}

func (s *UserSuite) TestUserDisconnectFacebook(c *C) {
	// test facebook only disconnect
	s.user.HasPassword = false
	s.user.TwitterID = ""
	s.user.GoogleID = ""
	s.user.FacebookID = "testID"

	c.Assert(s.user.DisconnectFacebook(), ErrorMatches, app.ErrUserNoDisconnect.Error())

	s.user.HasPassword = true
	c.Assert(s.user.DisconnectFacebook(), Equals, nil)
}

func (s *UserSuite) TestUserDisconnectGoogle(c *C) {
	// test facebook only disconnect
	s.user.HasPassword = false
	s.user.TwitterID = ""
	s.user.FacebookID = ""
	s.user.GoogleID = "testID"

	c.Assert(s.user.DisconnectGoogle(), ErrorMatches, app.ErrUserNoDisconnect.Error())

	s.user.HasPassword = true
	c.Assert(s.user.DisconnectGoogle(), Equals, nil)
}

func (s *UserSuite) TestUserDisconnectTwitter(c *C) {
	// test facebook only disconnect
	s.user.HasPassword = false
	s.user.GoogleID = ""
	s.user.FacebookID = ""
	s.user.TwitterID = "testID"

	c.Assert(s.user.DisconnectTwitter(), ErrorMatches, app.ErrUserNoDisconnect.Error())

	s.user.HasPassword = true
	c.Assert(s.user.DisconnectTwitter(), Equals, nil)
}

//TODO: LinkGoogle,Twitter,Facebook  Test oauth?  How to fake this?

func (s *UserSuite) TestUserSendMessage(c *C) {
	var subject = "test subject"

	c.Assert(s.user.SendMessage(s.other, subject, "test message"), Equals, nil)

	notifications, err := s.other.UnreadNotifications(time.Time{}, 10)

	c.Assert(err, Equals, nil)

	found := false

	for i := range notifications {
		if notifications[i].Subject == subject {
			found = true
			break
		}
	}

	c.Assert(found, Equals, true)
}

func (s *UserSuite) TestUserPosts(c *C) {
	//test not owner, published only
	posts, err := s.user.Posts(s.other, "", time.Time{}, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 4)

	//test bad status
	_, err = s.user.Posts(s.user, "badStatus", time.Time{}, 10)
	c.Assert(err, ErrorMatches, app.ErrPostBadStatus.Error())

	//test blank status
	posts, err = s.user.Posts(s.user, "", time.Time{}, 10)
	c.Assert(err, Equals, nil)

	c.Assert(posts, HasLen, 6)

	foundDraft, foundPub, foundClosed := false, false, false

	for i := range posts {
		if posts[i].Key == s.postPub.Key {
			foundPub = true
		}

		if posts[i].Key == s.postDraft.Key {
			foundDraft = true
		}
		if posts[i].Key == s.postClosed.Key {
			foundClosed = true
		}
	}

	c.Assert(foundPub, Equals, true)
	c.Assert(foundDraft, Equals, true)
	c.Assert(foundClosed, Equals, true)

	//test draft status

	posts, err = s.user.Posts(s.user, app.PostStatusDraft, time.Time{}, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 1)
	c.Assert(posts[0].Key, Equals, s.postDraft.Key)

	//test negative limit

	posts, err = s.user.Posts(s.user, "", time.Time{}, -10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 1)
}

func (s *UserSuite) TestUserGetSavedPosts(c *C) {
	s.user.SavePost(s.post1)
	s.user.SavePost(s.post2)
	c.Assert(s.user.Update(), Equals, nil)

	// test not owner
	_, err := s.user.GetSavedPosts(s.other, "", 0, 10)
	c.Assert(err, ErrorMatches, app.ErrUserPrivatePosts.Error())

	// test negative limit
	posts, err := s.user.GetSavedPosts(s.user, "", 0, -10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 1)

	posts, err = s.user.GetSavedPosts(s.user, "", 0, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 2)

	found1, found2 := false, false

	for i := range posts {
		if posts[i].Key == s.post1.Key {
			found1 = true
		}

		if posts[i].Key == s.post2.Key {
			found2 = true
		}
	}

	c.Assert(found1, Equals, true)
	c.Assert(found2, Equals, true)
}

func (s *UserSuite) TestUserSavePost(c *C) {
	otherPost, err := app.PostNew("test title", "test content", "buysell", app.PostFormatStandard, s.other,
		[]data.Key{s.town1.Key}, nil, data.EmptyUUID, false, false, false)
	defer s.testData.deletePost(c, otherPost)
	c.Assert(err, Equals, nil)

	//save own post
	s.user.SavePost(s.post1)
	c.Assert(s.user.Update(), Equals, nil)

	//save other post
	s.user.SavePost(otherPost)
	c.Assert(s.user.Update(), Equals, nil)

	posts, err := s.user.GetSavedPosts(s.user, "", 0, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 2)

	foundOwn, foundOther := false, false

	for i := range posts {
		if posts[i].Key == s.post1.Key {
			foundOwn = true
		}
		if posts[i].Key == otherPost.Key {
			foundOther = true
		}
	}
	c.Assert(foundOther, Equals, true)
	c.Assert(foundOwn, Equals, true)

	// save same post again
	s.user.SavePost(otherPost)
	c.Assert(s.user.Update(), Equals, nil)
	posts, err = s.user.GetSavedPosts(s.user, "", 0, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 2)
}

func (s *UserSuite) TestUserRemoveSavedPost(c *C) {
	s.user.SavePost(s.post1)
	c.Assert(s.user.Update(), Equals, nil)

	// test removing a post not saved
	s.user.RemoveSavedPost(s.post2)
	c.Assert(s.user.Update(), Equals, nil)

	posts, err := s.user.GetSavedPosts(s.user, "", 0, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 1)

	s.user.SavePost(s.post2)
	c.Assert(s.user.Update(), Equals, nil)

	posts, err = s.user.GetSavedPosts(s.user, "", 0, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 2)

	s.user.RemoveSavedPost(s.post2)
	c.Assert(s.user.Update(), Equals, nil)

	posts, err = s.user.GetSavedPosts(s.user, "", 0, 10)
	c.Assert(err, Equals, nil)
	c.Assert(posts, HasLen, 1)
	c.Assert(posts[0].Key, Equals, s.post1.Key)
}

func (s *UserSuite) TestUserComments(c *C) {
	// add 2 comments
	comment1, err := app.CommentNew(s.user, s.post1, "test comment 1")
	defer s.deleteComment(c, comment1)
	c.Assert(err, Equals, nil)

	comment2, err := app.CommentNew(s.user, s.post1, "test comment 2")
	defer s.deleteComment(c, comment2)
	c.Assert(err, Equals, nil)

	comments, err := s.user.Comments(s.user, time.Time{}, 10)
	c.Assert(err, Equals, nil)
	c.Assert(comments, HasLen, 2)

	found1, found2 := false, false

	for i := range comments {
		if comments[i].Key == comment1.Key {
			found1 = true
		}

		if comments[i].Key == comment2.Key {
			found2 = true
		}
	}

	c.Assert(found1, Equals, true)
	c.Assert(found2, Equals, true)

	// test min limit

	comments, err = s.user.Comments(s.user, time.Time{}, -10)
	c.Assert(err, Equals, nil)
	c.Assert(comments, HasLen, 1)
}

//TODO: Forgotpassword, reset password, email confirmation
