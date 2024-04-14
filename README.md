# metal

## cluster-api-hetzner

This folder contains directions for bootstrapping a management cluster for the k8s [Cluster API](https://cluster-api.sigs.k8s.io/) in Hetzner.
Using this cluster we can create additional clusters in Hetzner.
It uses [Syself's Hetzner Cluster API Provider](https://github.com/syself/cluster-api-provider-hetzner).

## cluster-api-hivelocity

This folder contains directions for bootstrapping a management cluster for the k8s [Cluster API](https://cluster-api.sigs.k8s.io/) in Hivelocity.
Using this cluster we can create additional clusters in Hivelocity.
It uses [Syself's Hivelocity Cluster API Provider](https://github.com/hivelocity/cluster-api-provider-hivelocity).

## hcloud-tf-setup

This folder contains a first attempt at automating k8s cluster setup in hetzner via the terraform CDK.
This was abandoned in favor of cluster-api.
