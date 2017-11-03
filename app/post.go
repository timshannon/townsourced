// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

const (
	// PostMaxImages is the max number of images in a post
	PostMaxImages      = 10
	postMaxTitle       = 300
	postMaxRetrieve    = 200
	postReEditDuration = 300 * time.Second
)

const (
	// PostStatusDraft not public visible yet
	PostStatusDraft = data.PostStatusDraft
	// PostStatusPublished publicly visible
	PostStatusPublished = data.PostStatusPublished
	//PostStatusClosed not searchable but link is still valid
	PostStatusClosed = data.PostStatusClosed
)

//PostFormat is the types of formats available for posts
const (
	PostFormatStandard = "standard"
	PostFormatArticle  = "article"
	PostFormatGallery  = "gallery"
	PostFormatPoster   = "poster"
)

/* PostSort sets the sort order or post search results */
const (
	PostSearchSortNone           = data.PostSearchSortNone
	PostSearchPriceSortHighToLow = data.PostSearchPriceSortHighToLow
	PostSearchPriceSortLowToHigh = data.PostSearchPriceSortLowToHigh
)

var (
	postCategories = []string{"buysell", "jobs", "event", "notice", "housing", "volunteer"}
)

var (
	//ErrPostNoTitle is the error when a post is created without a title
	ErrPostNoTitle = fail.New("Title is required for a new post")
	//ErrPostLongTitle is the error when a post is created without a title
	ErrPostLongTitle = fail.New("The Title is too long. The max is " + strconv.Itoa(postMaxTitle))
	//ErrPostNoContent is the error when a post is created without content
	ErrPostNoContent = fail.New("Post is empty.  Post content is required")
	//ErrPostBadCategory is the error when a post is created with an invalid category
	ErrPostBadCategory = fail.New("Invalid category")
	//ErrPostBadStatus is the error when a post is created with an invalid status
	ErrPostBadStatus = fail.New("Invalid status")
	//ErrPostBadFormat is the error when a post is created with an invalid format
	ErrPostBadFormat = fail.New("Invalid format")
	//ErrPostInvalidTown is the error when a post is created with an invalid town
	ErrPostInvalidTown = fail.New("One or more towns are invalid for the post")
	//ErrPostNoTown is the error when a post is created without a town
	ErrPostNoTown = fail.New("No town was specified for the post")
	//ErrPostTooManyImages is the error when a post is created with too many images
	ErrPostTooManyImages = fail.New("Too many images associated to this post.  The max is " +
		strconv.Itoa(PostMaxImages))
	//ErrPostNotOwner is the error when a post is edited by someone who isn't the creator of the post
	ErrPostNotOwner = fail.New("You cannot edit this post because you did not create it.")
	//ErrPostNotFound is an error when a specific post cannot be found
	ErrPostNotFound = fail.New("A Post cannot be found with the key specified")
	//ErrPostNotDraft is an error when trying to update a post that isn't in a draft status
	ErrPostNotDraft = fail.New("You cannot update a post that isn't in draft status")
	//ErrPostNotClosed is an error when trying to reopen a post that isn't in a closed status
	ErrPostNotClosed = fail.New("You cannot ReOpen a post that isn't in closed status")
	//ErrPostNotPublished is an error when trying to unpublish a post that isn't in a published status
	ErrPostNotPublished = fail.New("You cannot Unpublish a post that isn't in a published status")
	//ErrPostNoFeature is an error when no featured image is present
	ErrPostNoFeature = fail.New("A post must contain a featured image.  Please choose one.")
	//ErrPostEditExpired is when the edit period has passed for a published post
	ErrPostEditExpired = fail.New(fmt.Sprintf("A post can only be unpublished within %v after it's been published.",
		postReEditDuration))
)

var (
	rxHash    = regexp.MustCompile("\\B#{1}[a-zA-Z0-9-_]+")
	rxMention = regexp.MustCompile("\\B@{1}[a-zA-Z0-9-]+")
	rxPrice   = regexp.MustCompile("\\B\\$[0-9](,*[0-9]*)*(\\.*[0-9]){0,2}")
)

