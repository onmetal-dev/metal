package serverprovider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/floshodan/hrobot-go/hrobot"
)

type Hetzner struct {
	client                   *hrobot.Client
	authorizedKeyFingerprint string
	// testMode sends test:true to the server ordering endpoint in order to not actually buy a server
	testMode bool
}

var _ ServerProvider = &Hetzner{}

type HetznerOption func(*Hetzner) error

func WithHrobotClient(client *hrobot.Client) HetznerOption {
	return func(h *Hetzner) error {
		if client == nil {
			return errors.New("hrobot client cannot be nil")
		}
		h.client = client
		return nil
	}
}

func WithAuthorizedKeyFingerprint(fingerprint string) HetznerOption {
	return func(h *Hetzner) error {
		h.authorizedKeyFingerprint = fingerprint
		return nil
	}
}

func WithTestMode(testMode bool) HetznerOption {
	return func(h *Hetzner) error {
		h.testMode = testMode
		return nil
	}
}

func NewHetzner(opts ...HetznerOption) (*Hetzner, error) {
	h := &Hetzner{}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	var errs []string
	if h.client == nil {
		errs = append(errs, "hrobot client is required")
	}
	if h.authorizedKeyFingerprint == "" {
		errs = append(errs, "authorized key fingerprint is required")
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", "))
	}
	return h, nil
}

// Add this helper function at the end of the file
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

const HetznerSlug = "hetzner"

func (h *Hetzner) Slug() string {
	return HetznerSlug
}

func (h *Hetzner) GetCurrentOfferings() ([]Offering, error) {
	products, _, err := h.client.Order.ProductList(context.TODO(), &hrobot.OrderServerListOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get product list: %w", err)
	}

	offerings := make([]Offering, len(products))
	for i, product := range products {
		prices := make([]Price, len(product.Prices))
		for j, price := range product.Prices {
			prices[j] = Price{
				Currency:      "EUR",
				Location:      price.Location,
				AmountMonthly: parseFloat(price.Price.Gross),
				AmountSetup:   parseFloat(price.PriceSetup.Gross),
			}
		}
		addons := make([]Addon, len(product.OrderableAddons))
		for j, addon := range product.OrderableAddons {
			var prices []Price
			if pricesSlice, ok := addon.Prices.([]interface{}); ok {
				for _, priceInterface := range pricesSlice {
					if priceMap, ok := priceInterface.(map[string]interface{}); ok {
						location, _ := priceMap["location"].(string)
						price, priceOk := priceMap["price"].(map[string]interface{})
						priceSetup, setupOk := priceMap["price_setup"].(map[string]interface{})

						if priceOk && setupOk {
							grossPrice, _ := price["gross"].(string)
							grossSetup, _ := priceSetup["gross"].(string)

							prices = append(prices, Price{
								Currency:      "EUR",
								Location:      location,
								AmountMonthly: parseFloat(grossPrice),
								AmountSetup:   parseFloat(grossSetup),
							})
						}
					}
				}
			}

			addons[j] = Addon{
				Id:     addon.ID,
				Name:   addon.Name,
				Min:    addon.Min,
				Max:    addon.Max,
				Prices: prices,
			}
		}
		offerings[i] = Offering{
			Id:          product.ID,
			Name:        product.Name,
			Description: strings.Join(product.Description, ", "),
			Locations:   product.Location,
			Prices:      prices,
			Addons:      addons,
		}
	}

	return offerings, nil
}

func (h *Hetzner) OrderServer(order Order) (Transaction, error) {
	orderOpts := &hrobot.OrderServerOpts{
		Product_ID:     order.OfferingId,
		Dist:           "Rescue system",
		Authorized_Key: h.authorizedKeyFingerprint,
		Location:       order.LocationId,
		Addons:         "primary_ipv4", // eventually this might include other things like extra storage, etc., but for now just default to the basic ipv4 addon
		Test:           h.testMode,
	}
	serverOrder, _, err := h.client.Order.OrderServer(context.Background(), orderOpts)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to order server: %w", err)
	}
	return Transaction{
		Id:         serverOrder.ID,
		Status:     convertHetznerTxStatus(serverOrder.Status),
		OfferingId: serverOrder.Product.ID,
		Location:   serverOrder.Product.Location,
		AddonIds:   convertAddonIdsToStringSlice(serverOrder.Addons),
		ServerId:   convertServerNumber(serverOrder.ServerNumber),
	}, nil
}

func (h *Hetzner) GetTransaction(id string) (Transaction, error) {
	serverTransaction, _, err := h.client.Order.GetServerTransactionById(context.TODO(), id)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, fmt.Errorf("failed to get transaction: %w", err)
	}
	return Transaction{
		Id:         id,
		Status:     convertHetznerTxStatus(serverTransaction.Status),
		ServerId:   convertServerNumber(serverTransaction.ServerNumber),
		OfferingId: serverTransaction.Product.ID,
		Location:   serverTransaction.Product.Location,
		AddonIds:   convertAddonIdsToStringSlice(serverTransaction.Addons),
	}, nil
}

func convertServerNumber(serverNumber interface{}) string {
	if serverId, ok := serverNumber.(float64); ok {
		return fmt.Sprintf("%d", int(serverId))
	}
	return ""
}

func convertAddonIdsToStringSlice(addons []interface{}) []string {
	addonIds := make([]string, 0, len(addons))
	for _, addon := range addons {
		if addonStr, ok := addon.(string); ok {
			addonIds = append(addonIds, addonStr)
		}
	}
	return addonIds
}

func convertHetznerTxStatus(status string) TransactionStatus {
	switch status {
	case "ready":
		return TransactionStatusCompleted
	case "in process":
		return TransactionStatusPending
	case "cancelled":
		return TransactionStatusCanceled
	default:
		return TransactionStatusPending // default to pending if status is unknown
	}
}

func (h *Hetzner) GetServer(serverId string) (Server, error) {
	singleServer, _, err := h.client.Server.GetServerById(context.Background(), serverId)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return Server{}, ErrServerNotFound
		}
		return Server{}, fmt.Errorf("failed to get server: %w", err)
	}

	return Server{
		Id:            fmt.Sprintf("%d", singleServer.ServerNumber),
		ProviderSlug:  h.Slug(),
		OfferingId:    singleServer.Product,
		Location:      singleServer.Dc,
		Status:        convertHetznerServerStatus(singleServer.Status),
		StatusDetails: "",
		Ipv4:          singleServer.ServerIP,
		Ipv6:          singleServer.ServerIpv6Net,
	}, nil
}

func convertHetznerServerStatus(status string) ServerStatus {
	switch status {
	case "ready":
		return ServerStatusRunning
	case "in process":
		return ServerStatusPending
	default:
		return ServerStatusPending
	}
}
