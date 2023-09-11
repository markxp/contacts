package contacts

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// UnmarshalXML implements xml.Unmarshaler.
// In the unmarhal processing, common element or server-only element will be read.
func (c *ContactKind) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type decodeContactKind struct {
		XMLName  xml.Name `xml:"http://www.w3.org/2005/Atom entry"`
		Etag     string   `xml:"etag,attr"`
		Category struct {
			Term string `xml:"term,attr"`
		} `xml:"category"`
		ID                      string                      `xml:"id"`
		Updated                 time.Time                   `xml:"updated"`
		Title                   string                      `xml:"title"`
		Content                 string                      `xml:"content"`
		Name                    GDName                      `xml:"http://schemas.google.com/g/2005 name"`
		Email                   []GDEmail                   `xml:"http://schemas.google.com/g/2005 email"`
		Deleted                 bool                        `xml:"http://schemas.google.com/g/2005 deleted"`
		PhoneNumber             []GDPhoneNumber             `xml:"http://schemas.google.com/g/2005 phoneNumber"`
		StructuredPostalAddress []GDStructuredPostalAddress `xml:"http://schemas.google.com/g/2005 structuredPostalAddress"`
		Link                    []Link                      `xml:"http://www.w3.org/2005/Atom link"`
		// gd:extendedProperty*
		ExtendedProperty []GDExtendedProperty `xml:"http://schemas.google.com/g/2005 extendedProperty"`
		// gd:im*
		IM []GDIM `xml:"http://schemas.google.com/g/2005 im"`
		// gd:organization*
		Organization []GDOrganization `xml:"http://schemas.google.com/g/2005 organization"`
	}

	var o decodeContactKind
	err := d.DecodeElement(&o, &start)
	if err != nil {
		return err
	}
	const contactTerm = "http://schemas.google.com/contact/2008#contact"
	if o.Category.Term != contactTerm {
		return fmt.Errorf("xml type not match: expect %s, got %s", contactTerm, o.Category.Term)
	}

	c.Name = GDName{
		GivenName:      o.Name.GivenName,
		AdditionalName: o.Name.AdditionalName,
		FamilyName:     o.Name.FamilyName,
		Prefix:         o.Name.Prefix,
		Suffix:         o.Name.Suffix,
		FullName:       o.Name.FullName,
	}
	c.Email = make([]GDEmail, 0, len(o.Email))
	c.Email = append(c.Email, o.Email...)
	c.IM = make([]GDIM, 0, len(o.IM))
	c.IM = append(c.IM, o.IM...)
	c.PhoneNumber = make([]GDPhoneNumber, 0, len(o.PhoneNumber))
	c.PhoneNumber = append(c.PhoneNumber, o.PhoneNumber...)
	c.StructuredPostalAddress = make([]GDStructuredPostalAddress, 0, len(o.StructuredPostalAddress))
	c.StructuredPostalAddress = append(c.StructuredPostalAddress, o.StructuredPostalAddress...)

	for _, l := range o.Link {
		switch l.Related {
		case "http://schemas.google.com/contacts/2008/rel#photo":
			c.photoLink = l.Href
		case "self":
			c.selfLink = l.Href
		case "edit":
			c.editLink = l.Href
		}
	}

	c.deleted = o.Deleted
	c.id = o.ID
	c.updated = o.Updated
	c.content = o.Content
	c.etag = o.Etag

	c.ExtendedProperty = make(map[string]string, len(o.ExtendedProperty))
	for _, pair := range o.ExtendedProperty {
		c.ExtendedProperty[pair.Name] = pair.Value
	}
	return nil
}