// Post is a townsourced post, it can exist in multiple towns as once
type Post struct {
	Key             data.UUID           `json:"key,omitempty" gorethink:",omitempty"`
	Title           string              `json:"title,omitempty" gorethink:",omitempty"`
	Content         string              `json:"content,omitempty" gorethink:",omitempty"`
	Category        string              `json:"category,omitempty" gorethink:",omitempty"`
	TownKeys        []data.Key          `json:"townKeys,omitempty"`
	Images          []data.UUID         `json:"images,omitempty"`
	FeaturedImage   data.UUID           `json:"featuredImage,omitempty" gorethink:",omitempty"`
	Status          string              `json:"status,omitempty" gorethink:",omitempty"`
	Format          string              `json:"format,omitempty" gorethink:",omitempty"`
	Creator         data.Key            `json:"creator,omitempty" gorethink:",omitempty"`
	Moderation      []Moderated         `json:"moderation,omitempty"`
	HashTags        []data.Key          `json:"hashTags,omitempty"`
	Prices          []float64           `json:"prices,omitempty"`
	Reported        map[data.Key]string `json:"reported,omitempty" gorethink:",omitempty"`
	AllowComments   bool                `json:"allowComments,omitempty"`
	NotifyOnComment bool                `json:"notifyOnComment,omitempty"`
	Published       time.Time           `json:"published,omitempty" gorethink:",omitempty"`
	StatusLine      string              `json:"statusLine,omitempty" gorethink:","`

	data.Version
	creatorUser *User

	towns []*Town
}

// Moderated contains which moderators have moderated this post and their reason
type Moderated struct {
	Town   data.Key `json:"town,omitempty" gorethink:",omitempty"`
	Who    data.Key `json:"who,omitempty" gorethink:",omitempty"`
	Reason string   `json:"reason,omitempty" gorethink:",omitempty"`
}

// PostGet retrieves a specific post
func PostGet(key data.UUID) (*Post, error) {
	p := &Post{}
	err := data.PostGet(p, key)
	if err == data.ErrNotFound {
		return nil, ErrPostNotFound
	}

	if err != nil {
		return nil, err
	}

	return p, nil
}

// PostGetByTowns returns the list of published posts by passed in town keys
// if category is blank, then returns all categories
func PostGetByTowns(who *User, townKeys []data.Key, category string, since time.Time, limit int,
	showModerated bool) ([]Post, error) {
	var posts []Post

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(postMaxRetrieve)))

	if len(townKeys) == 0 {
		return posts, nil
	}

	if category != "" && !isPostCategory(category) {
		return nil, fail.New("Invalid category", category)
	}

	var towns []*Town
	err := data.Towns(&towns, townKeys...)
	if err != nil {
		return nil, err
	}

	for i := range towns {
		if !towns[i].CanSearch(who) {
			return nil, fail.NewFromErr(fmt.Errorf("The town %s is private and you are not currently a member.",
				towns[i].Name))
		}

		if showModerated && !towns[i].mod(who).active() {
			return nil, fail.NewFromErr(fmt.Errorf("You do not have access to view moderated posts in %s",
				towns[i].Name))
		}
	}

	err = data.PostGetByTowns(&posts, townKeys, category, since, limit, showModerated)

	if err == data.ErrNotFound {
		return posts, nil
	}
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// PostGetByLocation retrieves posts based on the passed in location information
func PostGetByLocation(longitude, latitude, milesDistant float64, category string, since time.Time,
	limit int) ([]Post, []Town, error) {
	var posts []Post

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(postMaxRetrieve)))

	if category != "" && !isPostCategory(category) {
		return nil, nil, fail.New("Invalid category", category)
	}

	// get the 5 closest towns
	// all towns will already be public
	towns, err := TownSearchDistance(longitude, latitude, milesDistant, 0, 5)
	if err != nil {
		return nil, nil, err
	}

	if len(towns) == 0 {
		return nil, nil, nil
	}

	townKeys := make([]data.Key, len(towns))

	for i := range townKeys {
		townKeys[i] = towns[i].Key
	}

	err = data.PostGetByTowns(&posts, townKeys, category, since, limit, false)

	if err == data.ErrNotFound {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	return posts, towns, nil
}

