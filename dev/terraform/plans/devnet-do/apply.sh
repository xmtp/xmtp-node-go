#!/bin/bash
set -eo pipefail
plan_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function tf() {
    terraform -chdir="${plan_dir}" "$@"
}

# Build and push docker images.
# TODO: fix snormorexmtp repo org references
export DOCKER_NODE_IMAGE="snormorexmtp/xmtpd:dev"
export DOCKER_E2E_IMAGE="snormorexmtp/xmtpd-e2e:dev"
(
    cd "${plan_dir}/../../../../"
    dev/docker/node/build
    dev/docker/e2e/build
)
DOCKER_NODE_IMAGE_FULL="$(docker inspect --format='{{index .RepoDigests 0}}' "${DOCKER_NODE_IMAGE}")"
DOCKER_E2E_IMAGE_FULL="$(docker inspect --format='{{index .RepoDigests 0}}' "${DOCKER_E2E_IMAGE}")"

export TF_VAR_node_container_image="${DOCKER_NODE_IMAGE_FULL}"
export TF_VAR_e2e_container_image="${DOCKER_E2E_IMAGE_FULL}"

# Initialize terraform.
tf init -upgrade

# Create clusters.
tf apply -auto-approve -target=digitalocean_kubernetes_cluster.cluster

# Apply the rest.
tf apply -auto-approve
