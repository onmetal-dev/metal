package cellprovider

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"

	metallbv1beta1 "go.universe.tf/metallb/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createResource(ctx context.Context, ctrlClient client.Client, obj client.Object) error {
	resourceName := reflect.TypeOf(obj).Elem().Name()
	err := ctrlClient.Create(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", resourceName, err)
	}
	return nil
}

func createOrUpdateResource(ctx context.Context, ctrlClient client.Client, obj client.Object) error {
	resourceName := reflect.TypeOf(obj).Elem().Name()
	err := ctrlClient.Create(ctx, obj)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// Get the existing resource
			existing := obj.DeepCopyObject().(client.Object)
			err = ctrlClient.Get(ctx, client.ObjectKeyFromObject(obj), existing)
			if err != nil {
				return fmt.Errorf("failed to get existing %s: %w", resourceName, err)
			}

			// Update the existing resource with new spec and annotations
			existingValue := reflect.ValueOf(existing).Elem()
			newValue := reflect.ValueOf(obj).Elem()
			// only set Spec if the field exists, since some resources like Secrets don't have it
			if specField := existingValue.FieldByName("Spec"); specField.IsValid() {
				specField.Set(newValue.FieldByName("Spec"))
			}
			if dataField := existingValue.FieldByName("Data"); dataField.IsValid() {
				dataField.Set(newValue.FieldByName("Data"))
			}
			existingMeta := existingValue.FieldByName("ObjectMeta")
			newMeta := newValue.FieldByName("ObjectMeta")
			existingMeta.FieldByName("Annotations").Set(newMeta.FieldByName("Annotations"))
			err = ctrlClient.Update(ctx, existing)
			if err != nil {
				return fmt.Errorf("failed to update %s: %w", resourceName, err)
			}
		} else {
			return fmt.Errorf("failed to create %s: %w", resourceName, err)
		}
	}
	return nil
}

func (p *TalosClusterCellProvider) createOrUpdateMetalLBIPAddressPool(ctx context.Context, k8sClient *kubernetes.Clientset, ctrlClient client.Client) error {
	nodeCIDRs, err := getNodeCidrs(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to get node CIDRs: %w", err)
	}

	pool := &metallbv1beta1.IPAddressPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cell-ip-pool",
			Namespace: "metallb-system",
		},
		Spec: metallbv1beta1.IPAddressPoolSpec{
			Addresses: nodeCIDRs,
		},
	}
	if err = createOrUpdateResource(ctx, ctrlClient, pool); err != nil {
		return err
	}
	return p.createOrUpdateMetalLBL2Advertisement(ctx, ctrlClient)
}

func (p *TalosClusterCellProvider) createOrUpdateMetalLBL2Advertisement(ctx context.Context, ctrlClient client.Client) error {
	l2Advertisement := &metallbv1beta1.L2Advertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cell-ip-pool-l2advertisement",
			Namespace: "metallb-system",
		},
		Spec: metallbv1beta1.L2AdvertisementSpec{},
	}
	return createOrUpdateResource(ctx, ctrlClient, l2Advertisement)
}

func getNodeCidrs(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var cidrs []string
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeExternalIP {
				cidrs = append(cidrs, addr.Address+"/32")
				break
			} else if addr.Type == corev1.NodeInternalIP {
				ip := net.ParseIP(addr.Address)
				// for some reason the public IP ends up in the internal ip list
				if ip != nil && !ip.IsPrivate() {
					cidrs = append(cidrs, addr.Address+"/32")
					break
				}
			}
		}
	}

	return cidrs, nil
}
