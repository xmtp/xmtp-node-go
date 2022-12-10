#!/bin/bash
set -eo pipefail
plan_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function tf() {
    terraform -chdir="${plan_dir}" "$@"
}

# Build docker images.
# TODO: fix snormorexmtp repo org reference
export DOCKER_NODE_IMAGE="snormorexmtp/xmtpd:dev"
export DOCKER_E2E_IMAGE="snormorexmtp/xmtpd-e2e:dev"
(
    cd "${plan_dir}/../../../../"
    dev/docker/node/build-local
    dev/docker/e2e/build-local
)

export TF_VAR_node_container_image="${DOCKER_NODE_IMAGE}"
export TF_VAR_e2e_container_image="${DOCKER_E2E_IMAGE}"
export TF_VAR_kubeconfig_path="${plan_dir}/.xmtp/kubeconfig.yaml"

# Initialize terraform.
tf init -upgrade

# Create clusters.
tf apply -auto-approve -target=kind_cluster.cluster
KIND_CLUSTER="$(tf output -json | jq -r '.cluster_name.value')"

# Load local docker images into kind cluster.
kind load docker-image "${DOCKER_NODE_IMAGE}" --name "${KIND_CLUSTER}" &
kind load docker-image "${DOCKER_E2E_IMAGE}" --name "${KIND_CLUSTER}" &
wait

# Apply the rest.
tf apply -auto-approve
