package contacts

import (
	"encoding/xml"
	"strconv"
	"strings"
	"testing"
)

func TestGDName(t *testing.T) {
	bs := []byte(`	<gd:name>
	<gd:fullName>FIRST_NAME LAST_NAME</gd:fullName>
	<gd:givenName>FIRST_NAME</gd:givenName>
	<gd:familyName>LAST_NAME</gd:familyName>
</gd:name>`)

	var n GDName
	err := xml.Unmarshal(bs, &n)
	if err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}
	// CHECK FIELDS
	if n.FullName != "FIRST_NAME LAST_NAME" || n.GivenName != "FIRST_NAME" || n.FamilyName != "LAST_NAME" {
		t.Fatalf("xml unmarshal error: missing value mappings")
	}

	nn := GDName{
		FullName: "FIRST_NAME LAST_NAME",
	}

	b, err := xml.Marshal(nn)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}
	if string(b) != `<gd:name><gd:fullName>FIRST_NAME LAST_NAME</gd:fullName></gd:name>` {
		t.Fatalf("xml marshal error: not match, got %s", b)
	}
}

func TestGDEmail(t *testing.T) {
	bs := []byte(`<gd:email address="fubar@gmail.com" rel="http://schemas.google.com/g/2005#home" label="Personal" primary="true"></gd:email>`)
	var m GDEmail
	err := xml.Unmarshal(bs, &m)
	if err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if m.Address != "fubar@gmail.com" || m.Primary != true || m.Related != "http://schemas.google.com/g/2005#home" || m.Label != "Personal" {
		t.Fatalf("xml unmarshal error: missing value mappings")
	}

	b, err := xml.Marshal(m)
	if string(b) != string(bs) {
		t.Fatalf("xml marshal error: not match, got %s", string(b))
	}
}

func TestGDPhoneNumber(t *testing.T) {
	bs := []byte(`<gd:phoneNumber rel="http://schemas.google.com/g/2005#work" uri="tel:+1-425-555-8080;ext=52585">
  (425) 555-8080 ext. 52585
</gd:phoneNumber>`)

	var n GDPhoneNumber
	if err := xml.Unmarshal(bs, &n); err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if n.Related != "http://schemas.google.com/g/2005#work" || n.URI != "tel:+1-425-555-8080;ext=52585" ||
		n.Primary != false || n.Label != "" || n.DialNumber != "(425) 555-8080 ext. 52585" {
		t.Fatalf("xml unmarshal error: not match")
	}

	b, err := xml.Marshal(n)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	s := string(b)
	if !strings.Contains(s, "<gd:phoneNumber") || !strings.Contains(s, `rel="http://schemas.google.com/g/2005#work"`) ||
		!strings.Contains(s, `uri="tel:+1-425-555-8080;ext=52585"`) {

		t.Fatalf("xml marshal error: not match")
	}

	if strings.Contains(s, "primary") {
		t.Fatalf("xml marshal: not emit empty value")
	}
}

func TestGDIM(t *testing.T) {
	bs := []byte(`<gd:im protocol="http://schemas.google.com/g/2005#MSN" address="foo@bar.msn.com" rel="http://schemas.google.com/g/2005#home" primary="true"/>`)
	var im GDIM
	if err := xml.Unmarshal([]byte(bs), &im); err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if im.Address != "foo@bar.msn.com" || im.Label != "" || strconv.FormatBool(im.Primary) != strconv.FormatBool(true) || im.Protocol != "http://schemas.google.com/g/2005#MSN" || im.Related != "http://schemas.google.com/g/2005#home" {
		t.Fatalf("xml unmarshal error: not match")
	}

	b, err := xml.Marshal(im)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}
	s := string(b)

	if !strings.Contains(s, `<gd:im`) || !strings.Contains(s, `protocol="http://schemas.google.com/g/2005#MSN"`) || !strings.Contains(s, `address="foo@bar.msn.com"`) ||
		!strings.Contains(s, `rel="http://schemas.google.com/g/2005#home"`) || !strings.Contains(s, `primary="true"`) {

		t.Fatalf("xml marshal error: not match")
	}
}

func TestGDPostalAddress(t *testing.T) {
	bs := []byte(`<gd:structuredPostalAddress mailClass='http://schemas.google.com/g/2005#letters' label='John at Google'>
  <gd:street>1600 Amphitheatre Parkway</gd:street>
  <gd:city>Mountain View</gd:city>
  <gd:region>CA</gd:region>
  <gd:postcode>94043</gd:postcode>
</gd:structuredPostalAddress>`)

	var a GDStructuredPostalAddress
	if err := xml.Unmarshal(bs, &a); err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if a.Agent != "" || a.City != "Mountain View" || a.Country != "" || a.Region != "CA" ||
		a.PostCode != "94043" || a.Street != "1600 Amphitheatre Parkway" {

		t.Fatalf("xml unmarshal error: not match")
	}

	b, err := xml.Marshal(a)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}

	s := string(b)
	if !strings.Contains(s, "<gd:structuredPostalAddress") ||
		!strings.Contains(s, `<gd:street>1600 Amphitheatre Parkway</gd:street>`) ||
		!strings.Contains(s, `<gd:city>Mountain View</gd:city>`) ||
		!strings.Contains(s, `<gd:region>CA</gd:region>`) ||
		!strings.Contains(s, `<gd:postcode>94043</gd:postcode>`) {

		t.Fatalf("xml marshal error: not match")
	}
}