// MarshalXML implements xml.Marshaler.
// It hides unnecessory fields when sending a request to server.
func (c ContactKind) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type encodeContactKind struct {
		Name                    GDName                      `xml:"gd:name"`
		Email                   []GDEmail                   `xml:"gd:email,omitempty"`
		PhoneNumber             []GDPhoneNumber             `xml:"gd:phoneNumber,omitempty"`
		StructuredPostalAddress []GDStructuredPostalAddress `xml:"gd:structuredPostalAddress,omitempty"`
		Content                 string                      `xml:"content"`
		// atom:category
		Category struct {
			Scheme string `xml:"scheme,attr"`
			Term   string `xml:"term,attr"`
		} `xml:"category"`

		// gd:extendedProperty*
		ExtendedProperty []GDExtendedProperty `xml:"gd:extendedProperty,omitempty"`
		IM               []GDIM               `xml:"gd:im,omitempty"`

		// Organization []GDOrganization `xml:"gd:organization"`
	}

	type category struct {
		Scheme string `xml:"scheme,attr"`
		Term   string `xml:"term,attr"`
	}

	var cat category = category{
		Scheme: "http://schemas.google.com/g/2005#kind",
		Term:   "http://schemas.google.com/contact/2008#contact",
	}

	var o encodeContactKind
	o.Content = c.content
	o.Name = GDName{
		GivenName:      c.Name.GivenName,
		AdditionalName: c.Name.AdditionalName,
		FamilyName:     c.Name.FamilyName,
		Prefix:         c.Name.Prefix,
		Suffix:         c.Name.Suffix,
		FullName:       c.Name.FullName,
	}
	o.Email = make([]GDEmail, 0, len(c.Email))
	o.Email = append(o.Email, c.Email...)
	o.PhoneNumber = make([]GDPhoneNumber, len(c.PhoneNumber))
	o.PhoneNumber = append(o.PhoneNumber, c.PhoneNumber...)
	o.StructuredPostalAddress = make([]GDStructuredPostalAddress, len(c.StructuredPostalAddress))
	o.StructuredPostalAddress = append(o.StructuredPostalAddress, c.StructuredPostalAddress...)

	o.IM = make([]GDIM, len(c.IM))
	o.IM = append(o.IM, c.IM...)

	o.ExtendedProperty = make([]GDExtendedProperty, len(c.ExtendedProperty))
	for k, v := range c.ExtendedProperty {
		o.ExtendedProperty = append(o.ExtendedProperty, GDExtendedProperty{
			Name:  k,
			Value: v,
		})
	}

	start.Name = xml.Name{Space: "", Local: "entry"}
	attrs := make([]xml.Attr, 0, 2)
	attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "xmlns:atom"}, Value: "http://www.w3.org/2005/Atom"})
	attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "xmlns:gd"}, Value: "http://schemas.google.com/g/2005"})
	start.Attr = attrs
	o.Category = cat

	return e.EncodeElement(o, start)
}

// GDName allows storing person's name in a structured way. Consists of given name, additional name, family name, prefix, suffix and full name.
type GDName struct {
	GivenName      string
	AdditionalName string
	FamilyName     string
	Prefix         string
	Suffix         string
	FullName       string
}

// UnmarshalXML implements xml.Unmarshaler.
func (n *GDName) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type decodeGDName struct {
		GivenName      string `xml:"givenName"`
		AdditionalName string `xml:"additionalName"`
		FamilyName     string `xml:"familyName"`
		Prefix         string `xml:"namePrefix"`
		Suffix         string `xml:"nameSuffix"`
		FullName       string `xml:"fullName"`
	}

	var o decodeGDName
	if err := d.DecodeElement(&o, &start); err != nil {
		return err
	}
	n.GivenName = o.GivenName
	n.AdditionalName = o.AdditionalName
	n.FamilyName = o.FamilyName
	n.Prefix = o.Prefix
	n.Suffix = o.Suffix
	n.FullName = strings.TrimSpace(o.FullName)

	return nil
}

// MarshalXML implements xml.Marshaler.
func (n GDName) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "gd:name",
	}

	type encodedGDName struct {
		GivenName      string `xml:"gd:givenName,omitempty"`
		AdditionalName string `xml:"gd:additionalName,omitempty"`
		FamilyName     string `xml:"gd:familyName,omitempty"`
		Prefix         string `xml:"gd:namePrefix,omitempty"`
		Suffix         string `xml:"gd:nameSuffix,omitempty"`
		FullName       string `xml:"gd:fullName,omitempty"`
	}

	o := encodedGDName{
		GivenName:      n.GivenName,
		AdditionalName: n.AdditionalName,
		FamilyName:     n.FamilyName,
		Prefix:         n.Prefix,
		Suffix:         n.Suffix,
		FullName:       strings.TrimSpace(n.FullName),
	}
	return e.EncodeElement(o, start)
}

// GDEmail saves an email address.
// It's "rel" field has 3 possible values.
// - http://schemas.google.com/g/2005#home
// - http://schemas.google.com/g/2005#other
// - http://schemas.google.com/g/2005#work
// If it uses "http://schemas.google.com/g/2005#other" in the "rel" field,
// you should use "label" to express the real relation of the entity.
type GDEmail struct {
	Address     string `xml:"address,attr"`
	Related     string `xml:"rel,attr,omitempty"`
	Label       string `xml:"label,attr,omitempty"`
	Primary     bool   `xml:"primary,attr,omitempty"`
	DisplayName string `xml:"displayName,attr,omitempty"`
}

// MarshalXML implements xml.Marshaler.
func (m GDEmail) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "gd:email",
	}

	type encodedGDEmail struct {
		Address     string `xml:"address,attr"`
		Related     string `xml:"rel,attr,omitempty"`
		Label       string `xml:"label,attr,omitempty"`
		Primary     bool   `xml:"primary,attr,omitempty"`
		DisplayName string `xml:"displayName,attr,omitempty"`
	}

	var obj encodedGDEmail = encodedGDEmail(m)

	return e.EncodeElement(obj, start)
}

