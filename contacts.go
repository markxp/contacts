package contacts

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ScopePeopleAPI is the Google OAuth2 Authorization scopes.
// The legacy scope https://www.google.com/m8/feeds is an alias
// of the https://www.googleapis.com/auth/contacts scope.
var ScopePeopleAPI = []string{
	"https://www.googleapis.com/auth/contacts",
	"https://www.googleapis.com/auth/contacts.other.readonly",
	"https://www.googleapis.com/auth/directory.readonly",
}

// baseURL is the base endpoint of Domain Shared Contacts
const baseURL = "https://www.google.com/m8/feeds"

// hTransport adds custom header that Domain Shared Contacts API need.
type trapnsport struct{ base http.RoundTripper }

func (rt *trapnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("GData-Version", "3.0")
	switch req.Method {
	case http.MethodPost, http.MethodPut:
		req.Header.Set("Content-Type", "application/atom+xml")
	default:
	}

	return rt.base.RoundTrip(req)
}

// Service talks to Domain Shared Contact API.
type Service interface {
	// CreateContact creates a contact. Its return value is the saved version at server side.
	CreateContact(ctx context.Context, p *ContactKind) (*ContactKind, error)

	// GetContact retreives a contact data. If etag is provided, it uses conditional retreives (returns nil, nil for HTTP 304 NOT MODIFIED)
	GetContact(ctx context.Context, id, projection, etag string) (*ContactKind, error)

	// ListContacts retreives contacts. If the feed etag is provided, it uses conditional retreives (returns nil, nil for HTTP 304 NOT MODIFIED)
	ListContacts(ctx context.Context, projection, feedEtag string, queries ...func(url.Values)) ([]*ContactKind, *QueryStatus, error)

	// UpdateContact changes a contact data. If etag is provided, only the version is met will run updates.
	// If etag equals to '*', it overwrites the current version.
	UpdateContact(ctx context.Context, id, etag string, p *ContactKind) (*ContactKind, error)

	// DeleteContact deletes a contact. If etag is provided, only the version is met will be deleted.
	// If etag equals to '*', it overwrites the current version.
	DeleteContact(ctx context.Context, id, etag string) error
}

// In the Domain Shared Contacts API, several elements are slightly more restrictive than the contact kind.
// For the following elements, you supply either a rel attribute or a label attribute, but not both:
//
// gd:email
// gd:im
// gd:organization
// gd:phoneNumber
// gd:structuredPostalAddress
//
// When you create or update a shared contact, if you supply both rel and label, or neither,
// for any of those elements, then the server rejects the entry.

// ContactKind is the contact API used, atom-xml based structure.
// It represents a person's contact data. Such as address, name, email, etc...
type ContactKind struct {
	Name                    GDName
	Email                   []GDEmail
	PhoneNumber             []GDPhoneNumber
	StructuredPostalAddress []GDStructuredPostalAddress
	IM                      []GDIM
	ExtendedProperty        map[string]string

	deleted   bool
	editLink  string
	photoLink string
	selfLink  string
	id        string
	updated   time.Time
	content   string
	etag      string
}

// GetEditLink returns the edit link of the contact entry.
func (c ContactKind) GetEditLink() string { return c.editLink }

// GetPhotoLink returns the photo link of the contact entry.
func (c ContactKind) GetPhotoLink() string { return c.photoLink }

// GetID returns the ID of the contact entry.
func (c ContactKind) GetID() string {
	idx := strings.LastIndex(c.id, "/")
	return c.id[idx+1:]
}

// GetUpdated returns the last updated time of the contact entry.
func (c ContactKind) GetUpdated() time.Time { return c.updated }

// GetEtag returns the etag of the contact entry.
func (c ContactKind) GetEtag() string { return c.etag }

// Clone clones the contact.
func (c ContactKind) Clone() ContactKind {
	ret := ContactKind{
		Name:                    c.Name,
		Email:                   make([]GDEmail, len(c.Email)),
		PhoneNumber:             make([]GDPhoneNumber, len(c.PhoneNumber)),
		StructuredPostalAddress: make([]GDStructuredPostalAddress, len(c.StructuredPostalAddress)),
		IM:                      make([]GDIM, 0, len(c.IM)),
		ExtendedProperty:        make(map[string]string),
		deleted:                 c.deleted,
		editLink:                c.editLink,
		photoLink:               c.photoLink,
		selfLink:                c.selfLink,
		id:                      c.id,
		updated:                 c.updated,
		content:                 c.content,
		etag:                    c.etag,
	}
	for _, v := range c.Email {
		ret.Email = append(ret.Email, v)
	}
	for _, v := range c.PhoneNumber {
		ret.PhoneNumber = append(ret.PhoneNumber, v)
	}
	for _, v := range c.StructuredPostalAddress {
		ret.StructuredPostalAddress = append(ret.StructuredPostalAddress, v)
	}
	for _, v := range c.IM {
		ret.IM = append(ret.IM, v)
	}
	for k, v := range c.ExtendedProperty {
		ret.ExtendedProperty[k] = v
	}

	return ret
}

var endpointBaseURL = "https://www.google.com/m8/feeds/contacts/%s"

type service struct {
	base       *http.Client
	endpoint   string
	projection string
}

// NewService returns a Service that manipulate Domain Shread Contact API.
func NewService(client *http.Client, domain, defaultProjection string) (Service, error) {
	client.Transport = &trapnsport{base: client.Transport}
	return &service{client, fmt.Sprintf(endpointBaseURL, domain), setDefaultProjection(defaultProjection)}, nil
}

func setDefaultProjection(p string) string {
	if p == "" {
		return "full"
	}
	return p
}

