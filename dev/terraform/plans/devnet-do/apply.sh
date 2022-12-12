#!/bin/bash
set -eo pipefail
plan_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function tf() {
    terraform -chdir="${plan_dir}" "$@"
}

# Build and push docker images.
export DOCKER_NODE_IMAGE="snormorexmtp/xmtpd:dev"
export DOCKER_E2E_IMAGE="snormorexmtp/xmtpd-e2e:dev"
(
    cd "${plan_dir}/../../../../"
    export BUILDKIT_PROGRESS=plain
    pids=()
    dev/docker/node/build &
    pids+=("$!")
    dev/docker/e2e/build &
    pids+=("$!")
    for pid in ${pids[@]+"${pids[@]}"}; do
        wait "${pid}"
    done
)
DOCKER_NODE_IMAGE_FULL="$(docker inspect --format='{{index .RepoDigests 0}}' "${DOCKER_NODE_IMAGE}")"
DOCKER_E2E_IMAGE_FULL="$(docker inspect --format='{{index .RepoDigests 0}}' "${DOCKER_E2E_IMAGE}")"

export TF_VAR_node_container_image="${DOCKER_NODE_IMAGE_FULL}"
export TF_VAR_e2e_container_image="${DOCKER_E2E_IMAGE_FULL}"

# Initialize terraform.
tf init -upgrade

# Create clusters.
tf apply -auto-approve -target=module.cluster
CLUSTER_ID="$(tf output -json | jq -r '.cluster_id.value')"
kubeconfig_path="${plan_dir}/.xmtp/kubeconfig.yaml"
mkdir -p "$(dirname "${kubeconfig_path}")"
doctl k c kubeconfig show "${CLUSTER_ID}" > "${kubeconfig_path}"
echo

# Apply the rest.
tf apply -auto-approve

echo
echo "export KUBECONFIG=${kubeconfig_path}"
