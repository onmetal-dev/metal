package dnsprovider

import (
	"context"
	"fmt"
	"strings"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cloudflare/cloudflare-go"
	"github.com/onmetal-dev/metal/lib/glasskube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CloudflareDNSProvider struct {
	api      *cloudflare.API
	zoneId   string
	zoneName *string
}

var _ DNSProvider = &CloudflareDNSProvider{}

type CloudflareDNSProviderOption func(*CloudflareDNSProvider)

func WithApi(api *cloudflare.API) CloudflareDNSProviderOption {
	return func(p *CloudflareDNSProvider) {
		p.api = api
	}
}

func WithZoneId(zoneId string) CloudflareDNSProviderOption {
	return func(p *CloudflareDNSProvider) {
		p.zoneId = zoneId
	}
}

func NewCloudflareDNSProvider(opts ...CloudflareDNSProviderOption) (*CloudflareDNSProvider, error) {
	provider := &CloudflareDNSProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	var errs []string
	if provider.api == nil {
		errs = append(errs, "must provide a valid Cloudflare API")
	}
	if provider.zoneId == "" {
		errs = append(errs, "must provide a valid zoneId")
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("errors: %v", strings.Join(errs, ", "))
	}
	return provider, nil
}

func (p *CloudflareDNSProvider) Domain() (string, error) {
	if p.zoneName == nil {
		zone, err := p.api.ZoneDetails(context.Background(), p.zoneId)
		if err != nil {
			return "", err
		}
		p.zoneName = &zone.Name
	}
	return *p.zoneName, nil
}

func (p *CloudflareDNSProvider) FindOrCreateARecord(ctx context.Context, zoneID, recordName, recordContent string) error {
	domain, err := p.Domain()
	if err != nil {
		return err
	}
	dnsRecords, _, err := p.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: fmt.Sprintf("%s.%s", recordName, domain),
	})
	if err != nil {
		return fmt.Errorf("error listing DNS records: %v", err)
	} else if len(dnsRecords) > 0 {
		if dnsRecords[0].Content != recordContent {
			return fmt.Errorf("existing record content mismatch: %s != %s", dnsRecords[0].Content, recordContent)
		}
		return nil
	}
	if _, err = p.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    recordName,
		Content: recordContent,
	}); err != nil {
		return fmt.Errorf("error creating A record: %v", err)
	}
	return nil
}

func (p *CloudflareDNSProvider) CertManagerIssuer() (*CertManagerIssuer, error) {
	return &CertManagerIssuer{
		Secrets: []corev1.Secret{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cloudflare-api-token",
					Namespace: "cert-manager",
				},
				StringData: map[string]string{
					"api-token": p.api.APIToken,
				},
			},
		},
		Solvers: []cmacme.ACMEChallengeSolver{
			{
				DNS01: &cmacme.ACMEChallengeSolverDNS01{
					Cloudflare: &cmacme.ACMEIssuerDNS01ProviderCloudflare{
						Email: "rgarcia2009@gmail.com",
						APIToken: &cmmeta.SecretKeySelector{
							LocalObjectReference: cmmeta.LocalObjectReference{
								Name: "cloudflare-api-token",
							},
							Key: "api-token",
						},
					},
				},
			},
		},
	}, nil
}

func (p *CloudflareDNSProvider) ExternalDnsSetup() (*ExternalDnsSetup, error) {
	return &ExternalDnsSetup{
		Secrets: []corev1.Secret{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cloudflare-api-token",
					Namespace: "external-dns",
				},
				StringData: map[string]string{
					"api-token": p.api.APIToken,
				},
			},
		},
		GkPkgsToEnsure: []glasskube.EnsureClusterPackageOpts{
			{Name: "external-dns-cloudflare", Version: "v1.15.0+1", Repo: "metal", Namespace: "external-dns"},
		},
	}, nil
}
