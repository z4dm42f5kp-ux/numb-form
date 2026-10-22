package lookup

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// Page is what a web page offers for extraction: its visible text one
// line per block, its links, embedded schema.org JSON-LD, meta tags and
// microdata (itemprop) values.
type Page struct {
	Lines  []string
	Links  []Link
	JSONLD [][]byte
	Meta   map[string]string // name or property, lower case -> content
	Micro  map[string]string // itemprop -> first value
}

// Link is an <a href> with its text.
type Link struct {
	Href, Text string
}

// block elements start a new line of text.
var block = map[string]bool{
	"p": true, "div": true, "li": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"tr": true, "td": true, "th": true, "dd": true, "dt": true, "address": true, "section": true, "article": true,
	"footer": true, "header": true, "br": true, "hr": true, "table": true, "ul": true, "ol": true, "form": true,
	"nav": true, "main": true, "aside": true, "blockquote": true, "pre": true, "figcaption": true, "dl": true,
	"tbody": true, "thead": true, "caption": true, "label": true, "option": true, "button": true, "body": true,
}

// skipped elements hold no visible text.
var skipped = map[string]bool{"script": true, "style": true, "noscript": true, "template": true, "svg": true, "head": true, "iframe": true, "object": true}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

// nodeText returns the text inside a node, on one line.
func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
			return
		}
		if n.Type == html.ElementNode && skipped[n.Data] {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return textutil.OneLine(b.String())
}

// ParsePage parses HTML with golang.org/x/net/html, which reads broken
// markup the way browsers do.
func ParsePage(body []byte) Page {
	p := Page{Meta: map[string]string{}, Micro: map[string]string{}}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return p
	}
	var line strings.Builder
	flush := func() {
		if l := textutil.OneLine(line.String()); l != "" {
			p.Lines = append(p.Lines, l)
		}
		line.Reset()
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			line.WriteString(n.Data)
			return
		case html.ElementNode:
			switch n.Data {
			case "script":
				if strings.Contains(strings.ToLower(attr(n, "type")), "ld+json") && n.FirstChild != nil {
					p.JSONLD = append(p.JSONLD, []byte(n.FirstChild.Data))
				}
				return
			case "meta":
				k := strings.ToLower(attr(n, "property"))
				if k == "" {
					k = strings.ToLower(attr(n, "name"))
				}
				if v := textutil.OneLine(attr(n, "content")); k != "" && v != "" {
					if _, ok := p.Meta[k]; !ok {
						p.Meta[k] = v
					}
				}
			case "a":
				if h := strings.TrimSpace(attr(n, "href")); h != "" && !strings.HasPrefix(h, "#") {
					p.Links = append(p.Links, Link{Href: h, Text: nodeText(n)})
				}
			}
			if ip := attr(n, "itemprop"); ip != "" {
				v := attr(n, "content")
				if v == "" {
					v = nodeText(n)
				}
				for _, k := range strings.Fields(ip) {
					if _, ok := p.Micro[k]; !ok && v != "" {
						p.Micro[k] = textutil.OneLine(v)
					}
				}
			}
			if skipped[n.Data] && n.Data != "head" {
				return
			}
			if n.Data == "head" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && (c.Data == "meta" || c.Data == "script") {
						walk(c)
					}
				}
				return
			}
			if block[n.Data] {
				flush()
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if n.Type == html.ElementNode && block[n.Data] {
			flush()
		}
	}
	walk(doc)
	flush()
	return p
}

// HTMLText converts HTML to plain lines, one per block of text.
func HTMLText(s string) []string { return ParsePage([]byte(s)).Lines }

// metaAddress reads an address from Open Graph business meta tags or
// schema.org microdata.
func (p Page) metaAddress() string {
	get := func(keys ...string) string {
		for _, k := range keys {
			if v := p.Meta[k]; v != "" {
				return v
			}
			if v := p.Micro[k]; v != "" {
				return v
			}
		}
		return ""
	}
	street := get("business:contact_data:street_address", "og:street-address", "streetaddress", "streetAddress")
	city := get("business:contact_data:locality", "og:locality", "addressLocality")
	postal := get("business:contact_data:postal_code", "og:postal-code", "postalCode")
	country := get("business:contact_data:country_name", "og:country-name", "addressCountry")
	if street == "" || city == "" {
		return ""
	}
	parts := []string{street, strings.TrimSpace(postal + " " + city)}
	if country != "" {
		parts = append(parts, CountryName(country))
	}
	return strings.Join(parts, ", ")
}
