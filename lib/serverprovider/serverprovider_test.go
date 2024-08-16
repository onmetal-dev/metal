package serverprovider

import "testing"

// TestSuite represents a set of tests for a ServerProvider implementation
type TestSuite struct {
	GetCurrentOfferingsTests []GetCurrentOfferingsTest
	OrderServerTests         []OrderServerTest
}

type GetCurrentOfferingsTest struct {
	Name              string
	ExpectedOfferings func([]Offering) bool
	ExpectedError     func(error) bool
}

type OrderServerTest struct {
	Name                        string
	Input                       Order
	ExpectedTransaction         func(Transaction) bool
	ExpectedError               func(error) bool
	ExpectedGetTransaction      func(Transaction) bool
	ExpectedGetTransactionError func(error) bool
}

// NewTestSuite returns a generic test suite for the ServerProvider interface
func NewTestSuite(
	provider ServerProvider,
	getCurrentOfferingsTests []GetCurrentOfferingsTest,
	orderServerTests []OrderServerTest,
) TestSuite {
	return TestSuite{
		GetCurrentOfferingsTests: getCurrentOfferingsTests,
		OrderServerTests:         orderServerTests,
	}
}

// RunTests executes all tests in the test suite
func (ts TestSuite) RunTests(t *testing.T, provider ServerProvider) {
	t.Run("GetCurrentOfferings", func(t *testing.T) {
		for _, test := range ts.GetCurrentOfferingsTests {
			t.Run(test.Name, func(t *testing.T) {
				offerings, err := provider.GetCurrentOfferings()
				if !test.ExpectedError(err) {
					t.Errorf("Unexpected error: %v", err)
				}
				if !test.ExpectedOfferings(offerings) {
					t.Errorf("Unexpected offerings: %v", offerings)
				}
			})
		}
	})

	t.Run("OrderServer", func(t *testing.T) {
		for _, test := range ts.OrderServerTests {
			t.Run(test.Name, func(t *testing.T) {
				transaction, err := provider.OrderServer(test.Input)
				if !test.ExpectedError(err) {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !test.ExpectedTransaction(transaction) {
					t.Fatalf("Unexpected transaction: %v", transaction)
				}

				if test.ExpectedGetTransaction != nil || test.ExpectedGetTransactionError != nil {
					gotTransaction, err := provider.GetTransaction(transaction.Id)

					if test.ExpectedGetTransactionError != nil {
						if !test.ExpectedGetTransactionError(err) {
							t.Fatalf("Unexpected error in GetTransaction: %v", err)
						}
					} else if err != nil {
						t.Fatalf("Unexpected error in GetTransaction: %v", err)
					}

					if test.ExpectedGetTransaction != nil {
						if !test.ExpectedGetTransaction(gotTransaction) {
							t.Fatalf("Unexpected transaction from GetTransaction: %v", gotTransaction)
						}
					}
				}
			})
		}
	})

	t.Run("GetTransaction with garbage ID", func(t *testing.T) {
		_, err := provider.GetTransaction("garbage_id")
		if err != ErrTransactionNotFound {
			t.Errorf("Expected ErrTransactionNotFound, got %v", err)
		}
	})
}
