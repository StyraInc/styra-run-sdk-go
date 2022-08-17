package v1

import (
	"context"
	"testing"
)

const (
	token      = ""
	url        = ""
	path       = "rbac/user_bindings/acmecorp"
	query      = "tickets/resolve/allow"
	batches    = 3
	maxRetries = 3
)

var (
	data = map[string]interface{}{
		"alice": []string{"ADMIN"},
		"bob":   []string{"VIEWER"},
		"cesar": []string{"VIEWER"},
		"kevin": []string{"ADMIN"},
	}

	allow = map[string]string{
		"tenant":  "acmecorp",
		"subject": "alice",
	}

	deny = map[string]string{
		"tenant":  "acmecorp",
		"subject": "bob",
	}

	bindings = []string{"VIEWER", "ADMIN", "OWNER"}
)

func TestData(t *testing.T) {
	ctx := context.Background()
	myClient := New(
		&Settings{
			Token:             token,
			Url:               url,
			DiscoveryStrategy: Simple,
			MaxRetries:        maxRetries,
		},
	)

	var result interface{}
	if err := myClient.GetData(ctx, path, &result); err != nil {
		t.Error(err)
	}

	if err := myClient.PutData(ctx, path, data); err != nil {
		t.Error(err)
	}

	if err := myClient.Query(ctx, query, allow, &result); err != nil {
		t.Error(err)
	}

	if result, err := myClient.Check(ctx, query, allow); err != nil {
		t.Error(err)
	} else {
		_ = result
	}

	{
		queries := make([]Query, 0)
		for i := 0; i < batches; i++ {
			queries = append(
				queries,
				Query{
					Path:  query,
					Input: allow,
				},
			)

			queries = append(
				queries,
				Query{
					Path:  query,
					Input: deny,
				},
			)

			queries = append(
				queries,
				Query{
					Path:  "bogus",
					Input: deny,
				},
			)
		}

		if err := myClient.BatchQuery(ctx, queries, nil); err != nil {
			t.Error(err)
		}
	}

	if err := myClient.DeleteData(ctx, path); err != nil {
		t.Error(err)
	}

	if err := myClient.GetData(ctx, path, &result); err == nil {
		t.Error("this should fail ..")
	}

	if err := myClient.PutData(ctx, path, data); err != nil {
		t.Error(err)
	}
}