// PostSearch searches for post in the passed in towns
func PostSearch(who *User, searchText string, tags []string, towns []Town, category string,
	from, limit int, postSort string, minPrice, maxPrice float64, showModerated bool) ([]Post, error) {
	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(postMaxRetrieve)))
	from = int(math.Max(0, float64(from)))

	var posts []Post

	if len(towns) == 0 {
		return posts, nil
	}
	if category != "" && !isPostCategory(category) {
		return nil, fail.New("Invalid category", category)
	}

	if minPrice > -1 && maxPrice > -1 {
		if minPrice > maxPrice {
			minPrice, maxPrice = maxPrice, minPrice
		}
	}

	townKeys := make([]data.Key, len(towns))

	for i := range towns {
		if !towns[i].CanSearch(who) {
			return nil, fail.NewFromErr(fmt.Errorf("This town %s is private and you are not currently a member.",
				towns[i].Name))
		}

		if showModerated && !towns[i].mod(who).active() {
			return nil, fail.NewFromErr(fmt.Errorf("You do not have access to view moderated posts in %s",
				towns[i].Name))
		}

		townKeys[i] = towns[i].Key
	}

	result, err := data.PostSearch(searchText, tags, townKeys, category, from, limit, postSort, minPrice, maxPrice,
		showModerated)

	if err == data.ErrNotFound {
		return posts, nil
	}

	if err != nil {
		return nil, err
	}

	posts = make([]Post, result.Count())

	for i := range posts {
		err = result.Next(&posts[i])
		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

// PostNew creates a new Post
func PostNew(title, content, category, format string, creator *User, towns []data.Key, images []data.UUID, featuredImage data.UUID,
	allowComments, notifyOnComment, draft bool) (*Post, error) {
	post := &Post{
		Title:           title,
		Content:         content,
		Category:        category,
		Format:          format,
		TownKeys:        towns,
		Images:          images,
		FeaturedImage:   featuredImage,
		Creator:         creator.Username,
		creatorUser:     creator,
		AllowComments:   allowComments,
		NotifyOnComment: notifyOnComment,
	}

	if !allowComments {
		post.NotifyOnComment = false
	}

	if draft {
		post.Status = PostStatusDraft
	} else {
		post.Status = PostStatusPublished
	}

	err := post.validate()
	if err != nil {
		return nil, err
	}

	post.Rev()

	if !draft && len(post.Moderation) != 0 {
		// New post has been automoderated in at least one town
		// if they are publishing this, auto-set to draft so they have a
		// chance to fix the post before publishing
		post.Status = PostStatusDraft
	}

	key, err := data.PostInsert(post)
	if err != nil {
		return nil, err
	}
	post.Key = key

	if post.Status != PostStatusDraft {
		err = post.publish()
		if err != nil {
			return nil, err
		}
		// double write needed as post key is required
		// for notifications and search indexing, I may reconsider this
		// if I generate my own keys
		err = post.Update()
		if err != nil {
			return nil, err
		}
	}

	return post, nil
}

// validate a post
func (p *Post) validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return ErrPostNoTitle
	}

	if len(p.Title) > postMaxTitle {
		return ErrPostLongTitle
	}

	if strings.TrimSpace(p.Content) == "" {
		return ErrPostNoContent
	}

	err := p.checkCategory()
	if err != nil {
		return err
	}

	if p.Creator == data.EmptyKey {
		return fail.New("Post has no creator set", p)
	}

	err = postStatusCheck(p.Status)
	if err != nil {
		return err
	}

	err = p.checkFormat()
	if err != nil {
		return err
	}

	err = p.checkTowns()
	if err != nil {
		return err
	}
	err = p.checkImages()
	if err != nil {
		return err
	}

	return nil
}

