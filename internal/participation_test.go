package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

func Test_ListParticipationKeys(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetPartKeys(ctx, client)

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err = api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}

	keys, err := GetPartKeys(ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(keys)
}

func Test_ReadParticipationKey(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ReadPartKey(ctx, client, "unknown")

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err = api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}

	keys, err := GetPartKeys(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	if keys == nil {
		t.Fatal(err)
	}

	_, err = ReadPartKey(ctx, client, (*keys)[0].Id)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_GenerateParticipationKey(t *testing.T) {
	ctx := context.Background()

	// Create Client
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}

	// Generate error
	_, err = GenerateKeyPair(ctx, client, "", nil)
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err = api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}

	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     10000,
	}

	// This returns nothing and sucks
	key, err := GenerateKeyPair(ctx, client, "QNZ7GONNHTNXFW56Y24CNJQEMYKZKKI566ASNSWPD24VSGKJWHGO6QOP7U", &params)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(key)
}

func Test_DeleteParticipationKey(t *testing.T) {
	ctx := context.Background()
	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}
	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     10000,
	}
	key, err := GenerateKeyPair(ctx, client, "QNZ7GONNHTNXFW56Y24CNJQEMYKZKKI566ASNSWPD24VSGKJWHGO6QOP7U", &params)
	if err != nil {
		t.Fatal(err)
	}

	err = DeletePartKey(ctx, client, key.Id)
	if err != nil {
		t.Fatal(err)
	}
}
func Test_RemovePartKeyByID(t *testing.T) {
	// Test case: Remove an existing key
	t.Run("Remove existing key", func(t *testing.T) {
		keys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		expectedKeys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key3"},
		}
		RemovePartKeyByID(&keys, "key2")
		if len(keys) != len(expectedKeys) {
			t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(keys))
		}
		for i, key := range keys {
			if key.Id != expectedKeys[i].Id {
				t.Fatalf("expected key ID %s, got %s", expectedKeys[i].Id, key.Id)
			}
		}
	})

	// Test case: Remove a non-existing key
	t.Run("Remove non-existing key", func(t *testing.T) {
		keys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		expectedKeys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		RemovePartKeyByID(&keys, "key4")
		if len(keys) != len(expectedKeys) {
			t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(keys))
		}
		for i, key := range keys {
			if key.Id != expectedKeys[i].Id {
				t.Fatalf("expected key ID %s, got %s", expectedKeys[i].Id, key.Id)
			}
		}
	})

	// Test case: Remove a key from an empty list
	t.Run("Remove key from empty list", func(t *testing.T) {
		keys := []api.ParticipationKey{}
		RemovePartKeyByID(&keys, "key1")
		if len(keys) != 0 {
			t.Fatalf("expected 0 keys, got %d", len(keys))
		}
	})
}