func TestContactKind(t *testing.T) {
	bs := []byte(`<entry xmlns='http://www.w3.org/2005/Atom' xmlns:gd='http://schemas.google.com/g/2005'>
  <category scheme='http://schemas.google.com/g/2005#kind' 
      term='http://schemas.google.com/contact/2008#contact'/>
  <title>Elizabeth Bennet</title>
	<id>http://www.google.com/m8/feeds/contacts/legispect.com/base/20017e218fa39973</id>
	<updated>2023-08-18T09:54:17.202Z</updated>
  <content>My good friend, Liz.  A little quick to judge sometimes, but nice girl.</content>
  <gd:email rel='http://schemas.google.com/g/2005#work' primary='true' address='liz@gmail.com'/>
  <gd:email rel='http://schemas.google.com/g/2005#home' address='liz@example.org'/>
  <gd:phoneNumber rel='http://schemas.google.com/g/2005#work' primary='true'>
    (206)555-1212
  </gd:phoneNumber>
  <gd:phoneNumber rel='http://schemas.google.com/g/2005#home'>
    (206)555-1213
  </gd:phoneNumber>
  <gd:phoneNumber rel='http://schemas.google.com/g/2005#mobile'>
    (206) 555-1212
  </gd:phoneNumber>
  <gd:im rel='http://schemas.google.com/g/2005#home' 
      protocol='http://schemas.google.com/g/2005#GOOGLE_TALK' 
      address='liz@gmail.com'/>
  <gd:structuredPostalAddress rel='http://schemas.google.com/g/2005#work' primary='true'>
    <gd:street>1600 Amphitheatre Pkwy</gd:street>
    <gd:city>Mountain View</gd:city>
		<gd:region>CA</gd:region>
		<gd:postcode>94043</gd:postcode>
  </gd:structuredPostalAddress>
  <gd:structuredPostalAddress rel='http://schemas.google.com/g/2005#home'>
		<gd:street>800 Main Street</gd:street>
    <gd:city>Mountain View</gd:city>
		<gd:region>CA</gd:region>
		<gd:postcode>94041</gd:postcode>
  </gd:structuredPostalAddress>
</entry>`)

	var c ContactKind
	if err := xml.Unmarshal(bs, &c); err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if c.id != "http://www.google.com/m8/feeds/contacts/legispect.com/base/20017e218fa39973" ||
		c.updated.IsZero() {

		t.Fatalf("xml unmarshal: missing metadata")
	}

	if c.content != "My good friend, Liz.  A little quick to judge sometimes, but nice girl." ||
		len(c.Email) != 2 || len(c.PhoneNumber) != 3 || len(c.IM) != 1 ||
		len(c.StructuredPostalAddress) != 2 {

		t.Logf("%s, %d, %d, %d, %d", c.content, len(c.Email), len(c.PhoneNumber), len(c.IM), len(c.StructuredPostalAddress))
		t.Fatalf("xml unmarshal error: not match")
	}

	if c.Email[0].Address != "liz@gmail.com" {
		t.Fatalf("xml unmarshal error: inner xml not match in Email")
	}

	b, err := xml.Marshal(c)
	if err != nil {
		t.Fatalf("xml marshal error: %v", err)
	}
	var s = string(b)

	if !strings.Contains(s, `<entry`) || !strings.Contains(s, `</entry>`) ||
		!strings.Contains(s, `<category`) {

		t.Logf("%s", s)
		t.Fatalf("xml marshal error: %v", err)
	}

	c = ContactKind{}
	bs = []byte(`<entry xmlns='http://www.w3.org/2005/Atom' xmlns:gd='http://schemas.google.com/g/2005'>
  <category scheme='http://schemas.google.com/g/2005#kind' 
      term='http://schemas.google.com/contact/2008#contact'/>
	
	<gd:name>
  <gd:givenName>Winston</gd:givenName>
  <gd:additionalName>Leonard</gd:additionalName>
  <gd:familyName>Spencer-Churchill</gd:familyName>
  <gd:namePrefix>Sir</gd:namePrefix>
  <gd:nameSuffix>OG</gd:nameSuffix>
</gd:name>

</entry>
	`)

	err = xml.Unmarshal(bs, &c)
	if err != nil {
		t.Fatalf("xml unmarshal error: %v", err)
	}

	if c.Name.GivenName != "Winston" || c.Name.AdditionalName != "Leonard" || c.Name.FamilyName != "Spencer-Churchill" {
		t.Fatalf("xml unmarshal: not match in name, got %q %q %q", c.Name.GivenName, c.Name.AdditionalName, c.Name.FamilyName)
	}

}
