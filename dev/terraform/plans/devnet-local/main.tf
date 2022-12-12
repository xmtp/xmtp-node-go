terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
    cloudflare = {
      source = "cloudflare/cloudflare"
    }
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig_path
  }
}

provider "cloudflare" {
  api_token = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRZTUVWXYZ"
}

module "cluster" {
  source = "../../modules/xmtp-cluster-kind"

  kubeconfig_path = startswith(var.kubeconfig_path, "/") ? var.kubeconfig_path : abspath(var.kubeconfig_path)
  num_xmtp_nodes  = var.num_xmtp_node_pool_nodes

  cluster_name                = "xmtp-devnet-local"
  http_node_port              = 32080
  https_node_port             = 32443
  node_pool_label_key         = "node-pool"
  xmtp_nodes_pool_label_value = "xmtp-nodes"
  xmtp_utils_pool_label_value = "xmtp-utils"
}

module "devnet" {
  source     = "../../modules/xmtp-devnet"
  depends_on = [module.cluster]

  node_container_image = var.node_container_image
  e2e_container_image  = var.e2e_container_image
  nodes                = var.nodes
  node_keys            = var.node_keys
  enable_e2e          = var.enable_e2e
  enable_chat_app     = var.enable_chat_app
  enable_monitoring   = var.enable_monitoring

  nodes_namespace                  = "xmtp-nodes"
  monitoring_namespace             = "monitoring"
  ingress_namespace                = "ingress"
  chat_app_namespace               = "chat-app"
  e2e_namespace                    = "e2e"
  hostnames                        = ["localhost", "xmtp.local"]
  public_api_url                   = "http://localhost"
  http_node_port                   = 32080
  https_node_port                  = 32443
  node_pool_label_key              = "node-pool"
  ingress_node_pool_label_value    = "xmtp-utils"
  monitoring_node_pool_label_value = "xmtp-utils"
  chat_app_node_pool_label_value   = "xmtp-utils"
  e2e_node_pool_label_value        = "xmtp-utils"
  xmtp_nodes_node_pool_label_value = "xmtp-nodes"
  ingress_service_type             = "NodePort"
  xmtp_node_storage_class_name     = "standard"
  xmtp_node_storage_request        = "1Gi"
  xmtp_node_cpu_request            = "10m"
  use_proxy_protocol               = false
  one_instance_per_k8s_node        = false
}