func postStatusCheck(status string) error {
	if status != PostStatusClosed && status != PostStatusDraft && status != PostStatusPublished {
		return ErrPostBadStatus
	}
	return nil
}

func (p *Post) checkImages() error {
	if len(p.Images) > PostMaxImages {
		return ErrPostTooManyImages
	}

	if len(p.Images) == 0 {
		return nil
	}

	if p.FeaturedImage == data.EmptyUUID {
		if len(p.Images) > 1 {
			return ErrPostNoFeature
		}
		p.FeaturedImage = p.Images[0]
	}

	//Mark all images as in use
	// if a bad image key is passed in, then the update will fail, and the post
	// will fail validation
	// bad image keys  mixed with good ones will result in orphaned images in the DB
	// not that big of a deal, and can be cleaned up later with offline processing
	// i.e. find all images not in posts, or profiles, or towns
	featureFound := false
	for i := range p.Images {
		err := imageSetUsed(p.Images[i])
		if err != nil {
			return err
		}
		if p.Images[i] == p.FeaturedImage {
			featureFound = true
		}
	}

	if !featureFound {
		return ErrPostNoFeature
	}

	return nil
}

func (p *Post) checkTowns() error {
	if len(p.TownKeys) == 0 {
		if p.Status == PostStatusDraft {
			// towns aren't required until publishing
			return nil
		}

		return ErrPostNoTown
	}

	p.towns = nil // reset cached town list so it can be checked

	towns, err := p.Towns()
	if err == data.ErrNotFound {
		return ErrPostInvalidTown
	}

	if len(towns) != len(p.TownKeys) {
		return ErrPostInvalidTown
	}

	for i := range towns {
		reason, err := towns[i].autoModerate(p)
		if err != nil {
			return err
		}
		if reason != "" {
			err = p.addModeration(towns[i], nil, reason)
			if err != nil {
				return err
			}
		} else {
			err = p.removeModeration(towns[i])
			if err != nil {
				return err
			}
		}

		creator, err := p.creator()
		if err != nil {
			return err
		}

		if towns[i].Key == AnnouncementTown && !creator.Admin {
			return fail.New("You cannot post to the Announcements town")
		}

		if towns[i].Private && !towns[i].Invited(creator) {
			return fail.NewFromErr(fmt.Errorf("You cannot post to the town %s (%s) because it is private"+
				" and you currently do not have an invite", towns[i].Name, towns[i].Key))
		}
	}

	return nil
}

func (p *Post) checkFormat() error {
	if p.Format != PostFormatStandard && p.Format != PostFormatArticle && p.Format != PostFormatGallery &&
		p.Format != PostFormatPoster {
		return ErrPostBadFormat
	}
	return nil
}

func isPostCategory(category string) bool {
	for i := range postCategories {
		if postCategories[i] == category {
			return true
		}
	}
	return false
}

func (p *Post) checkCategory() error {
	if p.Category == "" {
		if p.Status == PostStatusDraft {
			// category isn't required until publishing
			return nil
		}
		return fail.New("You must specify a category")
	}

	if !isPostCategory(p.Category) {
		return ErrPostBadCategory
	}
	return nil
}

func (p *Post) parseHashTags() {
	hashTags := rxHash.FindAllString(p.Content, -1)
	p.HashTags = make([]data.Key, len(hashTags))
	for i := range p.HashTags {
		p.HashTags[i] = data.NewKey(hashTags[i][1:]) // drop leading #
	}
}

func (p *Post) parsePrices() {
	prices := rxPrice.FindAllString(p.Content, -1)
	prices = append(prices, rxPrice.FindAllString(p.Title, -1)...)

	p.Prices = make([]float64, 0) // reset price list to empty
	for i := range prices {
		price, err := strconv.ParseFloat(strings.Replace(prices[i][1:], ",", "", -1), 64) // drop leading $
		if err == nil {
			p.Prices = append(p.Prices, price)
		}
	}
}

