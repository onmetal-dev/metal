# cluster-api on Hivelocity

## Bootstrap a management cluster in Hivelocity

Generate an SSH keypair, make a copy of the public key in ~/.ssh/hivelocity.pub since this is where the makefile expects it.

```
ssh-keygen -t ed25519 # follow prompts to put it at ./id_ed25519
cp id_ed25519.pub ~/.ssh/hivelocity.pub

```

Go to https://www.hivelocity.net/pricing?computeType=dedicated,outlet&deploymentTime=instant and buy two dedicated servers.

```
cd cluster-api-provider-hivelocity
export CLUSTER_NAME=hivelocity-management-cluster
export CONTROL_PLANE_MACHINE_COUNT=1
export HIVELOCITY_CONTROL_PLANE_DEVICE_TYPE=hivelocity-management-cluster-control-plane
export HIVELOCITY_WORKER_DEVICE_TYPE=hivelocity-management-cluster-worker
export HIVELOCITY_API_KEY=<hivelocity api key>
export HIVELOCITY_SSH_KEY=caph-ssh-key
export KUBERNETES_VERSION=v1.29.2
export WORKER_MACHINE_COUNT=1
export HIVELOCITY_REGION=NYC1
make tilt-up
```

Press the space bar once prompted to load the Tilt web UI.
Click on the "Create Hivelocity Cluster" button in the top nav.
