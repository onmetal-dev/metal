package serverprovider

import (
	"os"
	"testing"

	"github.com/floshodan/hrobot-go/hrobot"
)

func TestHetznerProvider(t *testing.T) {
	token := os.Getenv("TEST_HETZNER_TOKEN")
	sshKeyFingerprint := os.Getenv("TEST_HETZNER_SSH_KEY_FINGERPRINT")
	if token == "" || sshKeyFingerprint == "" {
		t.Skip("Skipping Hetzner tests: TEST_HETZNER_TOKEN or TEST_HETZNER_SSH_KEY_FINGERPRINT not set")
	}
	client := hrobot.NewClient(hrobot.WithToken(token))
	hetznerProvider, err := NewHetzner(
		WithHrobotClient(client),
		WithAuthorizedKeyFingerprint(sshKeyFingerprint),
		WithTestMode(true),
	)
	if err != nil {
		t.Fatalf("Failed to create Hetzner provider: %v", err)
	}

	// Define test cases
	getCurrentOfferingsTests := []GetCurrentOfferingsTest{
		{
			Name: "Get current offerings",
			ExpectedOfferings: func(offerings []Offering) bool {
				return len(offerings) > 0
			},
			ExpectedError: func(err error) bool {
				return err == nil
			},
		},
	}

	orderServerTests := []OrderServerTest{
		{
			Name: "Order server",
			Input: Order{
				OfferingId: "AX102",
				LocationId: "HEL1",
				AddonIds:   []string{"primary_ipv4"},
			},
			ExpectedTransaction: func(tx Transaction) bool {
				return tx.Id != "" && tx.Status == TransactionStatusCanceled // in test mode Hetzner immediately cancels the order
			},
			ExpectedError: func(err error) bool {
				return err == nil
			},
			ExpectedGetTransactionError: func(err error) bool {
				return err == ErrTransactionNotFound // in test mode Hetzner immediately cancels the order, and so it returns 404
			},
		},
	}
	// Create test suite
	testSuite := NewTestSuite(
		hetznerProvider,
		getCurrentOfferingsTests,
		orderServerTests,
	)

	// Run tests
	testSuite.RunTests(t, hetznerProvider)
}
