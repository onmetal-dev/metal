#!/usr/bin/env bash
set -e

cat << END > /tmp/openapischema.yaml
openapi: 3.0.0
info:
  title: Argo Rollouts Status Schemas
  description: Pulling just the `status` fields from Argo CRDs
  version: 0.0.1
paths: {}
components:
  schemas:
    Rollout:
$(curl -s https://raw.githubusercontent.com/argoproj/argo-rollouts/master/manifests/crds/rollout-crd.yaml | yq '.spec.versions[0].schema.openAPIV3Schema' | sed -e 's/^/      /')
END
bunx openapi-typescript /tmp/openapischema.yaml -o ./gen/argoSchemas.ts

bunx cdk8s-cli import -o ./gen/ --language typescript https://raw.githubusercontent.com/argoproj/argo-rollouts/master/manifests/crds/rollout-crd.yaml
bunx cdk8s-cli import -o ./gen/ --language typescript https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/main/config/crd/experimental/gateway.networking.k8s.io_httproutes.yaml
bunx cdk8s-cli import -o ./gen/ --language typescript k8s