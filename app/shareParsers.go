// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

package app

import (
	"strings"

	"git.townsourced.com/townsourced/goquery"
	"github.com/timshannon/townsourced/fail"
)

type shareDomainParser interface {
	matchDomain(string) bool
	title() string
	content() string
	images() []string
}

func (d *shareDoc) parser() (shareDomainParser, error) {
	domain := d.Url.Host

	switch {
	case (*craigslistParser)(d).matchDomain(domain):
		return (*craigslistParser)(d), nil
	case (*ebayclassifiedsParser)(d).matchDomain(domain):
		return (*ebayclassifiedsParser)(d), nil
	case (*ebayParser)(d).matchDomain(domain):
		return (*ebayParser)(d), nil
	case (*kijijiParser)(d).matchDomain(domain):
		return (*kijijiParser)(d), nil

	case (*noParser)(d).matchDomain(domain):
		return nil, fail.New("Cannot parse from this domain", domain)
	default:
		return (*defaultParser)(d), nil
	}
}

// no parser, known domains that can't be parsed from the server side
type noParser shareDoc

func (p *noParser) matchDomain(domain string) bool {
	return strings.HasSuffix(domain, "facebook.com") ||
		strings.HasSuffix(domain, "offerup.com") ||
		strings.HasSuffix(domain, "offerupnow.com")
}
func (p *noParser) title() string {
	return ""
}

func (p *noParser) content() string {
	return ""
}

func (p *noParser) images() []string {
	return nil
}

// default / fallback parser
type defaultParser shareDoc

func (d *defaultParser) matchDomain(domain string) bool {
	return true
}

func (d *defaultParser) title() string {
	var title string

	// try opengraph
	d.Find("head > meta").Each(func(i int, s *goquery.Selection) {
		if prop, exists := s.Attr("property"); exists {
			if prop == "og:title" {
				title, exists = s.Attr("content")
				if exists {
					return
				}
			}
		}
	})

	//try schema
	if strings.TrimSpace(title) == "" {
		d.Find("meta").Each(func(i int, s *goquery.Selection) {
			if prop, exists := s.Attr("itemprop"); exists {
				if prop == "headline" || prop == "title" || prop == "name" {
					title, exists = s.Attr("content")
					if exists {
						return
					}
				}
			}
		})

	}

	// nothing? use the page title
	if strings.TrimSpace(title) == "" {
		title = d.Find("title").Text()
	}

	return title
}

func (d *defaultParser) content() string {
	content := ""

	//opengraph
	d.Find("head > meta").Each(func(i int, s *goquery.Selection) {
		if prop, exists := s.Attr("property"); exists {
			if prop == "og:description" {
				content, exists = s.Attr("content")
				if exists {
					return
				}
			}
		}
	})

	//try schema
	if strings.TrimSpace(content) == "" {
		d.Find("meta").Each(func(i int, s *goquery.Selection) {
			if prop, exists := s.Attr("itemprop"); exists {
				if prop == "description" {
					content, exists = s.Attr("content")
					if exists {
						return
					}
				}
			}
		})

	}

	content += "\n\n" + `*See [` + d.Url.Host + `](` + d.Url.String() + `) for more information*`

	return content
}

func (d *defaultParser) images() []string {
	var images []string
	d.Find("head > meta").Each(func(i int, s *goquery.Selection) {
		if prop, exists := s.Attr("property"); exists {
			if prop == "og:image" {
				img, exists := s.Attr("content")
				if exists {
					images = append(images, img)
					return
				}
			}
		}
	})

	return images
}

// craigslist parser
type craigslistParser shareDoc

func (c *craigslistParser) matchDomain(domain string) bool {
	return strings.HasSuffix(domain, "craigslist.org")
}
func (c *craigslistParser) title() string {
	title := ""
	price := ""

	selection := c.Document.Find(".postingtitletext")

	title = selection.Find("#titletextonly").Text()
	price = selection.Find(".price").Text()

	if price != "" {
		return title + " - " + price
	}

	return title
}

func (c *craigslistParser) content() string {
	return documentToMarkdown(c.Document.Find("#postingbody"), c.Url, "")
}

