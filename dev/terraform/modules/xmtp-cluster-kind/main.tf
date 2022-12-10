terraform {
  required_providers {
    kind = {
      source = "tehcyx/kind"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
  }
}

provider "kubernetes" {
  config_path = kind_cluster.cluster.kubeconfig_path
}

provider "helm" {
  kubernetes {
    config_path = kind_cluster.cluster.kubeconfig_path
  }
}

// Cluster
//

resource "kind_cluster" "cluster" {
  name            = var.cluster_name
  kubeconfig_path = var.kubeconfig_path
  wait_for_ready  = true
}

module "cluster-monitoring" {
  source     = "../../modules/xmtp-monitoring"
  depends_on = [kind_cluster.cluster]
}

// Nodes namespace
//

resource "kubernetes_namespace" "xmtp-nodes" {
  depends_on = [kind_cluster.cluster]

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
}

// API service
//

resource "kubernetes_service" "nodes-api" {
  depends_on = [kubernetes_namespace.xmtp-nodes]
  metadata {
    name      = "nodes-api"
    namespace = var.nodes_namespace
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