func (p *Post) sendMentions() error {
	users := rxMention.FindAllString(p.Content, -1)
	for i := range users {
		username := data.NewKey(users[i][1:]) // drop leading @
		if username == p.Creator {
			//don't notify self of mention
			continue
		}
		u, err := UserGet(username)
		if err == ErrUserNotFound {
			continue
		}
		if err != nil {
			return err
		}
		err = u.mentionPost(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Post) checkCreator(who *User) error {
	if who == nil {
		return ErrPostNotOwner
	}
	if who.Username != p.Creator {
		return ErrPostNotOwner
	}
	if p.creatorUser == nil {
		p.creatorUser = who
	}
	return nil
}

// Update updates the post
func (p *Post) Update() error {
	err := data.PostUpdate(p, p.Key)
	if err != nil {
		return err
	}

	if p.Status == PostStatusPublished {
		err := data.PostIndex(p, p.Key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Post) creator() (*User, error) {
	if p.creatorUser != nil {
		return p.creatorUser, nil
	}

	return UserGet(p.Creator)
}

// Moderate puts a post into moderation for a specific town, and makes it not publically
// visible for the given town
func (p *Post) Moderate(town *Town, who *User, reason string) error {
	if !town.IsMod(who) {
		return ErrTownNotMod
	}

	err := p.addModeration(town, who, reason)
	if err != nil {
		return err
	}

	sub, msg, err := messages.use("msgPostModerated").Execute(struct {
		Post   *Post
		Town   *Town
		Reason string
	}{
		Post:   p,
		Town:   town,
		Reason: reason,
	})
	if err != nil {
		return err
	}

	creator, err := UserGet(p.Creator)
	if err != nil {
		return err
	}

	err = who.SendMessage(creator, sub, msg)
	if err != nil {
		return err
	}
	return nil
}

func (p *Post) addModeration(town *Town, who *User, reason string) error {
	if p.isModerated(town) {
		return nil
	}

	if len(strings.TrimSpace(reason)) == 0 {
		return fail.New("A reason is required when moderating a post")
	}

	username := data.EmptyKey
	if who != nil {
		username = who.Username
	}

	newMod := Moderated{
		Town:   town.Key,
		Who:    username,
		Reason: reason,
	}

	for i := range p.Moderation {
		if p.Moderation[i].Town == town.Key {
			p.Moderation[i] = newMod
			return nil
		}
	}

	p.Moderation = append(p.Moderation, newMod)
	return nil
}

// RemoveModeration removes moderation for a specific town, i.e. makes it publically
// visible again
func (p *Post) RemoveModeration(town *Town, who *User) error {
	if !town.IsMod(who) {
		return ErrTownNotMod
	}

	return p.removeModeration(town)
}

func (p *Post) removeModeration(town *Town) error {
	if !p.isModerated(town) {
		//nothing to do
		return nil
	}

	for i := range p.Moderation {
		if p.Moderation[i].Town == town.Key {
			p.Moderation = append(p.Moderation[:i], p.Moderation[i+1:]...)
			return nil
		}
	}

	return nil
}

// Publish publishes the post to make it visible publically
func (p *Post) Publish(who *User) error {
	err := p.CanEdit(who)
	if err != nil {
		return nil
	}

	return p.publish()
}

// Note: this function is called by PostNew as well as after a post updated
func (p *Post) publish() error {
	p.Status = PostStatusPublished
	p.Published = time.Now()

	p.parseHashTags()
	p.parsePrices()

	return p.sendMentions()
}

// Unpublish puts a post back into draft status so it can be edited.
// This can only be done within a certain amount of time
func (p *Post) Unpublish(who *User) error {
	err := p.checkCreator(who)
	if err != nil {
		return nil
	}

	if p.Status != PostStatusPublished {
		return ErrPostNotPublished
	}

	if time.Now().After(p.Published.Add(postReEditDuration)) {
		return ErrPostEditExpired
	}

	err = data.PostRemoveIndex(p.Key)
	if err != nil {
		return err
	}

	p.Status = PostStatusDraft
	return nil

}

// Close closes a post to hide it from searches and from the main listing
// but the post is still accessible from a direct link
func (p *Post) Close(who *User) error {
	err := p.checkCreator(who)
	if err != nil {
		return nil
	}

	if p.Status == PostStatusClosed {
		return nil
	}

	if p.Status == PostStatusPublished {
		err = data.PostRemoveIndex(p.Key)
		if err != nil {
			return err
		}
	}
	p.Status = PostStatusClosed
	return nil
}

// ReOpen reopens a post that has been closed
func (p *Post) ReOpen(who *User) error {
	err := p.checkCreator(who)
	if err != nil {
		return nil
	}

	if p.Status != PostStatusClosed {
		return ErrPostNotClosed
	}

	p.Status = PostStatusPublished
	return nil
}

func (p *Post) isModerated(t *Town) bool {
	for i := range p.Moderation {
		if p.Moderation[i].Town == t.Key {
			return true
		}
	}

	return false
}

// Etag returns an appropriate string to use as an HTTP etag
func (p *Post) Etag() string {
	return p.Version.Ver()
}

// SetVer  prepares the post data for an update
// based on the passed in vertag, if the vertag doesn't
// match the current record, then the update won't complete
func (p *Post) SetVer(verTag string) {
	p.VerTag = verTag
}

// CanEdit returns nil if the passed in user can edit
// otherwise it returns an error showing why the user can't edit
func (p *Post) CanEdit(who *User) error {
	if p.Status != PostStatusDraft {
		return ErrPostNotDraft
	}
	err := p.checkCreator(who)
	if err != nil {
		return err
	}

	return nil
}

// SetTitle sets the title on a draft version of a post
func (p *Post) SetTitle(who *User, title string) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	if strings.TrimSpace(p.Title) == "" {
		return ErrPostNoTitle
	}
	p.Title = title
	return nil
}

// SetContent sets the content on a draft version of a post
func (p *Post) SetContent(who *User, content string) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	if strings.TrimSpace(p.Content) == "" {
		return ErrPostNoContent
	}
	p.Content = content
	// re-run town moderation if content changed
	return p.checkTowns()
}

// SetCategory sets the category on a draft version of a post
func (p *Post) SetCategory(who *User, category string) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	p.Category = category
	return p.checkCategory()
}

// SetFormat sets the format on a draft version of a post
func (p *Post) SetFormat(who *User, format string) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	p.Format = format
	return p.checkFormat()
}

