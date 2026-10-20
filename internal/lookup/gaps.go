package lookup

import "context"

// needs reports what is still missing: representatives confirmed by an
// official source (or by two kinds of source), and an address from a
// register or the company's own website.
func needs(r *Result) (reps, addr bool) {
	reps = true
	for _, p := range r.People {
		if p.Verified() {
			reps = false
			break
		}
	}
	addr = true
	for _, a := range r.Addresses {
		if a.Source.Kind <= KindCompany {
			addr = false
			break
		}
	}
	return reps, addr
}

// addressWords is what people add to a search for a company's address.
func addressWords(country string) string {
	switch country {
	case "DE", "AT", "CH":
		return "Adresse Sitz"
	case "BR", "PT":
		return "endereço sede"
	case "FR", "BE", "LU", "MA", "SN", "CI", "CM":
		return "adresse siège social"
	case "US", "GB", "IE", "ZA", "NG", "KE", "GH", "IN", "AU", "CA", "NZ", "SG":
		return "address headquarters"
	}
	return "dirección domicilio"
}

// fillGaps looks further when representatives or the address are still
// missing: directory pages already found, then one search for the
// representatives and one for the address (their snippets and the
// directory pages they find), then OpenStreetMap for an address.
func fillGaps(ctx context.Context, c *Client, cfg Config, q Query, region, country string, hits []Hit, useSearch bool, all *Result) {
	read := 0
	readDirs := func(h []Hit) {
		if !cfg.enabled("directory") || read >= 3 {
			return
		}
		jobs := directoryJobs(c, q, h, 3-read)
		read += len(jobs)
		all.merge(runJobs(ctx, jobs))
	}
	if reps, addr := needs(all); reps || addr {
		readDirs(hits)
	}
	if reps, _ := needs(all); reps && useSearch {
		var n []string
		h := search(ctx, c, `"`+q.Name+`" `+roleWords(country), region, &n)
		all.Notes = append(all.Notes, n...)
		all.merge(mineSnippets(h, q))
		readDirs(h)
	}
	if _, addr := needs(all); addr && useSearch && len(all.Addresses) == 0 {
		var n []string
		h := search(ctx, c, `"`+q.Name+`" `+addressWords(country), region, &n)
		all.Notes = append(all.Notes, n...)
		all.merge(mineSnippets(h, q))
		readDirs(h)
	}
	if len(all.Addresses) == 0 && cfg.enabled("osm") {
		r, err := osmAddress(ctx, c, q)
		if err != nil {
			r.Notes = append(r.Notes, "OpenStreetMap: "+sanitizeErr(err))
		}
		all.merge(r)
	}
}