// getProjection returns request-scoped projection value.
// If request-scoped projection is not set, use default projection value.
func (s service) getPojection(p string) string {
	if p != "" {
		return p
	}

	return s.projection
}

func (s *service) CreateContact(ctx context.Context, p *ContactKind) (*ContactKind, error) {
	buf := &bytes.Buffer{}
	e := xml.NewEncoder(buf)
	err := e.Encode(p)
	if err != nil {
		defer e.Close()
		return nil, err
	}
	e.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"/"+s.projection, buf)
	if err != nil {
		return nil, fmt.Errorf("CreateContact error: could not create new request: %w", err)
	}

	res, err := s.base.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CreateContact error: %w", err)
	}

	switch res.StatusCode {
	case http.StatusCreated:
		d := xml.NewDecoder(res.Body)
		defer res.Body.Close()
		var ct ContactKind
		err = d.Decode(&ct)
		if err != nil {
			return nil, err
		}
		return &ct, nil
	case http.StatusConflict:
		return nil, fmt.Errorf("CreateContact error: version conflict")
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return nil, fmt.Errorf("CreateContact error: %s", res.Status)
	default:
		return nil, fmt.Errorf("CreateContact error: unknown with %s", res.Status)
	}

}

func (s *service) GetContact(ctx context.Context, id string, projection string, etag string) (*ContactKind, error) {
	return s.getContact(ctx, id, projection, etag, "could not get a contact from GetContact")
}

func (s *service) getContact(ctx context.Context, id string, projection string, etag string, errPrefix string) (*ContactKind, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/%s", s.endpoint, s.getPojection(projection), id), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}
	if etag != "" && etag != "*" {
		req.Header.Set("If-None-Match", etag)
	}

	res, err := s.base.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}

	if res.StatusCode == http.StatusNotModified {
		// use empty value as a signal
		// this obviously is not the best way, but let's ues it now.
		return nil, nil
	}

	dec := xml.NewDecoder(res.Body)
	defer res.Body.Close()
	var contact ContactKind
	err = dec.Decode(&contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// QueryStatus stores the querying state of the feed.
type QueryStatus struct {
	Updated time.Time
	Etag    string
}

// By default, the entries in a feed aren't ordered.
func (s *service) ListContacts(ctx context.Context, projection, etag string, queries ...func(url.Values)) ([]*ContactKind, *QueryStatus, error) {
	params := url.Values{}
	var u string
	if len(queries) > 0 {
		// add strict
		withStrict()(params)
		for _, q := range queries {
			q(params)
		}

		u = fmt.Sprintf("%s/%s?%s", s.endpoint, s.getPojection(projection), params.Encode())
	} else {
		u = fmt.Sprintf("%s/%s", s.endpoint, s.getPojection(projection))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("ListContacts error: could not create a HTTP request: %w", err)
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	type feed struct {
		Etag    string    `xml:"etag,attr"`
		Updated time.Time `xml:"updated"`
		//		TotalResults int           `xml:"totalResults"`
		Links    []Link        `xml:"link"`
		Contacts []ContactKind `xml:"http://www.w3.org/2005/Atom entry"`
	}

	st := new(QueryStatus)
	ret := make([]*ContactKind, 0, 20)
	var f *feed
	for req != nil {
		res, err := s.base.Do(req)
		if err != nil {
			return nil, nil, err
		}
		f = new(feed)
		dec := xml.NewDecoder(res.Body)
		if err = dec.Decode(f); err != nil {
			defer res.Body.Close()
			return nil, nil, fmt.Errorf("ListContact error: %w", err)
		}
		res.Body.Close()
		for _, ct := range f.Contacts {
			o := ct.Clone()
			ret = append(ret, &o)
		}

		for _, l := range f.Links {
			if l.Related == "next" {
				req, _ = http.NewRequestWithContext(ctx, http.MethodGet, l.Href, nil)
				break
			}
			req = nil
		}
		if req == nil {
			st.Etag = f.Etag
			st.Updated = f.Updated
		}
	}

	return ret, st, nil
}

func (s *service) UpdateContact(ctx context.Context, id, etag string, p *ContactKind) (*ContactKind, error) {
	op, err := s.getContact(ctx, id, "full", "", "UpdateContact error: could not get a contact")
	if err != nil {
		return nil, err
	}

	if op.etag != etag && etag != "*" {
		return nil, fmt.Errorf("UpdateContact error: etag not match")
	}

	url := op.editLink
	buf := &bytes.Buffer{}
	enc := xml.NewEncoder(buf)
	// maybe merge op and p
	err = enc.Encode(p)
	if err != nil {
		defer enc.Close()
		return nil, fmt.Errorf("could not encode xml payload from UpdateContact: %w", err)
	}
	enc.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, buf)
	if err != nil {
		return nil, fmt.Errorf("could not create a HTTP request from UpdateContact: %w", err)
	}

	// If-Match
	req.Header.Set("If-Match", etag)

	res, err := s.base.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expect get HTTP status OK, got: %s", res.Status)
	}

	dec := xml.NewDecoder(res.Body)
	defer res.Body.Close()
	var ret ContactKind
	if err = dec.Decode(&ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

// DeleteContact delete a contact.
func (s *service) DeleteContact(ctx context.Context, id, etag string) error {
	op, err := s.getContact(ctx, id, "thin", "", "could not get a contact from DeleteContact")
	if err != nil {
		return err
	}

	if op.etag != etag && etag != "*" {
		return fmt.Errorf("UpdateContact error: etag not match")
	}

	url := op.editLink
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("DeleteContact error: could not create a HTTP request: %w", err)
	}

	// If-Match
	req.Header.Set("If-Match", etag)
	_, err = s.base.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteContact error: failed to call: %w", err)
	}

	return err
}
