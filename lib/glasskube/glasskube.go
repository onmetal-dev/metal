// package glasskube contains logic that helps with installing and managing glasskyube packages in a cluster.
package glasskube

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	gkv1alpha1 "github.com/glasskube/glasskube/api/v1alpha1"
	gkclient "github.com/glasskube/glasskube/pkg/client"
	"github.com/onmetal-dev/metal/lib/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// EnsureClusterPackageOpts contains options for installing a cluster package.
type EnsureClusterPackageOpts struct {
	Name      string
	Version   string
	Repo      string
	Values    map[string]gkv1alpha1.ValueConfiguration
	Namespace string
}

// EnsureClusterPackage makes sure a package is installed with the given options. It updates the package if it can, otherwise it installs.
func EnsureClusterPackage(ctx context.Context, gkClient gkclient.PackageV1Alpha1Client, opts EnsureClusterPackageOpts) error {
	if opts.Repo == "" {
		opts.Repo = "glasskube"
	}
	log := logger.FromContext(ctx)
	log = log.With("package", opts.Name).With("version", opts.Version).With("repo", opts.Repo)
	var existingPkg gkv1alpha1.ClusterPackage
	update := false
	if err := gkClient.ClusterPackages().Get(ctx, opts.Name, &existingPkg); err == nil {
		if existingPkg.Spec.PackageInfo.Version == opts.Version {
			log.Debug("package already correct")
			return nil // already installed and correct version
		}
		update = true
	} else if !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("error getting package: %v", err)
	}

	var pkg *gkv1alpha1.ClusterPackage
	var err error
	if update {
		log.Info("updating package")
		pkg = &existingPkg
		pkg.Spec.PackageInfo.Version = opts.Version
		if opts.Values != nil {
			pkg.Spec.Values = opts.Values
		}
		err = gkClient.ClusterPackages().Update(ctx, pkg, metav1.UpdateOptions{})
	} else {
		log.Info("installing package")
		pkgBuilder := gkclient.PackageBuilder(opts.Name).WithAutoUpdates(true).WithRepositoryName(opts.Repo).WithVersion(opts.Version)
		if opts.Values != nil {
			pkgBuilder = pkgBuilder.WithValues(opts.Values)
		}
		if opts.Namespace != "" {
			pkgBuilder = pkgBuilder.WithNamespace(opts.Namespace)
		}
		pkg = pkgBuilder.BuildClusterPackage()
		err = gkClient.ClusterPackages().Create(ctx, pkg, metav1.CreateOptions{})
	}
	if err != nil {
		return fmt.Errorf("error creating package: %v", err)
	}
	watcher, err := gkClient.ClusterPackages().Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error watching package: %v", err)
	}
	defer watcher.Stop()
	for event := range watcher.ResultChan() {
		if obj, ok := event.Object.(*gkv1alpha1.ClusterPackage); ok && isSameResource(*obj, *pkg) {
			if event.Type == watch.Added || event.Type == watch.Modified {
				if status := gkclient.GetStatus(obj.GetStatus()); status != nil {
					log.Info("package installation complete", slog.String("status", status.Status))
					return nil
				}
			} else if event.Type == watch.Deleted {
				return fmt.Errorf("created package has been deleted unexpectedly")
			}
		}
	}
	return fmt.Errorf("failed to confirm package installation status")
}

func isSameResource(a, b gkv1alpha1.ClusterPackage) bool {
	return a.GetName() == b.GetName() &&
		a.GroupVersionKind() == b.GroupVersionKind() &&
		a.GetNamespace() == b.GetNamespace()
}
