package cellprovider

import (
	"context"
	"encoding/base64"
	"fmt"

	gkclient "github.com/glasskube/glasskube/pkg/client"
	"github.com/onmetal-dev/metal/lib/glasskube"
	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const registryNamespace = "registry"
const registryPlaintextSecretName = "registry-auth-plaintext"
const dockerconfigjsonSecretName = "registry-dockerconfigjson"

func cellRegistryHostname(cellId string) string {
	return fmt.Sprintf("registry.%s", cellHostname(cellId))
}

type cellRegistryCredentials struct {
	Username string
	Password string
}

// cellRegistryCredentials returns the username and password for the registry.
// The username and password are stored in a secret in the registry namespace.
func (p *TalosClusterCellProvider) cellRegistryCredentials(ctx context.Context, ctrlClient client.Client) (*cellRegistryCredentials, error) {
	plaintextSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryPlaintextSecretName,
			Namespace: registryNamespace,
		},
	}
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: registryPlaintextSecretName, Namespace: registryNamespace}, plaintextSecret); err != nil {
		return nil, fmt.Errorf("failed to get plaintext secret: %w", err)
	}
	return &cellRegistryCredentials{
		Username: string(plaintextSecret.Data["username"]),
		Password: string(plaintextSecret.Data["password"]),
	}, nil
}

// copyImagePullSecretToNamespace copies the dockerconfigjson secret from the registry namespace to the given namespace
func copyImagePullSecretToNamespace(ctx context.Context, ctrlClient client.Client, srcNamespace, dstNamespace string) error {
	srcSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerconfigjsonSecretName,
			Namespace: srcNamespace,
		},
	}
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: dockerconfigjsonSecretName, Namespace: srcNamespace}, srcSecret); err != nil {
		return fmt.Errorf("failed to get dockerconfigjson secret: %w", err)
	}

	// create a new secret object and copy the necessary fields
	dstSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerconfigjsonSecretName,
			Namespace: dstNamespace,
		},
		Data: srcSecret.Data,
		Type: srcSecret.Type,
	}

	if err := createOrUpdateResource(ctx, ctrlClient, dstSecret); err != nil {
		return fmt.Errorf("failed to create or update dockerconfigjson secret: %w", err)
	}
	return nil
}

