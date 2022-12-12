terraform {
  required_providers {
    kind = {
      source = "tehcyx/kind"
    }
  }
}

resource "kind_cluster" "cluster" {
  name            = var.cluster_name
  kubeconfig_path = var.kubeconfig_path
  wait_for_ready  = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"
    node {
      role = "control-plane"
      labels = {
        (var.node_pool_label_key) = "control-plane"
      }
    }
    node {
      role = "worker"
      labels = {
        (var.node_pool_label_key) = var.xmtp_utils_pool_label_value
        "ingress-ready"           = "true"
      }
      extra_port_mappings {
        container_port = var.http_node_port
        host_port      = 80
      }
      extra_port_mappings {
        container_port = var.https_node_port
        host_port      = 443
      }
    }
    dynamic "node" {
      for_each = range(var.num_xmtp_nodes)
      content {
        role = "worker"
        labels = {
          (var.node_pool_label_key) = var.xmtp_nodes_pool_label_value
        }
      }
    }
  }
}
