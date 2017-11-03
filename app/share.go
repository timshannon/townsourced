// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package app

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"git.townsourced.com/townsourced/goquery"
	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

var (
	//ErrShareURL is the error returned when the share URL could not be retrieved
	ErrShareURL = fail.New("The URL could not be retrieved or processed")
)

type shareDoc struct {
	*goquery.Document
	useragent string
}

// Share processes a share call to add a temporary post from a url or from the passed in params
// it'll generate a draft post from the passed in params and return that post
// Note this function does not save the post in the database, it only returns a temporary post object
// from the passed in variables
// if full then we'll attempt to convert the entire page to markdown in the post
func Share(who *User, strURL, title, content, useragent string, town data.Key, images []string, imageKeys []data.UUID,
	selector string) (*Post, error) {
	//if title, content or images are specified, use those

	// if not try to find them in the passed in URL
	// Try open graph tags
	// try schema.org tags
	// try specific domain parsers
	// if title, content or images can't be found via one of the above, then
	// use page title, and guessed page content

	uri, err := url.Parse(strURL)
	if err != nil {
		return nil, ErrShareURL
	}

	post := &Post{
		Title:           strings.TrimSpace(title),
		Content:         htmlStringToMarkdown(content, uri),
		Format:          PostFormatStandard,
		Images:          imageKeys,
		AllowComments:   true,
		NotifyOnComment: true,
	}

	if town != data.EmptyKey {
		post.TownKeys = []data.Key{town}
	}

	// images discovered if the URL is loaded, will be combined with the passed in images
	var discoveredImages []string

	if post.Title == "" || post.Content == "" && strings.TrimSpace(strURL) != "" {

		doc, err := loadShareURL(strURL, useragent)
		if err != nil {
			return nil, err
		}

		parser, err := doc.parser()
		if err != nil {
			return nil, err
		}

		if post.Title == "" {
			post.Title = parser.title()

		}

		if post.Content == "" {
			if selector != "" {
				post.Content = documentToMarkdown(doc.Document.Find(selector), doc.Url, "")
			} else {
				post.Content = parser.content()
			}
		}

		if len(images) == 0 {
			discoveredImages = parser.images()
			if len(discoveredImages) == 0 && selector != "" {
				all := doc.allImages()
				discoveredImages = append(discoveredImages, all...)
			}
		}
	}

	// images
	images = append(images, discoveredImages...)

	final := make([]string, 0, len(images))

	// remove duplicates
	for i := range images {
		exists := false
		for f := range final {
			if final[f] == images[i] {
				exists = true
				break
			}
		}
		if !exists {
			final = append(final, images[i])
		}
	}

	if len(final) > PostMaxImages {
		images = final[:PostMaxImages]
	} else {
		images = final
	}

	//fetch and upload images, and add to post
	for i := range images {

		lnk, err := url.Parse(images[i])
		if err != nil {
			log.Infof("Error building url for image request %s Error: %s", images[i], err)
			continue
		}

		lnk = uri.ResolveReference(lnk)

		req, err := http.NewRequest("GET", lnk.String(), nil)
		if err != nil {
			log.Infof("Error building request for image request %s Error: %s", lnk, err)
			continue
		}
		req.Header.Set("User-Agent", useragent)
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Infof("Error retrieving image from URL %s Error: %s", lnk, err)
			continue
		}

		img, err := ImageNew(who, resp.Header.Get("Content-Type"), resp.Body)
		if err != nil {
			log.Infof("Error inserting image from URL %s Error: %s", lnk, err)
			continue
		}

		post.Images = append(post.Images, img.Key)
	}

	if len(post.Images) > 0 {
		post.FeaturedImage = post.Images[0]
	}

	return post, nil
}

func loadShareURL(uri, useragent string) (*shareDoc, error) {
	doc := &shareDoc{
		useragent: useragent,
	}

	gDoc, err := doc.loadURL(uri)

	if err != nil {
		return nil, err
	}

	doc.Document = gDoc
	return doc, nil
}

func (d *shareDoc) loadURL(uri string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Infof("Error building request for share URL %s Error: %s", uri, err)
		return nil, ErrShareURL
	}
	req.Header.Set("User-Agent", d.useragent)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Infof("Error retrieving share URL %s Error: %s", uri, err)
		return nil, ErrShareURL
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Infof("Error loading response from share URL into goquery %s Error: %s", uri, err)
		return nil, ErrShareURL
	}

	return doc, nil
}

// grabs all images on the page, makes sure the default parser images (og / schema are included if present)
func (d *shareDoc) allImages() []string {
	images := (*defaultParser)(d).images()
	d.Find("img").Each(func(i int, s *goquery.Selection) {
		if img, exists := s.Attr("src"); exists {
			images = append(images, img)
			return
		}
	})

	return images
}

func htmlStringToMarkdown(content string, uri *url.URL) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(content))

	if err != nil {
		log.Infof("Error loading html string into goquery for markdown parsing: %s", err)
		return ""
	}

	return documentToMarkdown(doc.Selection, uri, "")
}

func documentToMarkdown(doc *goquery.Selection, u *url.URL, leader string) string {
	md := ""

	doc.Contents().Each(func(i int, s *goquery.Selection) {
		n := s.Nodes[0]

		switch n.Type {
		case html.TextNode:
			leftPad := " "
			rightPad := " "

			// set padding max 1 space on either side
			val := strings.TrimLeftFunc(n.Data, unicode.IsSpace)
			if (len(n.Data) - len(val)) < 1 {
				leftPad = ""
			}

			val2 := strings.TrimRightFunc(val, unicode.IsSpace)
			if (len(val) - len(val2)) < 1 {
				rightPad = ""
			}

			if val2 == "" {
				return
			}

			md += leftPad + val2 + rightPad
			if goquery.NodeName(s.Parent()) == "li" {
				md += "\n"
			}
		case html.ElementNode:
			switch goquery.NodeName(s) {
			case "li":
				if goquery.NodeName(s.Parent()) == "ol" {
					md += leader + "1. " + documentToMarkdown(s, u, leader+"\t")
				} else {
					md += leader + "* " + documentToMarkdown(s, u, leader+"\t")
				}
			case "br":
				md += "\n"
			case "h1":
				md += "\n" + leader + "# " + documentToMarkdown(s, u, leader) + "\n"
			case "h2":
				md += "\n" + leader + "## " + documentToMarkdown(s, u, leader) + "\n"
			case "h3":
				md += "\n" + leader + "### " + documentToMarkdown(s, u, leader) + "\n"
			case "p":
				md += "\n" + leader + documentToMarkdown(s, u, leader) + "\n"
			case "div":
				childVal := documentToMarkdown(s, u, leader)
				if strings.TrimSpace(childVal) == "" {
					md += leader + childVal
					return
				}

				md += "\n" + leader + childVal + "\n"
			case "hr":
				md += "\n***\n"
			case "b", "strong":
				md += leader + "**" + strings.TrimSpace(documentToMarkdown(s, u, leader)) + "** "
			case "em", "i":
				md += leader + "*" + strings.TrimSpace(documentToMarkdown(s, u, leader)) + "* "
			case "a":
				if href, ok := s.Attr("href"); ok {
					lnk, err := url.Parse(href)
					if err != nil {
						return
					}

					lnk = u.ResolveReference(lnk)

					md += leader + "[" + strings.TrimSpace(strings.Replace(s.Text(), "\n", " ", -1)) +
						"](" + lnk.String() + ")"
				}
			case "script", "head":
				return
			default:
				md += documentToMarkdown(s, u, leader)
			}
		default:
			return
		}

	})

	return md
}
