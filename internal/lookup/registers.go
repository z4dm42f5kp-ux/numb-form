package lookup

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/textutil"
)

var (
	viesAPI   = "https://ec.europa.eu/taxation_customs/vies/rest-api"
	brasilAPI = "https://brasilapi.com.br/api/cnpj/v1"
)

type viesResp struct {
	IsValid bool   `json:"isValid"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

func withheld(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || strings.Trim(s, "-") == ""
}

// vies checks an EU VAT number in VIES, the European Commission's
// service that queries the national tax registers. Several countries
// (Germany among them) confirm only that the number is valid.
func vies(ctx context.Context, c *Client, q Query, vat string) (Result, error) {
	var res Result
	cc, num := vat[:2], vat[2:]
	u := viesAPI + "/ms/" + url.PathEscape(cc) + "/vat/" + url.PathEscape(num)
	var r viesResp
	if err := c.getJSON(ctx, u, nil, &r); err != nil {
		return res, err
	}
	name := "EU VIES VAT " + vat
	src := Source{ID: "vies", Name: name, URL: "https://ec.europa.eu/taxation_customs/vies/#/vat-validation-result", Kind: KindGovernment}
	if !r.IsValid {
		res.Notes = append(res.Notes, "VIES: VAT number "+vat+" is not valid")
		return res, nil
	}
	if withheld(r.Name) {
		res.Notes = append(res.Notes, "VIES: VAT number "+vat+" is valid; "+cc+" does not publish the name and address")
		return res, nil
	}
	if !Mentions(normWords(r.Name), q) && !SameEntity(q, r.Name) {
		res.Notes = append(res.Notes, "VIES: VAT number "+vat+" belongs to "+textutil.OneLine(r.Name)+", not this company")
		return res, nil
	}
	src.Name += " (" + textutil.OneLine(r.Name) + ")"
	res.Sources = append(res.Sources, src)
	res.LegalNames = append(res.LegalNames, LegalName{Name: textutil.OneLine(r.Name), Source: src})
	if !withheld(r.Address) {
		var parts []string
		for _, l := range strings.Split(r.Address, "\n") {
			if l = textutil.OneLine(l); l != "" {
				parts = append(parts, l)
			}
		}
		addr := strings.Join(parts, ", ")
		iso := cc
		if iso == "EL" {
			iso = "GR" // VIES writes Greece as EL
		}
		if n := CountryName(iso); n != iso && !strings.Contains(strings.ToLower(addr), strings.ToLower(n)) {
			addr += ", " + n
		}
		res.Addresses = append(res.Addresses, Address{Text: addr, Source: src})
	}
	return res, nil
}

type cnpjResp struct {
	RazaoSocial      string `json:"razao_social"`
	TipoLogradouro   string `json:"descricao_tipo_de_logradouro"`
	Logradouro       string `json:"logradouro"`
	Numero           string `json:"numero"`
	Complemento      string `json:"complemento"`
	Bairro           string `json:"bairro"`
	CEP              string `json:"cep"`
	Municipio        string `json:"municipio"`
	UF               string `json:"uf"`
	Situacao         string `json:"descricao_situacao_cadastral"`
	NaturezaJuridica string `json:"natureza_juridica"`
	QSA              []struct {
		Nome         string `json:"nome_socio"`
		Qualificacao string `json:"qualificacao_socio"`
	} `json:"qsa"`
}

// cnpjRoles translates the Receita Federal qualifications that carry the
// power to represent the company. Plain partners ("Sócio") do not.
var cnpjRoles = []struct{ key, role string }{
	{"socio administrador", "Managing Partner"}, {"socio gerente", "Managing Partner"},
	{"administrador", "Administrator"}, {"diretor", "Director"}, {"presidente", "President"},
	{"titular", "Owner"}, {"representante", "Legal Representative"}, {"procurador", "Attorney-in-fact"},
	{"liquidante", "Liquidator"}, {"gerente", "Manager"}, {"empresario", "Owner"},
}

func cnpjRole(q string) (string, bool) {
	k := normWords(q)
	for _, r := range cnpjRoles {
		if strings.Contains(k, " "+r.key) {
			return r.role, true
		}
	}
	return "", false
}

// cnpj reads the Receita Federal company record (CNPJ) through BrasilAPI,
// a free, keyless mirror of the Receita Federal open data.
func cnpj(ctx context.Context, c *Client, q Query, num string) (Result, error) {
	var res Result
	var r cnpjResp
	if err := c.getJSON(ctx, brasilAPI+"/"+url.PathEscape(num), nil, &r); err != nil {
		return res, err
	}
	pretty := fmt.Sprintf("%s.%s.%s/%s-%s", num[:2], num[2:5], num[5:8], num[8:12], num[12:])
	if !Mentions(normWords(r.RazaoSocial), q) && !SameEntity(q, r.RazaoSocial) {
		res.Notes = append(res.Notes, "Receita Federal: CNPJ "+pretty+" belongs to "+textutil.OneLine(r.RazaoSocial)+", not this company")
		return res, nil
	}
	name := "Receita Federal CNPJ " + pretty + " (" + textutil.OneLine(r.RazaoSocial) + ", via BrasilAPI)"
	if r.Situacao != "" && !strings.EqualFold(r.Situacao, "ATIVA") {
		name += " [status " + r.Situacao + "]"
	}
	src := Source{ID: "brasil-cnpj", Name: name, URL: "https://solucoes.receita.fazenda.gov.br/servicos/cnpjreva/cnpjreva_solicitacao.asp", Kind: KindGovernment}
	res.Sources = append(res.Sources, src)
	res.LegalNames = append(res.LegalNames, LegalName{Name: textutil.OneLine(r.RazaoSocial), Source: src})
	street := strings.TrimSpace(titleCase(strings.TrimSpace(r.TipoLogradouro + " " + r.Logradouro)))
	var parts []string
	if street != "" {
		parts = append(parts, strings.TrimSpace(street+", "+r.Numero))
	}
	if r.Complemento != "" {
		parts = append(parts, titleCase(r.Complemento))
	}
	if r.Bairro != "" {
		parts = append(parts, titleCase(r.Bairro))
	}
	cep := digitsOnly(r.CEP)
	if len(cep) == 8 {
		cep = cep[:5] + "-" + cep[5:]
	}
	city := strings.TrimSpace(cep + " " + titleCase(r.Municipio))
	if r.UF != "" {
		city += " - " + strings.ToUpper(r.UF)
	}
	if city != "" {
		parts = append(parts, city)
	}
	if len(parts) > 0 {
		res.Addresses = append(res.Addresses, Address{Text: strings.Join(append(parts, "Brazil"), ", "), Source: src})
	}
	for _, s := range r.QSA {
		role, ok := cnpjRole(s.Qualificacao)
		if !ok || strings.TrimSpace(s.Nome) == "" {
			continue
		}
		n := strings.TrimSpace(s.Nome)
		if allUpper(n) {
			n = titleCase(n)
		}
		res.addPerson(Person{Name: n, Roles: []string{role}, Sources: []Source{src}})
	}
	return res, nil
}
