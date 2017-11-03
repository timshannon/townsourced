// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"io/ioutil"
	"testing"

	. "git.townsourced.com/townsourced/check"
	"git.townsourced.com/townsourced/elastic"
	"git.townsourced.com/townsourced/gomemcache/memcache"
	rt "git.townsourced.com/townsourced/gorethink"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
)

type dataClient struct{}

func (d *dataClient) database() *rt.Session {
	return data.DatabaseSession()
}
func (d *dataClient) search() *elastic.Client {
	return data.SearchClient()
}
func (d *dataClient) cache() *memcache.Client {
	return data.CacheClient()
}

type testData struct {
	client *dataClient

	user      *app.User
	other     *app.User
	moderator *app.User

	userPassword string

	town1       *app.Town
	town2       *app.Town
	townPrivate *app.Town

	postPub    *app.Post
	postDraft  *app.Post
	postClosed *app.Post

	post1 *app.Post
	post2 *app.Post
	post3 *app.Post
}

func (t *testData) setup(c *C) {
	var err error
	t.client = &dataClient{}

	//users
	t.userPassword = "unittestingpassword"
	t.user, err = app.UserNew(data.NewKey("unittestuser"), "unittestuser@townsourced.com", t.userPassword)
	c.Assert(err, Equals, nil)

	t.other, err = app.UserNew(data.NewKey("otherunittestuser"), "otherunittestuser@townsourced.com", t.userPassword)
	c.Assert(err, Equals, nil)

	t.moderator, err = app.UserNew(data.NewKey("moderatortestuser"), "moderatortestuser@townsourced.com", t.userPassword)
	c.Assert(err, Equals, nil)

	//towns
	t.town1, err = app.TownNew(data.NewKey("testTown1"), "Test Town 1", "Test Town 1", t.moderator, 100, 70, false)
	c.Assert(err, Equals, nil)

	t.town2, err = app.TownNew(data.NewKey("testTown2"), "Test Town 2", "Test Town 2", t.moderator, 100, 70, false)
	c.Assert(err, Equals, nil)

	t.townPrivate, err = app.TownNew(data.NewKey("testPrivate"), "Test Private Town", "Test Private Town",
		t.moderator, 100, 70, true)
	c.Assert(err, Equals, nil)

	//posts submitted to both town1 and town2
	t.postPub, err = app.PostNew("test published post", "test content", "buysell", app.PostFormatStandard, t.user,
		[]data.Key{t.town1.Key, t.town2.Key}, nil, data.EmptyUUID, true, true, false)
	c.Assert(err, Equals, nil)

	t.postDraft, err = app.PostNew("test draft post", "test content", "buysell", app.PostFormatStandard, t.user,
		[]data.Key{t.town1.Key, t.town2.Key}, nil, data.EmptyUUID, true, true, true)
	c.Assert(err, Equals, nil)

	t.postClosed, err = app.PostNew("test closed post", "test content", "buysell", app.PostFormatStandard, t.user,
		[]data.Key{t.town1.Key, t.town2.Key}, nil, data.EmptyUUID, true, true, false)

	c.Assert(err, Equals, nil)
	c.Assert(t.postClosed.Close(t.user), Equals, nil)
	c.Assert(t.postClosed.Update(), Equals, nil)

	t.post1, err = app.PostNew("test post 1", "test content", "buysell", app.PostFormatStandard, t.user,
		[]data.Key{t.town1.Key, t.town2.Key}, nil, data.EmptyUUID, true, true, false)
	c.Assert(err, Equals, nil)
	t.post2, err = app.PostNew("test post 2", "test content", "buysell", app.PostFormatStandard, t.user,
		[]data.Key{t.town1.Key, t.town2.Key}, nil, data.EmptyUUID, true, true, false)
	c.Assert(err, Equals, nil)
	t.post3, err = app.PostNew("test post 3", "test content", "buysell", app.PostFormatStandard, t.user,
		[]data.Key{t.town1.Key, t.town2.Key}, nil, data.EmptyUUID, true, true, false)
	c.Assert(err, Equals, nil)

}

func (t *testData) teardown(c *C) {
	//users
	t.deleteUser(c, t.user)
	t.deleteUser(c, t.other)
	t.deleteUser(c, t.moderator)

	//towns
	t.deleteTown(c, t.town1)
	t.deleteTown(c, t.town2)
	t.deleteTown(c, t.townPrivate)

	//posts
	t.deletePost(c, t.postPub)
	t.deletePost(c, t.postDraft)
	t.deletePost(c, t.postClosed)
	t.deletePost(c, t.post1)
	t.deletePost(c, t.post2)
	t.deletePost(c, t.post3)

	c.Assert(t.client.cache().FlushAll(), Equals, nil)
	c.Assert(t.client.cache().DeleteAll(), Equals, nil)
}

func (t *testData) deleteUser(c *C, user *app.User) {
	c.Assert(user, Not(Equals), nil)
	deleteUserResult, err := rt.DB(data.DefaultConfig().DB.Database).Table("user").
		Get(user.Username).Delete().RunWrite(t.client.database())
	c.Assert(err, Equals, nil)
	c.Assert(deleteUserResult.Errors, Equals, 0)
}

