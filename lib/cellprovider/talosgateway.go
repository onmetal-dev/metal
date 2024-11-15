package cellprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/samber/lo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const gatewayNamespace = "gateway"
const gatewayName = "gateway"

func cellHostname(cellId string) string {
	// if you are strict about hostnames (and k8s is) they cannot contain underscores
	cellIdForHostname := strings.ReplaceAll(cellId, "_", "-")
	return fmt.Sprintf("%s.up.onmetal.run", cellIdForHostname)
}

// createOrUpdateGateway creates or updates the gateway object in the gateway namespace.
// It assumes some things:
// - istio is installed
// - cert-manager is installed and a "letsencrypt-production" issuer is available
// - external-dns is installed
func (p *TalosClusterCellProvider) createOrUpdateGateway(ctx context.Context, k8sClient kubernetes.Interface, ctrlClient client.Client, cellId string) error {
	log := logger.FromContext(ctx)
	if err := ensureNamespaceWithLabels(ctx, k8sClient, gatewayNamespace, podSecurityLabels); err != nil {
		return fmt.Errorf("failed to ensure gateway namespace: %w", err)
	}
	wildcardDomain := fmt.Sprintf("*.%s", cellHostname(cellId))

	// Create Certificate
	certificateName := "gateway-external-wildcard-certificate"
	certificate := &cmv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certificateName,
			Namespace: gatewayNamespace,
		},
	}
	create := false
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: certificateName, Namespace: gatewayNamespace}, certificate); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get certificate: %w", err)
		}
		create = true
		log.Info("creating certificate", "name", certificateName, "namespace", gatewayNamespace, "dnsNames", wildcardDomain)
		certificate.Spec = cmv1.CertificateSpec{
			DNSNames: []string{wildcardDomain},
			IssuerRef: cmmeta.ObjectReference{
				Kind: "ClusterIssuer",
				Name: "letsencrypt-production",
			},
			SecretName: certificateName,
		}
		if err := createResource(ctx, ctrlClient, certificate); err != nil {
			return fmt.Errorf("failed to create or update Certificate: %w", err)
		}
	}

	// wait for the Certificate to be ready
	if err := waitForCertificateReady(ctx, ctrlClient, certificate.Name, certificate.Namespace, 180*time.Second); err != nil {
		return fmt.Errorf("timeout waiting for Certificate to be ready: %w", err)
	}
	if create {
		log.Info("certificate is ready", "name", certificateName, "namespace", gatewayNamespace)
	}

	gateway := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatewayName,
			Namespace: gatewayNamespace,
			Annotations: map[string]string{
				"external-dns.alpha.kubernetes.io/hostname": wildcardDomain,
			},
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: gatewayv1.ObjectName("istio"),
			Listeners: []gatewayv1.Listener{
				{
					Name:     gatewayv1.SectionName("http-subdomains-gateway"),
					Protocol: gatewayv1.HTTPProtocolType,
					Port:     gatewayv1.PortNumber(80),
					Hostname: (*gatewayv1.Hostname)(&wildcardDomain),
					AllowedRoutes: &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: lo.ToPtr(gatewayv1.FromNamespaces("All")),
						},
					},
				},
				{
					Name:     gatewayv1.SectionName("https-subdomains-gateway"),
					Protocol: gatewayv1.HTTPSProtocolType,
					Port:     gatewayv1.PortNumber(443),
					Hostname: lo.ToPtr(gatewayv1.Hostname(wildcardDomain)),
					TLS: &gatewayv1.GatewayTLSConfig{
						Mode: lo.ToPtr(gatewayv1.TLSModeTerminate),
						CertificateRefs: []gatewayv1.SecretObjectReference{
							{
								Kind:      lo.ToPtr(gatewayv1.Kind("Secret")),
								Name:      gatewayv1.ObjectName(certificateName),
								Namespace: lo.ToPtr(gatewayv1.Namespace(gatewayNamespace)),
							},
						},
					},
					AllowedRoutes: &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: lo.ToPtr(gatewayv1.FromNamespaces("All")),
						},
					},
				},
			},
		},
	}

	return createOrUpdateResource(ctx, ctrlClient, gateway)
}

// waitForCertificateReady waits for a Certificate to be in the Ready condition
func waitForCertificateReady(ctx context.Context, c client.Client, name, namespace string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		var cert cmv1.Certificate
		err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &cert)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil // Certificate not found, keep polling
			}
			return false, err // Unexpected error, stop polling
		}
		for _, condition := range cert.Status.Conditions {
			if condition.Type == cmv1.CertificateConditionReady {
				if condition.Status == "True" {
					return true, nil // Certificate is ready
				}
				return false, nil // Certificate is not ready, keep polling
			}
		}
		return false, nil // Ready condition not found, keep polling
	})
}
