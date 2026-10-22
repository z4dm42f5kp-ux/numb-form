package lookup

import (
	"sort"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/match"
	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

// Kind ranks how authoritative a source is. Lower is better.
type Kind int

// Source kinds, most authoritative first.
const (
	KindGovernment Kind = iota // a government company register
	KindRegulatory             // the GLEIF LEI register
	KindCompany                // the company's own legal notice
	KindCommunity              // community-edited data (Wikidata)
)

func (k Kind) String() string {
	switch k {
	case KindGovernment:
		return "government register"
	case KindRegulatory:
		return "regulatory register"
	case KindCompany:
		return "company website"
	case KindCommunity:
		return "community-edited, not official"
	}
	return "unknown"
}

// Source tells where a value came from.
type Source struct {
	ID   string // the source identifier used in *_sources settings
	Name string
	URL  string
	Kind Kind
}

func (s Source) String() string { return s.Name + " (" + s.Kind.String() + ")" }

// Address is a candidate headquarters or registered address.
type Address struct {
	Text   string
	Source Source
}

// Person is a candidate authorized representative.
type Person struct {
	Name    string
	Roles   []string // English
	Sources []Source
	// Untranslated holds roles that could not be translated to English.
	Untranslated []string
}

// Best returns the most authoritative kind among the person's sources.
func (p Person) Best() Kind {
	b := KindCommunity
	for _, s := range p.Sources {
		if s.Kind < b {
			b = s.Kind
		}
	}
	return b
}

// Verified reports whether the person is backed by a register or by two
// independent kinds of source. Unverified people get the caution text.
func (p Person) Verified() bool {
	if p.Best() <= KindRegulatory {
		return true
	}
	kinds := map[Kind]bool{}
	for _, s := range p.Sources {
		kinds[s.Kind] = true
	}
	return len(kinds) >= 2
}

// Excerpt is raw text from a legal notice, shown so the user can read it.
type Excerpt struct {
	Source Source
	Text   string
}

// LegalName is the full legal name of the company as a source writes it.
type LegalName struct {
	Name   string
	Source Source
}

// Result is everything found for one company.
type Result struct {
	Addresses  []Address
	People     []Person
	FormHints  []string
	LegalNames []LegalName
	Websites   []string // suggested by registers (Wikidata)
	Website    string   // the company's own website that was read
	IDs        []ID     // official identifiers printed on its website
	Register   []RegisterNote
	Excerpts   []Excerpt
	Notes      []string
	Sources    []Source
}

// Query is what is known about a company before the lookup.
type Query struct {
	Raw       string   // the spreadsheet cell
	Name      string   // the clean name: core and legal form, no places
	Form      string   // legal form
	Base      string   // the core name
	Prefix    bool     // the legal form comes first ("Ets Zorbu")
	Places    []string // cities and countries written after the name
	Countries []string // from the places, then the legal form
	Confident bool     // the first country is more than a guess
	Searches  []string // web searches to try, in order
}

// NewQuery builds a query from the literal spreadsheet name.
func NewQuery(name string) Query { return queryFor(name) }

var formEquiv = map[string]string{
	"limited": "ltd", "ltd": "ltd",
	"incorporated": "inc", "inc": "inc",
	"corporation": "corp", "corp": "corp",
	"aktiengesellschaft": "ag", "ag": "ag",
	"gesellschaftmitbeschrankterhaftung": "gmbh", "gmbh": "gmbh",
	"sociedadanonima": "sa", "sa": "sa",
	"sociedadlimitada": "ltda", "ltda": "ltda",
	"companiaanonima": "ca", "ca": "ca",
	"publiclimitedcompany": "plc", "plc": "plc",
}

// SameEntity reports whether a register name denotes the queried company.
// The legal form must agree when both sides have one.
func SameEntity(q Query, name string) bool {
	for _, a := range textutil.MatchKeys(q.Name) {
		for _, b := range textutil.MatchKeys(name) {
			if a == b {
				return true
			}
		}
	}
	f, base := match.LegalForm(name)
	if f == "" {
		// Try long forms such as "Limited" or "Aktiengesellschaft".
		for long := range formEquiv {
			k := textutil.MatchKeys(name)[0]
			if len(long) > 3 && strings.HasSuffix(k, long) {
				f, base = long, k[:len(k)-len(long)]
				break
			}
		}
	}
	bk := textutil.MatchKeys(base)[0]
	qk := textutil.MatchKeys(q.Base)[0]
	if bk == "" || bk != qk {
		return false
	}
	if q.Form == "" || f == "" {
		return false
	}
	ef, ok1 := formEquiv[textutil.MatchKeys(f)[0]]
	eq, ok2 := formEquiv[textutil.MatchKeys(q.Form)[0]]
	if ok1 && ok2 {
		return ef == eq
	}
	return textutil.MatchKeys(f)[0] == textutil.MatchKeys(q.Form)[0]
}

func (r *Result) addPerson(p Person) {
	p.Name = textutil.OneLine(p.Name)
	if p.Name == "" {
		return
	}
	key := textutil.MatchKeys(p.Name)[0]
	for i := range r.People {
		if textutil.MatchKeys(r.People[i].Name)[0] == key {
			r.People[i].Roles = union(r.People[i].Roles, p.Roles)
			r.People[i].Untranslated = union(r.People[i].Untranslated, p.Untranslated)
			r.People[i].Sources = append(r.People[i].Sources, p.Sources...)
			return
		}
	}
	r.People = append(r.People, p)
}

func (r *Result) merge(o Result) {
	r.Addresses = append(r.Addresses, o.Addresses...)
	for _, p := range o.People {
		r.addPerson(p)
	}
	r.FormHints = union(r.FormHints, o.FormHints)
	r.LegalNames = append(r.LegalNames, o.LegalNames...)
	if r.Website == "" {
		r.Website = o.Website
	}
	r.IDs = appendIDs(r.IDs, o.IDs)
	for _, n := range o.Register {
		dup := false
		for _, m := range r.Register {
			if strings.EqualFold(m.Text, n.Text) {
				dup = true
				break
			}
		}
		if !dup {
			r.Register = append(r.Register, n)
		}
	}
	r.Websites = union(r.Websites, o.Websites)
	r.Excerpts = append(r.Excerpts, o.Excerpts...)
	r.Notes = append(r.Notes, o.Notes...)
	r.Sources = append(r.Sources, o.Sources...)
}

func union(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range append(append([]string(nil), a...), b...) {
		k := strings.ToLower(strings.TrimSpace(s))
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, strings.TrimSpace(s))
	}
	return out
}

// sortResult orders candidates by authority.
func (r *Result) sortResult() {
	sort.SliceStable(r.Addresses, func(i, j int) bool { return r.Addresses[i].Source.Kind < r.Addresses[j].Source.Kind })
	sort.SliceStable(r.People, func(i, j int) bool { return r.People[i].Best() < r.People[j].Best() })
	sort.SliceStable(r.LegalNames, func(i, j int) bool { return r.LegalNames[i].Source.Kind < r.LegalNames[j].Source.Kind })
	// Drop duplicate address texts, keeping the most authoritative.
	seen := map[string]bool{}
	var out []Address
	for _, a := range r.Addresses {
		k := textutil.MatchKeys(a.Text)[0]
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, a)
	}
	r.Addresses = out
}

// Merge adds o to r and re-sorts.
func (r *Result) Merge(o Result) {
	r.merge(o)
	r.sortResult()
}
