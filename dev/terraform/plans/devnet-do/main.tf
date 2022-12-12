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
    cloudflare = {
      source = "cloudflare/cloudflare"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

provider "kubernetes" {
  host                   = module.cluster.cluster_endpoint
  cluster_ca_certificate = module.cluster.cluster_ca_certificate
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "doctl"
    args        = ["kubernetes", "cluster", "kubeconfig", "exec-credential", "--version=v1beta1", module.cluster.cluster_id]
  }
}

provider "helm" {
  kubernetes {
    host                   = module.cluster.cluster_endpoint
    cluster_ca_certificate = module.cluster.cluster_ca_certificate
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "doctl"
      args        = ["kubernetes", "cluster", "kubeconfig", "exec-credential", "--version=v1beta1", module.cluster.cluster_id]
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

module "cluster" {
  source = "../../modules/xmtp-cluster-do"

  cluster_name   = var.cluster_name
  do_token       = var.do_token
  cluster_region = var.do_cluster_region

  node_pool_label_key      = "node-pool"
  default_pool_label_value = "default"
  default_pool_node_size   = "s-1vcpu-2gb"

  num_xmtp_nodes_pool_nodes   = length(var.nodes)
  xmtp_nodes_pool_label_value = "xmtp-nodes"
  xmtp_nodes_pool_node_size   = "s-2vcpu-2gb"

  num_xmtp_utils_pool_nodes   = 2
  xmtp_utils_pool_label_value = "xmtp-utils"
  xmtp_utils_pool_node_size   = "s-1vcpu-2gb"
}

module "devnet" {
  source     = "../../modules/xmtp-devnet"
  depends_on = [module.cluster]

  node_container_image = var.node_container_image
  e2e_container_image  = var.e2e_container_image
  nodes                = var.nodes
  node_keys            = var.node_keys
  hostnames            = var.hostnames
  public_api_url       = "https://${var.hostnames[0]}"
  cert_manager_email   = var.cert_manager_email
  cloudflare_api_token = var.cloudflare_api_token
  cloudflare_zone_id   = var.cloudflare_zone_id
  enable_e2e          = var.enable_e2e
  enable_chat_app     = var.enable_chat_app
  enable_monitoring   = var.enable_monitoring

  nodes_namespace                  = "xmtp-nodes"
  monitoring_namespace             = "monitoring"
  ingress_namespace                = "ingress"
  chat_app_namespace               = "chat-app"
  e2e_namespace                    = "e2e"
  node_pool_label_key              = "node-pool"
  ingress_node_pool_label_value    = "xmtp-utils"
  monitoring_node_pool_label_value = "xmtp-utils"
  chat_app_node_pool_label_value   = "xmtp-utils"
  e2e_node_pool_label_value        = "xmtp-utils"
  xmtp_nodes_node_pool_label_value = "xmtp-nodes"
  ingress_service_type             = "LoadBalancer"
  xmtp_node_storage_class_name     = "do-block-storage"
  xmtp_node_storage_request        = "10Gi"
  xmtp_node_cpu_request            = "10m"
  use_proxy_protocol               = true
  one_instance_per_k8s_node        = true
}
