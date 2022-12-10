# xmtp-node-go

XMTP node software

## Instructions

### Install dependencies and start the DB

1. `dev/up`

### Run the tests

1. `dev/test`

### Start a local node

1. `dev/start`

### Debugging metrics

1. `dev/start --metrics`
2. Browse to http://localhost:9090 to see prometheus interface

## Deployments

Merging a PR to the `main` branch will trigger a new deployment via Github Actions and Terraform.

The default behavior is to deploy `main` to both the `dev` and `production` environments. If you'd like to deploy a different branch to `dev`, open a PR with an update to [.github/workflows/deploy.yml](https://github.com/xmtp/xmtp-node-go/blob/main/.github/workflows/deploy.yml#L29) switching from `main` to your branch. Remember to PR it back to `main` when you're done.

Changes will be automatically applied and no user action is required after merging.

### Kubernetes

Bootstrap the local `dev/terraform/plans/devnet-local` [kind](https://github.com/kubernetes-sigs/kind) cluster with:

```sh
dev/k8s/up
```

By default `dev/k8s/up` and most other commands will default to using the `devnet-local` plan, but you can override this by setting the `PLAN` environment variable.

Generate terraform variables and secrets files containing a desired number of nodes with:

```sh
dev/terraform/generate dev/terraform/plans/devnet-local 5
```

Overwrite existing terraform variables and secrets files with updated number of nodes with:

```sh
dev/terraform/generate dev/terraform/plans/devnet-local 2
```

A few things you can do from here:

```sh
# View grafana dashboards
alias "k=kubectl"
k get secret grafana -o jsonpath={.data.admin-password} --namespace monitoring | base64 --decode ; echo
k port-forward -n monitoring svc/grafana 8888:80
open http://localhost:8888

# View logs
alias "kn=kubectl -n xmtp-nodes"
kn get pods -w
kn logs -f -l app.kubernetes.io/name=xmtp-node
kn logs -f -l app=e2e

# Query the nodes API
kn port-forward svc/nodes-api 5555
curl -sS -XPOST localhost:5555/message/v1/publish -d '{"envelopes":[{"content_topic":"topic", "timestamp_ns": 1}]}'
curl -sS -XPOST localhost:5555/message/v1/query -d '{"content_topics":["topic"]}' | jq

# Add and sync a new local node via persistent-peer
kn port-forward svc/node-4tu 19001:9001
dev/start --p2p-port=0 --p2p-persistent-peer=/ip4/127.0.0.1/tcp/19001/p2p/12D3KooWMRKLtqptNqxsHpVobt2XZTa2f68XbS2vb553JVkGhrie
curl -sS -XPOST localhost:5555/message/v1/query | jq '.envelopes | length'
```

Tear it down with:

```sh
dev/k8s/down
```
