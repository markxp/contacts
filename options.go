package contacts

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// withReturnType changes the representation type. Support types are "atom", "rss", "json" payloads.
// Other types are:
// - json-in-script
// - atom-in-script
// - rss-in-script
// These three wraps payload with HTML script tag.
//
// If no withReturnType is called, "atom" is the default type.
//
// Deprecated: The whole library uses "atom" format though this function is useless.
func withReturnType(t string) func(url.Values) {
	return func(v url.Values) {
		v.Set("alt", t)
	}
}

// WithMaxResults override default maximum.
func WithMaxResults(n int) func(url.Values) {
	return func(v url.Values) {
		v.Set("max-results", fmt.Sprint(n))
	}
}

// WithStartIndex is the first retrived dataset. 1-based index.
// Note that this isn't a general cursoring mechanism.
// If you first send a query with ?start-index=1&max-results=10 and then send another query with ?start-index=11&max-results=10,
// the service cannot guarantee that the results are equivalent to ?start-index=1&max-results=20, because insertions and deletions could have taken place in between the two queries.
func WithStartIndex(n int) func(url.Values) {
	return func(v url.Values) {
		v.Set("start-index", fmt.Sprint(n))
	}
}

// WithUpdateMin changes the result set to changes that happended after t (inclusive).
func WithUpdateMin(t time.Time) func(url.Values) {
	return func(v url.Values) {
		v.Set("updated-min", t.Format(time.RFC3339))
	}
}

// WithUpdateMax changes the result set to changes that happended before t (exclusive).
func WithUpdateMax(t time.Time) func(url.Values) {
	return func(v url.Values) {
		v.Set("updated-max", t.Format(time.RFC3339))
	}
}

func withStrict() func(url.Values) {
	return func(v url.Values) {
		v.Set("strict", strconv.FormatBool(true))
	}
}

// WithShowDeleted shows deleted records in the result set.
// It's useful combine with WithUpdateMin.
func WithShowDeleted(b bool) func(url.Values) {
	return func(v url.Values) {
		v.Set("showdeleted", strconv.FormatBool(b))
	}
}

// WithSort sets sorting order for the result set.
// it accepts "ascending" or "descending" and sort by last modified time.
func WithSort(asc string) func(url.Values) {
	return func(v url.Values) {
		v.Set("orderby", "lastmodified")
		v.Set("sortorder", asc)
	}
}

// FilterByAuthor returns entries where the author name and/or email address match your query string.
// Support values: name or email
func FilterByAuthor(name string) func(url.Values) {
	return func(v url.Values) {
		v.Set("author", name)
	}
}

// FilterByCategory filters results with filters.
// filters uses the following syntax:
//
// To do an OR between terms, use a pipe character (|), URL-encoded as %7C.
// For example: http://www.example.com/feeds?category=Fritz%7CLaurie returns entries that match either category.
//
// To do an AND between terms, use a comma character (,).
// For example: http://www.example.com/feeds?category=Fritz,Laurie returns entries that match both categories.
func FilterByCategory(filters string) func(url.Values) {
	return func(v url.Values) {
		v.Set("category", filters)
	}
}

// WithTextQuery enables full-text queries on result sets.
// text must have the following formats:
//
// To exclude entries that match a given term, use the form q=-term.
// The search is case-insensitive.
//
// Example: to search for all entries that contain the exact phrase "Elizabeth Bennet" and the word "Darcy" but don't contain the word "Austen",
// use the following query: ?q="Elizabeth Bennet" Darcy -Austen
func WithTextQuery(texts []string) func(url.Values) {
	return func(v url.Values) {
		var b bytes.Buffer
		for idx, t := range texts {
			// put logical AND if more than one
			if idx != 0 {
				b.WriteString(" ")
			}
			if strings.HasPrefix(t, "-") {
				b.WriteString(fmt.Sprintf(`-"%s"`, strings.TrimPrefix(t, "-")))
			}
		}

		v.Set("q", b.String())
	}
}
