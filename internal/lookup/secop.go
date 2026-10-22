package lookup

import (
	"context"
	"net/url"
	"strings"
)

// Colombia's SECOP II supplier register on the government open data
// portal: every company that registered to sell to the state, with its
// address and legal representative. Free and keyless.
var secopAPI = "https://www.datos.gov.co/resource/qmzu-gj57.json"

type secopRecord struct {
	Nombre        string `json:"nombre"`
	NIT           string `json:"nit"`
	Direccion     string `json:"direccion"`
	Municipio     string `json:"municipio"`
	Departamento  string `json:"departamento"`
	Pais          string `json:"pais"`
	Representante string `json:"nombre_representante_legal"`
	Activa        string `json:"esta_activa"`
}

// secopQuote escapes a value for a SoQL string literal.
func secopQuote(s string) string { return strings.ReplaceAll(s, "'", "''") }

// secop finds the company in the SECOP II supplier register by its core
// name and takes its address and legal representative.
func secop(ctx context.Context, c *Client, q Query) (Result, error) {
	var res Result
	core := strings.ToUpper(strings.TrimSpace(q.Base))
	if len([]rune(core)) < 4 {
		return res, nil
	}
	v := url.Values{
		"$where": {"upper(nombre) like '%" + secopQuote(core) + "%'"},
		"$limit": {"20"},
	}
	var recs []secopRecord
	err := c.getJSON(ctx, secopAPI+"?"+v.Encode(), nil, &recs)
	tr(ctx).Ev("secop", "core", core, "n", len(recs), "err", err)
	if err != nil {
		return res, err
	}
	for _, r := range recs {
		text := normWords(r.Nombre + " " + r.Municipio + " " + r.Departamento + " colombia")
		if !Mentions(text, q) {
			tr(ctx).Ev("secop-skip", "name", r.Nombre)
			continue
		}
		name := "Colombia SECOP II supplier register, NIT " + r.NIT + " (" + r.Nombre + ")"
		if r.Activa != "" && !strings.EqualFold(r.Activa, "Si") {
			name += " [inactive]"
		}
		src := Source{ID: "secop", Name: name, URL: "https://www.datos.gov.co/d/qmzu-gj57", Kind: KindGovernment}
		res.Sources = append(res.Sources, src)
		res.LegalNames = append(res.LegalNames, LegalName{Name: r.Nombre, Source: src})
		if d := strings.TrimSpace(r.Direccion); d != "" && !strings.EqualFold(d, "No Provisto") {
			res.Addresses = append(res.Addresses, Address{Text: d + ", " + titleCase(r.Municipio) + ", " + titleCase(r.Departamento) + ", Colombia", Source: src})
		}
		if n := strings.TrimSpace(r.Representante); n != "" && !strings.EqualFold(n, "No Definido") {
			if allUpper(n) {
				n = titleCase(n)
			}
			res.addPerson(Person{Name: n, Roles: []string{"Legal Representative"}, Sources: []Source{src}})
		}
		break
	}
	return res, nil
}
