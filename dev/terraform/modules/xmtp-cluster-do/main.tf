terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

provider "kubernetes" {
  host = data.digitalocean_kubernetes_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(
    data.digitalocean_kubernetes_cluster.cluster.kube_config[0].cluster_ca_certificate
  )
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "doctl"
    args        = ["kubernetes", "cluster", "kubeconfig", "exec-credential", "--version=v1beta1", data.digitalocean_kubernetes_cluster.cluster.id]
  }
}

provider "helm" {
  kubernetes {
    host = data.digitalocean_kubernetes_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(
      data.digitalocean_kubernetes_cluster.cluster.kube_config[0].cluster_ca_certificate
    )
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "doctl"
      args        = ["kubernetes", "cluster", "kubeconfig", "exec-credential", "--version=v1beta1", data.digitalocean_kubernetes_cluster.cluster.id]
    }
  }
}

// Cluster
//

data "digitalocean_kubernetes_versions" "cluster_versions" {
  version_prefix = "1."
}

data "digitalocean_kubernetes_cluster" "cluster" {
  name       = var.cluster_name
  depends_on = [digitalocean_kubernetes_cluster.cluster]
}

resource "digitalocean_kubernetes_cluster" "cluster" {
  name         = var.cluster_name
  region       = var.cluster_region
  auto_upgrade = true
  version      = data.digitalocean_kubernetes_versions.cluster_versions.latest_version

  node_pool {
    name       = "default"
    size       = "s-1vcpu-2gb"
    node_count = 2
  }
}

resource "digitalocean_kubernetes_node_pool" "cluster-xmtp-nodes" {
  cluster_id = digitalocean_kubernetes_cluster.cluster.id

  name       = "xmtp-nodes"
  size       = "s-2vcpu-4gb"
  node_count = length(var.nodes)
  tags       = ["xmtp-nodes"]

  labels = {
    service = "xmtp-nodes"
  }
}

module "cluster-monitoring" {
  source     = "../../modules/xmtp-monitoring"
  depends_on = [digitalocean_kubernetes_cluster.cluster]
}

// Nodes namespace
//

resource "kubernetes_namespace" "xmtp-nodes" {
  depends_on = [digitalocean_kubernetes_node_pool.cluster-xmtp-nodes]

  metadata {
    name = var.nodes_namespace
  }
}

// Nodes
//

module "nodes" {
  source     = "../../modules/xmtp-node"
  depends_on = [kubernetes_namespace.xmtp-nodes]
  count      = length(var.nodes)

  name               = var.nodes[count.index].name
  namespace          = var.nodes_namespace
  node_id            = var.nodes[count.index].node_id
  p2p_public_address = var.nodes[count.index].p2p_public_address
  persistent_peers   = var.nodes[count.index].persistent_peers
  private_key        = var.node_keys[var.nodes[count.index].name]
  container_image    = var.node_container_image

  helm_values = [
    "${file("xmtp-node-helm-values.yaml")}"
  ]
}

// API service
//

resource "kubernetes_service" "nodes-api" {
  depends_on = [kubernetes_namespace.xmtp-nodes]

  lifecycle {
    ignore_changes = [
      # LB ID is added dynamically after creation by the DO k8s controller manager.
      metadata[0].annotations["kubernetes.digitalocean.com/load-balancer-id"],
    ]
  }
  metadata {
    name      = "nodes-api"
    namespace = var.nodes_namespace
    annotations = {
      "service.beta.kubernetes.io/do-loadbalancer-protocol"                         = "http2"
      "service.beta.kubernetes.io/do-loadbalancer-certificate-id"                   = "57971176-cff9-4c9f-8745-c1d2b0f7f831"
      "service.beta.kubernetes.io/do-loadbalancer-disable-lets-encrypt-dns-records" = "false"
      "service.beta.kubernetes.io/do-loadbalancer-http-idle-timeout-seconds"        = "600"
    }
  }
  spec {
    selector = {
      "app.kubernetes.io/name" = "xmtp-node"
    }
    port {
      name        = "http"
      port        = 5555
      target_port = 5555
    }
    type = "LoadBalancer"
  }
}

// E2E
//

module "e2e" {
  source     = "../../modules/xmtp-e2e"
  depends_on = [kubernetes_namespace.xmtp-nodes]

  namespace       = var.nodes_namespace
  container_image = var.e2e_container_image
  api_url         = "http://nodes-api:5555"
}
