// package dnsprovider contains logic to manage DNS providers in different providers
package dnsprovider

import (
	"context"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	"github.com/onmetal-dev/metal/lib/glasskube"
	corev1 "k8s.io/api/core/v1"
)

type CertManagerIssuer struct {
	Secrets []corev1.Secret
	Solvers []cmacme.ACMEChallengeSolver
}

type ExternalDnsSetup struct {
	Secrets        []corev1.Secret
	GkPkgsToEnsure []glasskube.EnsureClusterPackageOpts
}

type DNSProvider interface {
	Domain() (string, error)
	FindOrCreateARecord(ctx context.Context, zoneID, recordName, recordContent string) error

	// CertManagerIssuer returns configuration to allow setting up a cert-manager ClusterIssuer that can be used to create SSL certificates using the underlying DNS provider.
	CertManagerIssuer() (*CertManagerIssuer, error)

	// ExternalDnsSetup returns directions for how to install external-dns pointing to the underlying DNS provider.
	ExternalDnsSetup() (*ExternalDnsSetup, error)
}