// createOrUpdateRegistry creates or updates the private docker registry for the cluster
func (p *TalosClusterCellProvider) createOrUpdateRegistry(ctx context.Context, k8sClient kubernetes.Interface, ctrlClient client.Client, gkClient gkclient.PackageV1Alpha1Client, cellId string) error {
	if err := ensureNamespaceWithLabels(ctx, k8sClient, registryNamespace, podSecurityLabels); err != nil {
		return fmt.Errorf("error ensuring %s namespace: %w", registryNamespace, err)
	}

	// create a shared fs pvc for the registry. This is taken from the canonical example in the rook docs:
	// https://rook.io/docs/rook/latest/Storage-Configuration/Shared-Filesystem-CephFS/filesystem-storage/#consume-the-shared-filesystem-k8s-registry-sample
	pvcName := "registry-pvc"
	pvcSize := "10Gi"
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: registryNamespace,
		},
	}
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: pvcName, Namespace: registryNamespace}, pvc); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get PVC: %w", err)
		}
		pvc.Spec = corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(pvcSize),
				},
			},
			StorageClassName: lo.ToPtr("rook-cephfs"),
		}
		if err := createResource(ctx, ctrlClient, pvc); err != nil {
			return fmt.Errorf("failed to create or update PersistentVolumeClaim: %w", err)
		}
	}

	// Find / create the secret containing the basic auth credentials for the registry
	// This secret is just for record keeping / reference for when we need to use the registry.
	// It is not the secret used by the registry itself. The registry uses a secret that contains
	// the bcrypt'd password, which we will create separately.
	plaintextSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryPlaintextSecretName,
			Namespace: registryNamespace,
		},
	}
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: registryPlaintextSecretName, Namespace: registryNamespace}, plaintextSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get plaintext secret: %w", err)
		}
		// generate a random username and password
		username := randomAlphaNumericString(16)
		password := randomAlphaNumericString(32)
		// create the secret. This should be the equivalent of kubectl create secret generic registry-basic-auth --from-literal=username=<username> --from-literal=password=<password>
		plaintextSecret.Data = map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		}
		if err := createResource(ctx, ctrlClient, plaintextSecret); err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	// store a registry-dockerconfigjson secret for pods to use
	creds, err := p.cellRegistryCredentials(ctx, ctrlClient)
	if err != nil {
		return fmt.Errorf("failed to get registry credentials: %w", err)
	}
	dockerconfigjsonSecret := &corev1.Secret{
		Type: corev1.SecretTypeDockerConfigJson,
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerconfigjsonSecretName,
			Namespace: registryNamespace,
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(fmt.Sprintf(`{"auths":{"%s":{"auth":"%s"}}}`,
				cellRegistryHostname(cellId),
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", creds.Username, creds.Password))),
			)),
		},
	}
	if err := createOrUpdateResource(ctx, ctrlClient, dockerconfigjsonSecret); err != nil {
		return fmt.Errorf("failed to create dockerconfigjson secret: %w", err)
	}

	// create the bcrypt'd secret for the registry to use
	bcryptedSecretName := "registry-auth"
	bcryptedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bcryptedSecretName,
			Namespace: registryNamespace,
		},
	}
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: bcryptedSecretName, Namespace: registryNamespace}, bcryptedSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get bcrypted secret: %w", err)
		}
		username := string(plaintextSecret.Data["username"])
		password := string(plaintextSecret.Data["password"])
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		bcryptedSecret.Data = map[string][]byte{
			"htpasswd": []byte(fmt.Sprintf("%s:%s", username, hashedPassword)),
		}
		if err := createResource(ctx, ctrlClient, bcryptedSecret); err != nil {
			return fmt.Errorf("failed to create bcrypted secret: %w", err)
		}
	}

	// final secret is the config for the registry itself
	registryConfigSecretName := "registry-config"
	registryConfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryConfigSecretName,
			Namespace: registryNamespace,
		},
	}
	if err := ctrlClient.Get(ctx, client.ObjectKey{Name: registryConfigSecretName, Namespace: registryNamespace}, registryConfigSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get registry config secret: %w", err)
		}
		cookieSecret := randomAlphaNumericString(32)
		registryConfigSecret.Data = map[string][]byte{
			"config.yml": []byte(fmt.Sprintf(`
version: 0.1
http:
  secret: %s
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
auth:
  htpasswd:
    realm: basic-realm
    path: /etc/docker/auth/htpasswd
log:
  level: debug
  fields:
    service: registry
storage:
  filesystem:
    rootdirectory: /var/lib/registry
  delete:
    enabled: true
  maintenance:
    uploadpurging:
      enabled: true
      age: 168h
      interval: 24h
      dryrun: false
    readonly:
      enabled: false
`, cookieSecret)),
		}
		if err := createResource(ctx, ctrlClient, registryConfigSecret); err != nil {
			return fmt.Errorf("failed to create registry config secret: %w", err)
		}
	}

	// install the glasskube package
	if err := glasskube.EnsureClusterPackage(ctx, gkClient, glasskube.EnsureClusterPackageOpts{
		Name:      "registry",
		Repo:      "metal",
		Version:   "v22.4.11+1",
		Namespace: registryNamespace,
	}); err != nil {
		return fmt.Errorf("failed to ensure cluster package: %w", err)
	}

	// final step is to make sure there's an http route for the registry
	routeName := "registry-route"
	route := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Namespace: registryNamespace,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{
					{
						Kind:      lo.ToPtr(gatewayv1.Kind("Gateway")),
						Name:      "gateway",
						Namespace: lo.ToPtr(gatewayv1.Namespace("gateway")),
						Port:      lo.ToPtr(gatewayv1.PortNumber(443)),
					},
				},
			},
			Hostnames: []gatewayv1.Hostname{
				gatewayv1.Hostname(cellRegistryHostname(cellId)),
			},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  lo.ToPtr(gatewayv1.PathMatchPathPrefix),
								Value: lo.ToPtr("/"),
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Kind: lo.ToPtr(gatewayv1.Kind("Service")),
									Name: "docker-registry",
									Port: lo.ToPtr(gatewayv1.PortNumber(5000)),
								},
							},
						},
					},
				},
			},
		},
	}
	if err := createOrUpdateResource(ctx, ctrlClient, route); err != nil {
		return fmt.Errorf("failed to create or update HTTPRoute: %w", err)
	}

	return nil
}

func randomAlphaNumericString(length int) string {
	return lo.RandomString(length, lo.AlphanumericCharset)
}
