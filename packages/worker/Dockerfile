FROM debian:bookworm-slim
SHELL ["/bin/bash", "-c"]
RUN apt update && apt upgrade -y
RUN apt install wget curl unzip -y

# install bun
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH "$PATH:/root/.bun/bin"

# install ssh-keygen, htpasswd, jq, yq
RUN apt install openssh-client apache2-utils jq -y
RUN wget https://github.com/mikefarah/yq/releases/download/v4.44.1/yq_linux_amd64 -O /usr/bin/yq && chmod +x /usr/bin/yq

# install kubectl, helm, kustomize
RUN curl -LO "https://dl.k8s.io/release/v1.30.1/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && chmod 700 get_helm.sh && ./get_helm.sh
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash && mv kustomize /usr/local/bin/

# install clusterctl, argo rollouts plugin, minio client
RUN curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.7.2/clusterctl-linux-amd64 -o clusterctl && install -o root -g root -m 0755 clusterctl /usr/local/bin/clusterctl
RUN curl -LO "https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64" && chmod +x ./kubectl-argo-rollouts-linux-amd64 && mv ./kubectl-argo-rollouts-linux-amd64 /usr/local/bin/kubectl-argo-rollouts
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x mc && mv mc /usr/local/bin

# install the app
ADD . /app
WORKDIR /app
RUN bun install --no-save --frozen-lockfile --production
CMD bun run --filter @metal/worker start
