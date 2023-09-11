package main

import (
	"context"
	"fmt"
	"os"

	"github.com/markxp/contacts"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
)

func main() {
	ctx := context.Background()

	// generate OAuth2 token source
	//
	// Here we choose Google Cloud service account & Workspace impersonate a real user.
	ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: "directory-service@gcp-project.iam.gserviceaccount.com",
		Subject:         "user@legispect.com",
		Scopes:          contacts.ScopePeopleAPI,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create oauth2 token source: %v", err)
		os.Exit(1)
	}

	svc, err := contacts.NewService(oauth2.NewClient(ctx, ts), "legispect.com", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create contacts.Service: %v", err)
		os.Exit(1)
	}

	ret, st, err := svc.ListContacts(ctx, "full", "", contacts.WithMaxResults(1000))
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create contact: %v", err)
		os.Exit(1)
	}

	fmt.Printf("list status: updated=%s etag=%s\n", st.Updated.String(), st.Etag)

	for _, v := range ret {
		fmt.Printf("%s %s %s\n", v.GetID(), v.Name.FullName, v.GetEtag())
	}

}
