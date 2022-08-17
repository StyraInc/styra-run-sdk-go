package v1

import (
	"context"
	"testing"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
)

const (
	token      = ""
	url        = ""
	tenant     = "acmecorp"
	subject    = "alice"
	maxRetries = 3
)

var (
	users = []*User{
		{Id: "alice"},
		{Id: "bob"},
		{Id: "bryan"},
		{Id: "cesar"},
		{Id: "emily"},
		{Id: "gary"},
		{Id: "henry"},
		{Id: "kevin"},
	}
)

func TestRbac(t *testing.T) {
	ctx := context.Background()
	myClient := api.New(
		&api.Settings{
			Token:             token,
			Url:               url,
			DiscoveryStrategy: api.Simple,
			MaxRetries:        maxRetries,
		},
	)

	myRbac := New(
		&Settings{
			Client: myClient,
		},
	)

	authz := &api.Authz{
		Tenant:  tenant,
		Subject: subject,
	}

	if result, err := myRbac.GetRoles(ctx, authz); err != nil {
		t.Error(err)
	} else {
		_ = result
	}

	if result, err := myRbac.ListUserBindings(ctx, authz, users); err != nil {
		t.Error(err)
	} else {
		_ = result
	}

	bruce := &User{
		Id: "cesar",
	}

	if result, err := myRbac.GetUserBinding(ctx, authz, bruce); err != nil {
		t.Error(err)
	} else {
		_ = result
	}

	binding := &UserBinding{
		Roles: []string{"OWNER"},
	}

	if err := myRbac.PutUserBinding(ctx, authz, bruce, binding); err != nil {
		t.Error(err)
	}

	if err := myRbac.DeleteUserBinding(ctx, authz, bruce); err != nil {
		t.Error(err)
	}
}