// SetAllowComments sets the allow comments flag on a draft version of a post
func (p *Post) SetAllowComments(who *User, allowComments bool) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	p.AllowComments = allowComments
	if !allowComments {
		p.NotifyOnComment = false
	}
	return nil
}

// SetNotifyOnComment sets the notify on comment flag on a post
// can be done on already published posts
func (p *Post) SetNotifyOnComment(who *User, notifyOnComment bool) error {
	err := p.checkCreator(who)
	if err != nil {
		return err
	}

	p.NotifyOnComment = notifyOnComment

	return nil
}

// SetTowns sets the towns a draft post is associated to
func (p *Post) SetTowns(who *User, towns []data.Key) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	// clear existing town auto moderations so they can be re-run
	nonAuto := p.Moderation[:0]
	for i := range p.Moderation {
		if p.Moderation[i].Who != data.EmptyKey {
			nonAuto = append(nonAuto, p.Moderation[i])
		}
	}

	p.Moderation = nonAuto

	p.TownKeys = towns

	return p.checkTowns()
}

// SetImages updates a post's set of images
func (p *Post) SetImages(who *User, images []data.UUID, featuredImage data.UUID) error {
	err := p.CanEdit(who)
	if err != nil {
		return err
	}

	//loop through existing images and delete
	// any that don't exist in the new set
	for _, old := range p.Images {
		found := false
		for _, nu := range images {
			if old == nu {
				found = true
				break
			}
		}
		if !found {
			err := data.ImageDelete(old)
			if err == data.ErrNotFound {
				continue
			}
			if err != nil {
				return err
			}
		}
	}

	p.Images = images
	p.FeaturedImage = featuredImage

	return p.checkImages()
}