// GDPhoneNumber saves a phone number.
// It's "rel" field has many possible values.
// - http://schemas.google.com/g/2005#assistant
// - http://schemas.google.com/g/2005#callback
// - http://schemas.google.com/g/2005#car
// - http://schemas.google.com/g/2005#company_main
// - http://schemas.google.com/g/2005#fax
// - http://schemas.google.com/g/2005#home
// - http://schemas.google.com/g/2005#home_fax
// - http://schemas.google.com/g/2005#isdn
// - http://schemas.google.com/g/2005#main
// - http://schemas.google.com/g/2005#mobile
// - http://schemas.google.com/g/2005#other
// - http://schemas.google.com/g/2005#other_fax
// - http://schemas.google.com/g/2005#pager
// - http://schemas.google.com/g/2005#radio
// - http://schemas.google.com/g/2005#telex
// - http://schemas.google.com/g/2005#tty_tdd
// - http://schemas.google.com/g/2005#work
// - http://schemas.google.com/g/2005#work_fax
// - http://schemas.google.com/g/2005#work_mobile
// - http://schemas.google.com/g/2005#work_pager
// If "rel" equals to "http://schemas.google.com/g/2005#other",
// use "label" to express the real relation.
type GDPhoneNumber struct {
	Related    string `xml:"rel,attr,omitempty"`
	Label      string `xml:"label,attr,omitempty"`
	URI        string `xml:"uri,attr,omitempty"`
	Primary    bool   `xml:"primary,attr,omitempty"`
	DialNumber string `xml:",chardata"` // it may contain white spaces.
}

// UnmarshalXML implements xml.Unmarshaler.
func (n *GDPhoneNumber) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type decodeGDPhoneNumber struct {
		Related    string `xml:"rel,attr,omitempty"`
		Label      string `xml:"label,attr,omitempty"`
		URI        string `xml:"uri,attr,omitempty"`
		Primary    bool   `xml:"primary,attr,omitempty"`
		DialNumber string `xml:",chardata"` // it may contain white spaces.
	}

	var o decodeGDPhoneNumber
	err := d.DecodeElement(&o, &start)
	if err != nil {
		return err
	}

	n.Related = o.Related
	n.Label = o.Label
	n.URI = o.URI
	n.Primary = o.Primary
	n.DialNumber = strings.TrimSpace(o.DialNumber)

	return nil
}

// MarshalXML implements xml.Marshaler.
func (n GDPhoneNumber) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "gd:phoneNumber",
	}
	type encodeGDPhoneNumber struct {
		Related    string `xml:"rel,attr,omitempty"`
		Label      string `xml:"label,attr,omitempty"`
		URI        string `xml:"uri,attr,omitempty"`
		Primary    bool   `xml:"primary,attr,omitempty"`
		DialNumber string `xml:",chardata"` // it may contain white spaces.
	}
	var obj = encodeGDPhoneNumber(n)
	obj.DialNumber = strings.TrimSpace(obj.DialNumber)
	return e.EncodeElement(obj, start)
}

// GDIM saves an instant message account.
// It's "rel" field has the following possible values.
// - http://schemas.google.com/g/2005#home
// - http://schemas.google.com/g/2005#netmeeting
// - http://schemas.google.com/g/2005#other
// - http://schemas.google.com/g/2005#work
// If the "rel" field equals to "http://schemas.google.com/g/2005#other",
// it uses "label" to express the real relation.
type GDIM struct {
	Address  string `xml:"address,attr"`
	Label    string `xml:"label,attr,omitempty"`
	Related  string `xml:"rel,attr,omitempty"`
	Protocol string `xml:"protocol,attr,omitempty"`
	Primary  bool   `xml:"primary,attr,omitempty"`
}

// MarshalXML implements xml.Marshaler.
func (im GDIM) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "gd:im",
	}
	type encodeGDIM struct {
		Address  string `xml:"address,attr"`
		Label    string `xml:"label,attr,omitempty"`
		Related  string `xml:"rel,attr,omitempty"`
		Protocol string `xml:"protocol,attr,omitempty"`
		Primary  bool   `xml:"primary,attr,omitempty"`
	}
	var obj = encodeGDIM(im)
	return e.EncodeElement(obj, start)
}

// GDOrganization saves an organization occupation of the contact person.
// NOT IMPLEMENTED YET
type GDOrganization struct {
}

