package cellprovider

import (
	"context"
	"fmt"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/onmetal-dev/metal/lib/dnsprovider"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureSecret(ctx context.Context, ctrlClient client.Client, secret corev1.Secret) error {
	existingSecret := &corev1.Secret{}
	err := ctrlClient.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, existingSecret)
	if err == nil {
		secret.ResourceVersion = existingSecret.ResourceVersion
		if err := ctrlClient.Update(ctx, &secret); err != nil {
			return fmt.Errorf("failed to update secret %s: %w", secret.Name, err)
		}
	} else if !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to check if secret %s exists: %w", secret.Name, err)
	} else {
		if err := ctrlClient.Create(ctx, &secret); err != nil {
			return fmt.Errorf("failed to create secret %s: %w", secret.Name, err)
		}
	}
	return nil
}

func ensureClusterIssuer(ctx context.Context, ctrlClient client.Client, issuer cmv1.ClusterIssuer) error {
	existingIssuer := &cmv1.ClusterIssuer{}
	err := ctrlClient.Get(ctx, client.ObjectKey{Name: issuer.Name, Namespace: issuer.Namespace}, existingIssuer)
	if err == nil {
		issuer.ResourceVersion = existingIssuer.ResourceVersion
		if err := ctrlClient.Update(ctx, &issuer); err != nil {
			return fmt.Errorf("failed to update ClusterIssuer %s: %w", issuer.Name, err)
		}
	} else if !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to check if ClusterIssuer %s exists: %w", issuer.Name, err)
	} else {
		if err := ctrlClient.Create(ctx, &issuer); err != nil {
			return fmt.Errorf("failed to create ClusterIssuer %s: %w", issuer.Name, err)
		}
	}
	return nil
}

func ensureLetsEncryptClusterIssuer(ctx context.Context, ctrlClient client.Client, issuer *dnsprovider.CertManagerIssuer) error {
	for _, secret := range issuer.Secrets {
		if err := ensureSecret(ctx, ctrlClient, secret); err != nil {
			return fmt.Errorf("failed to ensure secret %s: %w", secret.Name, err)
		}
	}
	stagingIssuer := cmv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "letsencrypt-staging",
			Namespace: "cert-manager",
		},
		Spec: cmv1.IssuerSpec{
			IssuerConfig: cmv1.IssuerConfig{
				ACME: &cmacme.ACMEIssuer{
					Email:  "certs@onmetal.dev",
					Server: "https://acme-staging-v02.api.letsencrypt.org/directory",
					PrivateKey: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{
							Name: "letsencrypt-staging-private-key",
						},
					},
					Solvers: issuer.Solvers,
				},
			},
		},
	}
	if err := ensureClusterIssuer(ctx, ctrlClient, stagingIssuer); err != nil {
		return fmt.Errorf("failed to ensure staging ClusterIssuer: %w", err)
	}
	productionIssuer := cmv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "letsencrypt-production",
			Namespace: "cert-manager",
		},
		Spec: cmv1.IssuerSpec{
			IssuerConfig: cmv1.IssuerConfig{
				ACME: &cmacme.ACMEIssuer{
					Email:  "certs@onmetal.dev",
					Server: "https://acme-v02.api.letsencrypt.org/directory",
					PrivateKey: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{
							Name: "letsencrypt-production-private-key",
						},
					},
					Solvers: issuer.Solvers,
				},
			},
		},
	}
	if err := ensureClusterIssuer(ctx, ctrlClient, productionIssuer); err != nil {
		return fmt.Errorf("failed to ensure production ClusterIssuer: %w", err)
	}

	return nil
}
