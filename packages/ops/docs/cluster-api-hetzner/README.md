# cluster-api on Hetzner

## Bootstrap a management cluster in Hetzner

Unclear if this management cluster will be a singleton (used for all users), or if each user will have a management cluster.
Let's assume that it's a singleton (would be nice to not have to manage the management clusters, ha) and proceed:

1. Create the bootstrap cluster locally: `kind create cluster`
1. Initialize the management cluster: `clusterctl init --core cluster-api --bootstrap kubeadm --control-plane kubeadm --infrastructure hetzner` (from docs: https://github.com/syself/cluster-api-provider-hetzner/blob/main/docs/topics/preparation.md)
1. Create a new project in hetzner, e.g. "caph"
1. Generate an SSH key file (private and public), e.g. `id_ed25519{,.pub}` (`ssh-keygen -t ed25519`) and add an ssh key to the project under Security -> SSH Keys. Name it something like `caph-ssh-key`
1. Generate an API token with read and write access under Security -> API Tokens. Name it something like `caph-api-token`.
1. Generate a webservice user password in Hetzner Robot: https://robot.hetzner.com/preferences/index
1. Set up env

   ```
   export HCLOUD_TOKEN='<hcloud token>' \
   export HCLOUD_SSH_KEY="caph-ssh-key" \
   export HETZNER_ROBOT_USER='<hetzner robot user>' \
   export HETZNER_ROBOT_PASSWORD='<hetzner robot password>' \
   export HETZNER_SSH_PUB_PATH=$(pwd)/id_ed25519.pub \
   export HETZNER_SSH_PRIV_PATH=$(pwd)/id_ed25519 \
   export CLUSTER_NAME="management-cluster" \
   export HCLOUD_REGION="ash" \
   export CONTROL_PLANE_MACHINE_COUNT=1 \
   export WORKER_MACHINE_COUNT=3 \
   export KUBERNETES_VERSION=1.28.4 \
   export HCLOUD_CONTROL_PLANE_MACHINE_TYPE=cpx31 \
   export HCLOUD_WORKER_MACHINE_TYPE=cpx31
   ```

   Check that your limits (https://console.hetzner.cloud/limits) support the machine counts above.
   Regions: https://docs.hetzner.com/cloud/general/locations/
   Instance types: https://www.hetzner.com/cloud#pricing

1. Add HCLOUD_TOKEN etc. as a `hetzner` secret to the bootstrap cluster:
   ```
   kubectl create secret generic hetzner --from-literal=hcloud=$HCLOUD_TOKEN --from-literal=robot-user=$HETZNER_ROBOT_USER --from-literal=robot-password=$HETZNER_ROBOT_PASSWORD
   kubectl create secret generic robot-ssh --from-literal=sshkey-name=cluster --from-file=ssh-privatekey=$HETZNER_SSH_PRIV_PATH --from-file=ssh-publickey=$HETZNER_SSH_PUB_PATH
   # Patch the created secret so it is automatically moved to the target cluster later.
   kubectl patch secret hetzner -p '{"metadata":{"labels":{"clusterctl.cluster.x-k8s.io/move":""}}}'
   kubectl patch secret robot-ssh -p '{"metadata":{"labels":{"clusterctl.cluster.x-k8s.io/move":""}}}'
   ```
1. Generate the cluster yaml:

   ```
   clusterctl generate cluster $CLUSTER_NAME --flavor hcloud > cluster.yaml
   ```

   Can also specify a template, e.g. private network version: https://github.com/syself/cluster-api-provider-hetzner/tree/55aef633283cec17f993126d1fdfdd338f2915c9/templates/cluster-templates

   For doing private network in the US:

   ```
    clusterctl generate cluster $CLUSTER_NAME --flavor hcloud > cluster-unpatched.yaml
    envsubst < hcloudNetwork_patch_envsubst.yaml > hcloudNetwork_patch.yaml
    kustomize build . > cluster.yaml
   ```

1. Apply: `kubectl apply -f cluster.yaml`
1. Check status: `kubectl get cluster`
1. View cluster resources: `clusterctl describe cluster $CLUSTER_NAME`
1. Wait for the control plane node to come up (`kubectl get kubeadmcontrolplane`) and then pull down the kubeconfig:

   ```
   export CAPH_WORKER_CLUSTER_KUBECONFIG=/tmp/workload-kubeconfig
   clusterctl get kubeconfig $CLUSTER_NAME > $CAPH_WORKER_CLUSTER_KUBECONFIG
   ```

1. Install a CNI solution (Cilium):

   ```
   helm repo add cilium https://helm.cilium.io/
   KUBECONFIG=$CAPH_WORKER_CLUSTER_KUBECONFIG helm upgrade --install cilium cilium/cilium --version 1.14.4 \
       --namespace kube-system \
       -f cluster-api-provider-hetzner/templates/cilium/cilium.yaml
   ```

1. Install the CCM:

   ```
   helm repo add syself https://charts.syself.com
   helm repo update syself
   KUBECONFIG=$CAPH_WORKER_CLUSTER_KUBECONFIG helm upgrade --install ccm syself/ccm-hcloud --version 1.0.11 \
       --namespace kube-system \
       --set secret.name=hetzner \
       --set secret.tokenKeyName=hcloud \
       --set privateNetwork.enabled=false
   ```

1. Deploy the CSI:

   ```
   cat << EOF > csi-values.yaml
   storageClasses:
   - name: hcloud-volumes
   defaultStorageClass: true
   reclaimPolicy: Retain
   EOF

   KUBECONFIG=$CAPH_WORKER_CLUSTER_KUBECONFIG helm upgrade --install csi syself/csi-hcloud --version 0.2.0 \
   --namespace kube-system -f csi-values.yaml
   ```

1. Make sure everything looks good: `KUBECONFIG=/tmp/workload-kubeconfig kubectl get pods -A`

1. Merge the context into your kube config:

   ```
   KUBECONFIG=/tmp/workload-kubeconfig:~/.kube/config kubectl config view \
       --merge --flatten > out.txt
   mv out.txt ~/.kube/config
   kubectx # make sure you're pointing to the management cluster
   ```

1. To move the Cluster API objects from your bootstrap cluster to the new management cluster, you need first to install the Cluster API controllers. To install the components with the latest version, please run:
   ```
   clusterctl init --core cluster-api --bootstrap kubeadm --control-plane kubeadm --infrastructure hetzner
   kubectx kind-kind # switch back to bootstrap cluster
   clusterctl move --to-kubeconfig $CAPH_WORKER_CLUSTER_KUBECONFIG
   ```

And that's it! You now have a management cluster running in hetzner.
From here you could try deploying a new cluster using the management cluster.
You can also clean up the kind cluster with `kind delete cluster` which should also automatically remove it from kubectx.

## install external-dns on the mgmt cluster to make setting DNS records easy

We set up DNS records to route to apps running in users clusters this way.

```
export CF_API_TOKEN=<the key>
export CF_API_EMAIL=<email on the cf account>
kubectl create ns external-dns
kubectl create secret -n external-dns generic cloudflare-api-token --from-literal=apiToken=$CF_API_TOKEN --from-literal=email=$CF_API_EMAIL
echo 'provider:
  name: cloudflare
crd:
  create: true
sources:
  - crd
cloudflare:
  proxied: true
policy: sync
env:
  - name: CF_API_TOKEN
    valueFrom:
      secretKeyRef:
        name: cloudflare-api-token
        key: apiToken
        namespace: external-dns
  - name: CF_API_EMAIL
    valueFrom:
      secretKeyRef:
        name: cloudflare-api-token
        key: email
        namespace: external-dns
' > /tmp/values.yaml
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
helm upgrade --install external-dns external-dns/external-dns --values /tmp/values.yaml --namespace external-dns
```

Test it out:

```
echo 'apiVersion: externaldns.k8s.io/v1alpha1
kind: DNSEndpoint
metadata:
  name: examplednsrecord
spec:
  endpoints:
  - dnsName: "*.bar.onmetal.dev"
    #recordTTL: 180 # omit to use CFs "auto"
    recordType: A
    targets:
    - 192.168.99.216
    # cannot proxy local IPs but this is how you would do it if the IP was not a local IP
    #providerSpecific:
    #  - name: external-dns.alpha.kubernetes.io/cloudflare-proxied
    #    value: "true"
' > /tmp/test-a-record.yaml
kubectl apply -f /tmp/test-a-record.yaml
# go check it out in the cloudflare dashboard
kubectl delete -f /tmp/test-a-record.yaml
```

## install cert-manager on the mgmt cluster to create certs for users

We'd like for users to be able to run services over HTTPS on subdomains, e.g. <app name>-<env>.up.onmetal.dev.
To do this, set up cert-manager on the mgmt cluster so that we can create HTTPS certificates using LetsEncrypt.
When making a cert, Lets Encrypt requries you to prove that you are in control of the domain on the cert.
You can do this by answering a DNS challenge, i.e. putting a TXT record with a special value that Lets Encrypt provides.
We manage onmetal.dev DNS in Cloudflare, and luckily cert-manager has [support](https://cert-manager.io/docs/configuration/acme/dns01/cloudflare/) for generating Lets Encrypt certs and using Cloudflare's API to set up the TXT record to answer the DNS challenge.

```
export CLOUDFLARE_API_KEY=<the key>
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade --install cert-manager jetstack/cert-manager --version v1.14.5 --namespace cert-manager --set installCRDs=true --create-namespace
kubectl create -n cert-manager secret generic cloudflare-api-key-secret \
  --from-literal=api-key=$CLOUDFLARE_API_KEY
echo '---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    email: rgarcia2009@gmail.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-staging-private-key
    solvers:
    - dns01:
        cloudflare:
          email: rgarcia2009@gmail.com
          apiTokenSecretRef:
            name: cloudflare-api-key-secret
            key: api-key
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-production
spec:
  acme:
    email: certs@onmetal.dev
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-production-private-key
    solvers:
    - dns01:
        cloudflare:
          email: rgarcia2009@gmail.com
          apiTokenSecretRef:
            name: cloudflare-api-key-secret
            key: api-key
' | kubectl apply -n cert-manager -f -
```

Now you can make certificate requests by creating Certificate resources in the mgmt cluster, e.g.:

```
echo '---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  labels:
    name: test-subdomain-certificate
  name: test-subdomain-certificate
  namespace: cert-manager
spec:
  dnsNames:
  - 'test-service.up.onmetal.dev'
  issuerRef:
    kind: ClusterIssuer
    name: letsencrypt-production
  secretName: test-subdomain-certificate
' | kubectl apply -n cert-manager -f -
```

This will create a secret named `test-subdomain-certificate` containing the cert.
This can then be copied to a user's cluster via

```
kubectl get secret test-subdomain-certificate --namespace=cert-manager -oyaml | yq 'del(.metadata.annotations,.metadata.creationTimestamp,.metadata.labels,.metadata.namespace,.metadata.resourceVersion,.metadata.uid)' | \
  KUBECONFIG=<other cluster> kubectl apply -n <namespace secret will be used in in other cluster> -f -
```

## get the kubernetes web UI up and running

Summary of steps from the [docs](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/) + [other docs](https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/creating-sample-user.md).

1. Start it `kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml`

1. Proxy it locally: `kubectyl proxy`.

1. Create a service account

   ```
   cat << EOF > dashboard-serviceaccount.yaml
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: admin-user
     namespace: kubernetes-dashboard
   EOF

   kubectl apply -f dashboard-serviceaccount.yaml
   ```

1. Create cluster role binding

   ```
   cat << EOF > dashboard-clusterrolebinding.yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: admin-user
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: cluster-admin
   subjects:
   - kind: ServiceAccount
     name: admin-user
     namespace: kubernetes-dashboard
   EOF

   kubectl apply -f dashboard-clusterrolebinding.yaml
   ```

1. Create a token: `kubectl -n kubernetes-dashboard create token admin-user`

1. Navigate to http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#/login and use it to login.

## setting up flux cd

Useful overview: https://youtu.be/1DuxTlvmaNM?si=eXUb1hOReLbAFg9g (example repo: https://github.com/moonswitch-workshops/terraform-eks-flux/tree/main)
Commands summarized/adapted from https://fluxcd.io/flux/get-started/

```
export GITHUB_TOKEN=<personal access token w/ repo permissions>
export GITHUB_USER=rgarcia
flux check --pre
flux bootstrap github \
  --owner=$GITHUB_USER \
  --repository=gometal-infra \
  --branch=main \
  --path=./clusters/$CLUSTER_NAME \
  --personal
```

See github repo and comits ([1](https://github.com/rgarcia/gometal-infra/commits/01cf2c7) [2](https://github.com/rgarcia/gometal-infra/commits/8f5954a) [3](https://github.com/rgarcia/gometal-infra/commits/a47fb86)) for setup manifests for [Capacitor](https://github.com/gimlet-io/capacitor), a UI for flux.

## set up bitnami sealed secrets for being able to store secrets in git

Install the bitnami sealed-secrets helm chart via flux, see git [commit](https://github.com/rgarcia/gometal-infra/commit/1ea8384ecef7af0eaa5438ec414a86a856942352). This is the first set of steps from the flux [docs](https://fluxcd.io/flux/guides/sealed-secrets/).

Install kubeseal: `brew install kubeseal`

> At startup, the sealed-secrets controller generates a 4096-bit RSA key pair and persists the private and public keys as Kubernetes secrets in the flux-system namespace.

> You can retrieve the public key with:

```
kubeseal --fetch-cert \
--controller-name=sealed-secrets-controller \
--controller-namespace=flux-system \
> pub-sealed-secrets.pem
```

> The public key can be safely stored in Git, and can be used to encrypt secrets without direct access to the Kubernetes cluster.

The public key is in the github repo.

NOTE: the kubeseal command didn't work for me, but this did (via https://github.com/bitnami-labs/sealed-secrets/issues/368):

```
kubectl get secret \
  --namespace flux-system \
  --selector sealedsecrets.bitnami.com/sealed-secrets-key=active \
  --output jsonpath='{.items[0].data.tls\.crt}' \
| base64 -d > pub-sealed-secrets.pem
```

From this point, whenever you want to commit a secret it goes something like this:

1. Generate the Secret manifest

```
kubectl -n default create secret generic basic-auth \
--from-literal=user=admin \
--from-literal=password=change-me \
--dry-run=client \
-o yaml > basic-auth.yaml
```

2. Encrypt the secret with kubeseal:

```
kubeseal --format=yaml --cert=pub-sealed-secrets.pem \
< basic-auth.yaml > basic-auth-sealed.yaml
rm basic-auth.yaml
```

3. Apply basic-auth-sealed.yaml to the cluster.

## set up tailscale operator

Create an oauthclient and add tags to the tailscale ACL as described here: https://tailscale.com/kb/1236/kubernetes-operator.

Helm chart will be installed via flux and requires a sealed secret for the values.yaml file since it contains oauthclient credentials.

Using a sealed secret in a helm release is a little bit tricky. First set up the values.yaml file with the plaintext secret:

```
cat << EOF > values.yaml
oauth:
  clientId: "<in plaintext>"
  clientSecret: "<in plaintext>"
EOF
```

Other values for the tailscale operator can be found here: https://github.com/tailscale/tailscale/blob/main/cmd/k8s-operator/deploy/chart/values.yaml

Create the secret:

```
kubectl create secret generic tailscale-operator-values --from-file=values.yaml=./values.yaml --dry-run=client --namespace flux-system -o yaml > values-secret.yaml
rm values.yaml
```

Encrypt it

```
kubeseal --format=yaml --cert=pub-sealed-secrets.pem \
< values-secret.yaml > values-secret-sealed.yaml
rm values-secret.yaml
```

See git [commit](https://github.com/rgarcia/gometal-infra/commit/a81868d96b034ccca7c1b6e599ff38c547fc4521) that use this values-secret-sealed.yaml content to deploy the operator.

TODO: set up subnet router via the Connector CRD https://tailscale.com/kb/1236/kubernetes-operator

Other interesting stuff here: https://github.com/jaxxstorm

## set up signoz instead of prometheus / grafana / etc

single pane of glass seems nice:
https://signoz.io/docs/install/kubernetes/others/

`helm uninstall prometheus -n monitoring`

`kubectl patch storageclass hcloud-volumes -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}`

^ todo make this a fluxcd manifest

See github repo for setup manifests.

Troubleshooting:

```
kubectl -n monitoring run troubleshoot --image=signoz/troubleshoot \
  --restart='Never' -i --tty --rm --command -- ./troubleshoot checkEndpoint \
  --endpoint=signoz-release-otel-collector.monitoring.svc.cluster.local:4317
```

Adding cluster metric dashboards:

- https://github.com/SigNoz/dashboards/tree/main/k8s-infra-metrics
- TODO: is there a way to add these programatically?

Installed hotrod app (see github repo). Generated fake traffic with

```
kubectl --namespace hotrod run strzal --image=djbingham/curl \
  --restart='OnFailure' -i --tty --rm --command -- curl -X POST -F \
  'user_count=6' -F 'spawn_rate=2' http://locust-master:8089/swarm

```

https://github.com/SigNoz/signoz/tree/develop/sample-apps/hotrod

## TODO: setup backups

TODO: https://velero.io/

## TODO set up way to deploy stuff into AWS/GCP/Azure

- AWS-specific https://github.com/awslabs/aws-cloudformation-controller-for-flux
- not AWS-specific: https://github.com/pulumi/pulumi-kubernetes-operator

## turn on autoscaling?

first install metrics server: https://artifacthub.io/packages/helm/metrics-server/metrics-server

https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/

## run some operators?

- run OLM to manage them?
- clickhouse? very informative: https://youtu.be/sjeL_YC_n6A?si=zMdc5JuhmNYA33iR

## experiment with storage boxes for shared, s3-like storage between workloads: enable_csi_driver_smb

## Useful debugging commands

- are all pods coming up?

  ```
  KUBECONFIG=/tmp/workload-kubeconfig kubectl get pods -A
  ```

- cluster events

  ```
  kubectl get events -A --sort-by=.metadata.creationTimestamp
  ```

- Delete the cluster `kubectl delete cluster $CLUSTER_NAME`
- Delete the bootstrap cluster: `kind delete cluster`

- figure out what's not tracked by flux using this tool: https://github.com/raffis/gitops-zombies. Mostly CAPH, cluster-api, cert-manager stuff that is fine to not gitops for now.

# Archive

## set up monitoring with prometheus

From https://artifacthub.io/packages/helm/prometheus-community/prometheus

```
kubectl create namespace monitoring
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring
```

Find the name of the grafana pod and port forward it to log in:

```
kubectl port-forward prometheus-grafana-754c68c9d4-5v6lx 3000 -n monitoring
```

Navigate to http://localhost:3000. Default login is `admin` / `prom-operator`.
