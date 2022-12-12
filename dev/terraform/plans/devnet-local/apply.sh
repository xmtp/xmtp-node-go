#!/bin/bash
set -eo pipefail
plan_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function tf() {
    terraform -chdir="${plan_dir}" "$@"
}

# Build docker images.
export DOCKER_NODE_IMAGE="xmtpdev/xmtpd:dev"
export DOCKER_E2E_IMAGE="xmtpdev/xmtpd-e2e:dev"
(
    cd "${plan_dir}/../../../../"
    export BUILDKIT_PROGRESS=plain
    pids=()
    dev/docker/node/build-local &
    pids+=("$!")
    dev/docker/e2e/build-local &
    pids+=("$!")
    for pid in ${pids[@]+"${pids[@]}"}; do
        wait "${pid}"
    done
)
DOCKER_NODE_IMAGE_FULL="$(docker inspect --format='{{.Id}}' "${DOCKER_NODE_IMAGE}")"
DOCKER_E2E_IMAGE_FULL="$(docker inspect --format='{{.Id}}' "${DOCKER_E2E_IMAGE}")"

export TF_VAR_node_container_image="${DOCKER_NODE_IMAGE_FULL}"
export TF_VAR_e2e_container_image="${DOCKER_E2E_IMAGE_FULL}"
export TF_VAR_kubeconfig_path="${plan_dir}/.xmtp/kubeconfig.yaml"

# Initialize terraform.
tf init -upgrade

# Create clusters.
tf apply -auto-approve -target=module.cluster
KIND_CLUSTER="$(tf output -json | jq -r '.cluster_name.value')"
echo

# Load local docker images into kind cluster.
export KUBECONFIG="${TF_VAR_kubeconfig_path}"
nodes="$(kubectl get nodes -l "node-pool=xmtp-nodes" --no-headers -o custom-columns=":metadata.name" | paste -s -d, -)"
kind load docker-image "${DOCKER_NODE_IMAGE}" --name "${KIND_CLUSTER}" --nodes "${nodes}"

nodes="$(kubectl get nodes -l "node-pool=xmtp-utils" --no-headers -o custom-columns=":metadata.name" | paste -s -d, -)"
kind load docker-image "${DOCKER_E2E_IMAGE}" --name "${KIND_CLUSTER}" --nodes "${nodes}"

# Apply the rest.
tf apply -auto-approve
