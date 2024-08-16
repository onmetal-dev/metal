package main

import (
	"encoding/json"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/ovh/go-ovh/ovh"
)

type Config struct {
	OvhEndpoint     string `envconfig:"OVH_ENDPOINT" required:"true"`
	OvhClientId     string `envconfig:"OVH_CLIENT_ID" required:"true"`
	OvhClientSecret string `envconfig:"OVH_CLIENT_SECRET" required:"true"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

// PartialMe holds the first name of the currently logged-in user.
// Visit https://api.ovh.com/console/#/me#GET for the full definition
type PartialMe struct {
	Firstname string `json:"firstname"`
}

// Instantiate an OVH client and get the firstname of the currently logged-in user.
// Visit https://api.ovh.com/createToken/index.cgi?GET=/me to get your credentials.
func main() {
	//	var me PartialMe

	c := MustLoadConfig()
	client, _ := ovh.NewOAuth2Client(
		c.OvhEndpoint,
		c.OvhClientId,
		c.OvhClientSecret,
	)

	// expireTime := time.Now().Add(10 * time.Minute).Format(time.RFC3339)
	// body := map[string]interface{}{
	// 	"description":   "string",
	// 	"expire":        expireTime,
	// 	"ovhSubsidiary": "US",
	// }

	// type CartResponse struct {
	// 	CartID      string        `json:"cartId"`
	// 	Description string        `json:"description"`
	// 	Expire      string        `json:"expire"`
	// 	Items       []interface{} `json:"items"`
	// 	ReadOnly    bool          `json:"readOnly"`
	// }
	// var res CartResponse
	// if err := client.Post("/v1/order/cart", body, &res); err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Response: %v\n", res)

	type req struct {
		url string
	}
	for _, req := range []req{
		//		{url: "/v1/dedicated/server/region/availabilities"},
		// {url: "/v1/order/cartServiceOption/baremetalServers"},
		// {url: "/v1/order/cart/" + res.CartID},
		//{url: "/v1/order/catalog/formatted/dedicated?ovhSubsidiary=US"},
		{url: "/v1/order/catalog/public/baremetalServers?ovhSubsidiary=US"},
	} {
		var tmp interface{}
		if err := client.Get(req.url, &tmp); err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			d, _ := json.MarshalIndent(tmp, "", "  ")
			fmt.Print(string(d))
		}
	}
}