// Visible is whether or not a given post is visible to the passed in user
func (p *Post) Visible(u *User) (bool, error) {
	if p.Status == PostStatusDraft {
		if u == nil || p.Creator != u.Username {
			return false, nil
		}
	}

	if u != nil && u.Username == p.Creator {
		//Post is alwasy visible to it's creator
		return true, nil
	}

	moddedInAllTowns := true
	for _, tk := range p.TownKeys {
		found := false
		for i := range p.Moderation {
			if p.Moderation[i].Town == tk {
				found = true
				break
			}
		}

		if !found {
			moddedInAllTowns = false
			break
		}
	}

	//if every town it's posted to is moderated, then it's not visible to public
	if moddedInAllTowns && u == nil {
		return false, nil
	}

	//if it hasn't been modded in all towns then the post is directly accessible, although
	// it won't show up in lists under specific towns

	towns, err := p.Towns()
	if err != nil {
		return false, err
	}

	for i := range towns {
		if towns[i].CanSearch(u) {
			return true, nil
		}
	}

	// every town posted to is private and user isn't a member
	return false, nil
}

// Towns returns the towns a post is associated to
func (p *Post) Towns() ([]*Town, error) {
	if p.towns != nil {
		return p.towns, nil
	}

	var towns []*Town
	err := data.Towns(&towns, p.TownKeys...)
	if err != nil {
		return nil, err
	}

	p.towns = towns

	return towns, nil
}

// Report sends a notification to all of the moderators of the towns the post is a part of
// usually to notify them of why the post should be moderated
func (p *Post) Report(who *User, reason string) error {
	if who == nil {
		return fail.New("Only logged in users can report a post, please log in and try again")
	}

	if reason == "" {
		return fail.New("You must provide a reason why you think this post should be moderated")
	}

	if _, ok := p.Reported[who.Username]; ok {
		return fail.New("You have already reported this post.")
	}

	if p.Reported == nil {
		p.Reported = make(map[data.Key]string)
	}

	p.Reported[who.Username] = reason

	mods := make(map[data.Key]struct{})

	towns, err := p.Towns()
	if err != nil {
		return err
	}

	// send a notification to each mod of each town for this post
	// but only send one notification per mod
	for _, t := range towns {
		for m := range t.Moderators {
			if t.Moderators[m].active() {
				username := t.Moderators[m].Username
				if _, ok := mods[username]; !ok {
					sub, msg, err := messages.use("msgPostReport").Execute(struct {
						Reason string
						*Post
					}{
						Reason: reason,
						Post:   p,
					})
					mod, err := UserGet(username)
					if err != nil {
						return err
					}
					err = who.SendMessage(mod, sub, msg)
					if err != nil {
						return err
					}
					mods[username] = struct{}{} //mark as sent message

				}
			}
			//if town had no active mods, then the admins can still moderate, but we'll have to
			// build reports to find these i.e posts that have been reported more than x times
			// in towns with no active moderators and hasn't been moderated yet
		}
	}
	return nil
}

const markdownChars = "[]*>`!"

// RawContent returns the content of the post with markdown and new lines removed
// For use in description meta tags, and other areas where just the contents and not the styling is needed
func (p *Post) RawContent() string {
	//FIXME
	var raw bytes.Buffer

	for _, c := range p.Content {
		found := false
		for _, m := range markdownChars {
			if c == m {
				found = true
				break
			}
		}
		if !found {
			if c == '\n' || c == '\t' {
				_, _ = raw.WriteRune(' ')
			} else {
				_, _ = raw.WriteRune(c)
			}
		}
	}

	return raw.String()
}