// GDStructuredPostalAddress saves postal address.
// It's "rel" field has the following possible values
// - http://schemas.google.com/g/2005#work
// - http://schemas.google.com/g/2005#home
// - http://schemas.google.com/g/2005#other
//
// "mailClass" field has the following possible values
// - http://schemas.google.com/g/2005#both
// - http://schemas.google.com/g/2005#letters
// - http://schemas.google.com/g/2005#parcels
// - http://schemas.google.com/g/2005#neither
//
// "usage" field has the following possible values
// http://schemas.google.com/g/2005#general
// http://schemas.google.com/g/2005#local
type GDStructuredPostalAddress struct {
	Related   string
	MailClass string
	Usage     string
	Label     string
	Primary   bool

	Agent            string
	HouseName        string
	Pobox            string
	Neighborhood     string
	City             string
	Street           string
	Region           string
	SubRegion        string
	PostCode         string
	Country          string
	FormattedAddress string
}

// UnmarshalXML implements xml.Unmarshaler.
func (a *GDStructuredPostalAddress) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type decodedGDStructuredPostalAddress struct {
		Related   string `xml:"rel,attr"`
		MailClass string `xml:"mailClass,attr"`
		Usage     string `xml:"usage,attr"`
		Label     string `xml:"label,attr"`
		Primary   bool   `xml:"primary,attr"`

		Agent            string `xml:"agent"`
		HouseName        string `xml:"housename"`
		Pobox            string `xml:"pobox"`
		Neighborhood     string `xml:"neighborhood"`
		City             string `xml:"city"`
		Street           string `xml:"street"`
		Region           string `xml:"region"`
		SubRegion        string `xml:"subregion"`
		PostCode         string `xml:"postcode"`
		Country          string `xml:"country"`
		FormattedAddress string `xml:"formattedAddress"`
	}
	var o decodedGDStructuredPostalAddress
	err := d.DecodeElement(&o, &start)
	if err != nil {
		return err
	}

	a.Region = o.Related
	a.MailClass = o.MailClass
	a.Usage = o.Usage
	a.Label = o.Label
	a.Primary = o.Primary
	a.Agent = o.Agent
	a.HouseName = o.HouseName
	a.Pobox = o.Pobox
	a.Neighborhood = o.Neighborhood
	a.City = o.City
	a.Street = o.Street
	a.Region = o.Region
	a.SubRegion = o.SubRegion
	a.PostCode = o.PostCode
	a.Country = o.Country
	a.FormattedAddress = o.FormattedAddress

	return nil
}

// MarshalXML implements xml.Marshaler
func (a GDStructuredPostalAddress) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Space: "", Local: "gd:structuredPostalAddress"}
	type encodeGDStructuredPostalAddress struct {
		Related   string `xml:"rel,attr,omitempty"`
		MailClass string `xml:"mailClass,attr,omitempty"`
		Usage     string `xml:"usage,attr,omitempty"`
		Label     string `xml:"label,attr,omitempty"`
		Primary   bool   `xml:"primary,attr,omitempty"`

		Agent            string `xml:"gd:agent,omitempty"`
		HouseName        string `xml:"gd:housename,omitempty"`
		Pobox            string `xml:"gd:pobox,omitempty"`
		Neighborhood     string `xml:"gd:neighborhood,omitempty"`
		City             string `xml:"gd:city,omitempty"`
		Street           string `xml:"gd:street,omitempty"`
		Region           string `xml:"gd:region,omitempty"`
		SubRegion        string `xml:"gd:subregion,omitempty"`
		PostCode         string `xml:"gd:postcode,omitempty"`
		Country          string `xml:"gd:country,omitempty"`
		FormattedAddress string `xml:"gd:formattedAddress,omitempty"`
	}
	var o encodeGDStructuredPostalAddress

	o.Related = a.Related
	o.MailClass = a.MailClass
	o.Usage = a.Usage
	o.Label = a.Label
	o.Primary = a.Primary

	o.Agent = a.Agent
	o.HouseName = a.HouseName
	o.Pobox = a.Pobox
	o.Neighborhood = a.Neighborhood
	o.City = a.City
	o.Street = a.Street
	o.Region = a.Region
	o.SubRegion = a.SubRegion
	o.PostCode = a.PostCode
	o.Country = a.Country
	o.FormattedAddress = a.FormattedAddress

	return e.EncodeElement(o, start)
}

// GDExtendedProperty saves custom data as key-value pair.
type GDExtendedProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr,omitempty"`
	Realm string `xml:"realm,attr,omitempty"`
}

// MarshalXML implements xml.Marshaler.
func (p GDExtendedProperty) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Space: "", Local: "gd:extendedProperty"}
	type encodeGDExtendedProperty struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr,omitempty"`
		Realm string `xml:"realm,attr,omitempty"`
	}
	obj := encodeGDExtendedProperty(p)
	return e.EncodeElement(obj, start)
}

// Link saves link tags in a ContactKind
type Link struct {
	Related string `xml:"rel,attr"`
	Type    string `xml:"type,attr"`
	Href    string `xml:"href,attr"`
}