func (t *testData) deleteTown(c *C, town *app.Town) {
	c.Assert(town, Not(Equals), nil)
	deleteTownResult, err := rt.DB(data.DefaultConfig().DB.Database).Table("town").
		Get(town.Key).Delete().RunWrite(t.client.database())
	c.Assert(err, Equals, nil)
	c.Assert(deleteTownResult.Errors, Equals, 0)

	t.client.search().Delete().Index("townsourced").Type("town").Id(string(town.Key)).Do()
}

func (t *testData) deletePost(c *C, post *app.Post) {
	c.Assert(post, Not(Equals), nil)
	deletePostResult, err := rt.DB(data.DefaultConfig().DB.Database).Table("post").
		Get(post.Key).Delete().RunWrite(t.client.database())

	c.Assert(err, Equals, nil)
	c.Assert(deletePostResult.Errors, Equals, 0)

	t.client.search().Delete().Index("townsourced").Type("town").Id(string(post.Key)).Do()
}

func (t *testData) addImage(c *C, owner *app.User) *app.Image {
	var buf bytes.Buffer

	rgba := image.NewRGBA(image.Rect(0, 0, 1000, 1000))

	c.Assert(jpeg.Encode(&buf, rgba, nil), Equals, nil)

	img, err := app.ImageNew(owner, "image/jpeg", ioutil.NopCloser(&buf))
	c.Assert(err, Equals, nil)

	return img
}

func (t *testData) deleteImage(c *C, image *app.Image) {
	c.Assert(data.ImageDelete(image.Key), Equals, nil)
}

func (t *testData) deleteComment(c *C, comment *app.Comment) {
	c.Assert(comment, Not(Equals), nil)
	deleteResult, err := rt.DB(data.DefaultConfig().DB.Database).Table("comment").
		Get(comment.Key).Delete().RunWrite(t.client.database())

	c.Assert(err, Equals, nil)
	c.Assert(deleteResult.Errors, Equals, 0)
}

// gocheck hook
func Test(t *testing.T) {
	setUpGlobal(t)
	TestingT(t)
	tearDownGlobal(t)
}

// setups up the everything for all the application test suites
func setUpGlobal(t *testing.T) {
	dbCfg := data.DefaultConfig()
	dbCfg.Cache.Addresses = []string{"127.0.0.1:11211"}
	dbCfg.DB.Address = "127.0.0.1:28015"
	dbCfg.Search.Addresses = []string{"http://127.0.0.1:9200"}

	err := data.Init(dbCfg)
	if err != nil {
		t.Fatal(err)
	}

	//check if database is empty
	// if not abort all tests to ensure this doesn't accidentally get run in a live environment
	// town counts
	townCount, err := data.TownAllCount()
	if err != nil {
		t.Fatal(err)
	}

	if townCount != 0 {
		t.Fatalf("Town table is not empty! It has %d row(s)!", townCount)
	}

	// post counts
	postCount, err := data.PostAllCount()
	if err != nil {
		t.Fatal(err)
	}

	if postCount != 0 {
		t.Fatalf("Post table is not empty! It has %d row(s)!", postCount)
	}

	// user counts
	userCount, err := data.UserAllCount()
	if err != nil {
		t.Fatal(err)
	}

	if userCount != 0 {
		t.Fatalf("User table is not empty! It has %d row(s)!", postCount)
	}

	// initialize app layer
	appCFG := app.DefaultConfig()
	appCFG.DevMode = true
	appCFG.TestMode = true

	// note this requires the email templates to be fully built
	err = app.Init(appCFG, "testHost", "http://localhost:8080", "../")
	if err != nil {
		t.Fatal(err)
	}

	// clear memcache
	err = data.CacheClient().FlushAll()
	if err != nil {
		t.Fatalf("Error flushing all memcache entries: %s", err)
	}

	err = data.CacheClient().DeleteAll()
	if err != nil {
		t.Fatalf("Error deleting all memcache entries: %s", err)
	}
}

func tearDownGlobal(t *testing.T) {
	// delete databases
	var dbNames []string
	crs, err := rt.DBList().Run(data.DatabaseSession())
	if err != nil {
		t.Fatalf("Error getting database list: %s", err)
	}

	err = crs.All(&dbNames)
	if err != nil {
		t.Fatalf("Error getting DB names from cursor: %s", err)
	}

	for i := range dbNames {
		if dbNames[i] != "rethinkdb" {
			_, err = rt.DBDrop(dbNames[i]).RunWrite(data.DatabaseSession())
			if err != nil {
				t.Fatalf("Error dropping database %s: %s", dbNames[i], err)
			}
		}
	}

	// delete elasticsearch indexes
	response, err := data.SearchClient().DeleteIndex(data.DefaultConfig().Search.Index.Name).Do()
	if err != nil {
		t.Fatalf("Error dropping index %s: %s", data.DefaultConfig().Search.Index.Name, err)
	}
	if !response.Acknowledged {
		t.Fatal("Elastic Search index delete not acknowleged")
	}

	err = data.CacheClient().FlushAll()
	if err != nil {
		t.Fatalf("Error flushing all memcache entries: %s", err)
	}

	// clear memcache
	err = data.CacheClient().DeleteAll()
	if err != nil {
		t.Fatalf("Error deleting all memcache entries: %s", err)
	}
}

//App Test Suite for tests that don't need their own suite for setup and teardown
type AppSuite struct {
	*dataClient
}

var _ = Suite(&AppSuite{&dataClient{}})

func (s *AppSuite) SetUpTest(c *C) {
	// clear memcache before each test
	c.Assert(s.cache().DeleteAll(), Equals, nil)
}
