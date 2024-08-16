package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/floshodan/hrobot-go/hrobot"
	"github.com/kelseyhightower/envconfig"
	"github.com/onmetal-dev/metal/lib/talosprovider"
	"github.com/samber/lo"
	"github.com/siderolabs/talos/pkg/machinery/client"
)

type Config struct {
	HetznerToken      string `envconfig:"HETZNER_TOKEN" required:"true"`
	SshKeyBase64      string `envconfig:"SSH_KEY_BASE64" required:"true"`
	SshKeyPassword    string `envconfig:"SSH_KEY_PASSWORD" required:"true"`
	SshKeyFingerprint string `envconfig:"SSH_KEY_FINGERPRINT" required:"true"`
	ServerId          string `envconfig:"SERVER_ID" required:"true"`
	ServerIp          string `envconfig:"SERVER_IP" required:"true"`
	TransactionId     string `envconfig:"TRANSACTION_ID" required:"true"`
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

func main() {
	ctx := context.Background()
	c := MustLoadConfig()
	hrobotClient := hrobot.NewClient(hrobot.WithToken(c.HetznerToken))

	if lo.Contains(os.Args, "list-sshkeys") {
		ssh, _, _ := hrobotClient.SSHKey.List(ctx)
		for _, rec := range ssh {
			fmt.Println(rec.Name)
			fmt.Println(rec.Fingerprint)
			fmt.Println(rec.Data)
		}
	}

	if lo.Contains(os.Args, "list-products") {
		products, _, err := hrobotClient.Order.ProductList(ctx, &hrobot.OrderServerListOpts{
			Location: "hel1",
		})
		if err != nil {
			panic(err)
		}
		file, err := os.Create("cmd/sandbox-hetzner/products.json")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(products); err != nil {
			panic(err)
		}
	}

	if lo.Contains(os.Args, "order-server") {
		order := &hrobot.OrderServerOpts{
			Product_ID:     "AX102",
			Dist:           "Rescue system",
			Authorized_Key: c.SshKeyFingerprint,
			Location:       "HEL1",
			Addons:         "primary_ipv4",
			Test:           true,
		}
		tx, _, err := hrobotClient.Order.OrderServer(ctx, order)
		if err != nil {
			panic(err)
		}
		spew.Dump(tx)
	}

	if lo.Contains(os.Args, "check-transaction") {
		tx, _, err := hrobotClient.Order.GetServerTransactionById(ctx, c.TransactionId)
		if err != nil {
			panic(err)
		}
		bs, _ := json.MarshalIndent(tx, "", "  ")
		fmt.Println(string(bs))
	}

	if lo.Contains(os.Args, "check-server") {
		server, _, err := hrobotClient.Server.GetServerById(ctx, c.ServerId)
		if err != nil {
			panic(err)
		}
		bs, _ := json.MarshalIndent(server, "", "  ")
		fmt.Println(string(bs))
	}

	if lo.Contains(os.Args, "check-reset") {
		reset, _, err := hrobotClient.Reset.GetResetByServernumber(ctx, c.ServerId)
		if err != nil {
			panic(err)
		}
		bs, _ := json.MarshalIndent(reset, "", "  ")
		fmt.Println(string(bs))
	}

	if lo.Contains(os.Args, "execute-rescue-and-reset") {
		fmt.Println(c.SshKeyFingerprint)
		rescue, _, err := hrobotClient.Boot.ActivateRescue(ctx, c.ServerId, &hrobot.RescueOpts{
			OS:             "linux",
			Authorized_Key: c.SshKeyFingerprint,
			Keyboard:       "us",
		})
		if err != nil {
			panic(err)
		}
		bs, _ := json.MarshalIndent(rescue, "", "  ")
		fmt.Println(string(bs))
		reset, _, err := hrobotClient.Reset.ExecuteReset(ctx, c.ServerId, "hw")
		if err != nil {
			panic(err)
		}
		bs, _ = json.MarshalIndent(reset, "", "  ")
		fmt.Println(string(bs))
		log.Println("waiting for server to accept ssh connections")
		for {
			time.Sleep(5 * time.Second)
			conn, err := net.DialTimeout("tcp", c.ServerIp+":22", 5*time.Second)
			if err == nil {
				conn.Close()
				break
			}
		}
		log.Println("Server is up and running")
	}

	if lo.Contains(os.Args, "install-talos") {
		s := talosprovider.Server{
			Id:                    c.ServerId,
			Ip:                    c.ServerIp,
			Username:              "root",
			SshKeyPrivateBase64:   c.SshKeyBase64,
			SshKeyPrivatePassword: c.SshKeyPassword,
			SshKeyFingerprint:     c.SshKeyFingerprint,
		}
		hetzner, err := talosprovider.NewHetznerProvider(talosprovider.WithClient(hrobotClient))
		if err != nil {
			panic(err)
		}
		if err := hetzner.Install(ctx, s); err != nil {
			panic(err)
		}
		c, err := client.New(ctx, client.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), client.WithEndpoints(s.Ip))
		if err != nil {
			panic(err)
		}
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		disks, err := c.Disks(ctxWithTimeout)
		if err != nil {
			panic(err)
		}
		spew.Dump(disks)
	}

	//new_key, resp, _ := client.SSHKey.Create(context.Background(), data)

	//fmt.Println(new_key.Name, resp.Status)

}