func (c *craigslistParser) images() []string {
	var images []string

	c.Find("#thumbs > a").Each(func(i int, s *goquery.Selection) {
		if img, exists := s.Attr("href"); exists {
			images = append(images, img)
			return
		}
	})

	if len(images) == 0 {
		// find single image
		c.Find(".gallery  img").Each(func(i int, s *goquery.Selection) {
			if img, exists := s.Attr("src"); exists {
				images = append(images, img)
				return
			}
		})
	}

	for i := range images {
		images[i] = strings.Replace(images[i], "600x450", "1200x900", 1)
	}

	return images
}

// ebayclassifieds.com parser
type ebayclassifiedsParser shareDoc

func (p *ebayclassifiedsParser) matchDomain(domain string) bool {
	return strings.HasSuffix(domain, "ebayclassifieds.com")
}
func (p *ebayclassifiedsParser) title() string {
	return p.Document.Find("#ad-title").Text()
}

func (p *ebayclassifiedsParser) content() string {
	content := ""

	price := p.Find("#ad-title > span.price").Text()
	if strings.TrimSpace(price) != "" {
		content += "**Price**" + price + "\n\n"
	}

	details := p.Document.Find("#ad-details")
	content += documentToMarkdown(details.Find(".desc-text"), p.Url, "")

	return content
}

func (p *ebayclassifiedsParser) images() []string {
	var images []string

	p.Find("#ad-image-viewer .imageNavs img").Each(func(i int, s *goquery.Selection) {
		if img, exists := s.Attr("src"); exists {
			images = append(images, img)
		}
	})

	for i := range images {
		images[i] = strings.Replace(images[i], "$_5.", "$_20.", 1)
	}

	if len(images) == 0 {
		if img, exists := p.Find("#ad-image-viewer-hero img").Attr("src"); exists {
			images = append(images, img)
		}
	}

	return images
}

// ebay.com parser
type ebayParser shareDoc

func (p *ebayParser) matchDomain(domain string) bool {
	return strings.HasSuffix(domain, "ebay.com")
}
func (p *ebayParser) title() string {
	return (*defaultParser)(p).title()
}

func (p *ebayParser) content() string {
	if iframe, exists := p.Find("#desc_ifr").Attr("src"); exists {
		doc, err := (*shareDoc)(p).loadURL(iframe)
		if err != nil {
			return ""
		}

		return documentToMarkdown(doc.Selection, p.Url, "")
	}

	return ""
}

func (p *ebayParser) images() []string {
	var images []string

	p.Find("#vi_main_img_fs .tdThumb img").Each(func(i int, s *goquery.Selection) {
		if img, exists := s.Attr("src"); exists {
			images = append(images, img)
		}
	})

	for i := range images {
		images[i] = strings.Replace(images[i], "l64.", "l1600.", 1)
	}

	if len(images) == 0 {
		return (*defaultParser)(p).images()
	}

	return images
}

// kijiji.ca parser
type kijijiParser shareDoc

func (p *kijijiParser) matchDomain(domain string) bool {
	return strings.HasSuffix(domain, "kijiji.ca")
}

func (p *kijijiParser) title() string {
	title := p.Find("#ItemDetails header h1").Text()
	if strings.TrimSpace(title) == "" {
		title = p.Find("[itemprop=name]").Text()
	}

	if strings.TrimSpace(title) == "" {
		title = (*defaultParser)(p).title()
	}

	return title
}

func (p *kijijiParser) content() string {
	return documentToMarkdown(p.Find("#UserContent"), p.Url, "")
}

func (p *kijijiParser) images() []string {
	var images []string

	p.Find("#ImageSlideshow .large-image img").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("button") {
			return
		}

		if img, exists := s.Attr("src"); exists {
			images = append(images, img)
		}
	})

	if len(images) == 0 {
		p.Find("#ShownImage img").Each(func(i int, s *goquery.Selection) {
			if s.HasClass("button") {
				return
			}

			if img, exists := s.Attr("src"); exists {
				images = append(images, img)
			}
		})

	}

	for i := range images {
		images[i] = strings.Replace(images[i], "$_35.", "$_27.", 1)
	}

	if len(images) == 0 {
		return (*defaultParser)(p).images()
	}

	return images
}
