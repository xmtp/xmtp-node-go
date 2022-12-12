terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
  }
}

locals {
  num_nodes       = length(var.nodes)
  nodes_mid_index = floor(local.num_nodes / 2)
  nodes_group1    = slice(var.nodes, 0, local.nodes_mid_index)
  nodes_group2    = slice(var.nodes, local.nodes_mid_index, local.num_nodes)
}

module "ingress" {
  source = "../../modules/xmtp-ingress"

  namespace             = var.ingress_namespace
  node_pool_label_key   = var.node_pool_label_key
  node_pool_label_value = var.ingress_node_pool_label_value
  http_node_port        = var.http_node_port
  https_node_port       = var.https_node_port
  service_type          = var.ingress_service_type
  hostnames             = var.hostnames
  nodes_subdomains      = var.nodes[*].name
  utils_subdomains      = ["grafana", "chat"]
  cert_manager_email    = var.cert_manager_email
  ingress_class_name    = var.ingress_class_name

  nodes_namespace      = var.nodes_namespace
  node_api_http_port   = 5555
  cloudflare_api_token = var.cloudflare_api_token
  cloudflare_zone_id   = var.cloudflare_zone_id
  use_proxy_protocol   = var.use_proxy_protocol
}

module "monitoring" {
  source = "../../modules/xmtp-monitoring"
  count  = var.enable_monitoring ? 1 : 0

  namespace             = var.monitoring_namespace
  node_pool_label_key   = var.node_pool_label_key
  node_pool_label_value = var.monitoring_node_pool_label_value
  grafana_hostnames     = [for hostname in var.hostnames : "grafana.${hostname}"]
  ingress_class_name    = var.ingress_class_name
}

resource "kubernetes_namespace" "nodes" {
  metadata {
    name = var.nodes_namespace
  }
}

module "nodes_group1" {
  source     = "../../modules/xmtp-node"
  depends_on = [kubernetes_namespace.nodes]
  count      = length(local.nodes_group1)

  name                      = local.nodes_group1[count.index].name
  namespace                 = var.nodes_namespace
  node_id                   = local.nodes_group1[count.index].node_id
  p2p_public_address        = local.nodes_group1[count.index].p2p_public_address
  persistent_peers          = local.nodes_group1[count.index].persistent_peers
  private_key               = var.node_keys[local.nodes_group1[count.index].name]
  container_image           = var.node_container_image
  storage_class_name        = var.xmtp_node_storage_class_name
  storage_request           = var.xmtp_node_storage_request
  cpu_request               = var.xmtp_node_cpu_request
  hostnames                 = [for hostname in var.hostnames : "${local.nodes_group1[count.index].name}.${hostname}"]
  p2p_port                  = 9001
  api_http_port             = 5555
  metrics_port              = 8009
  node_pool_label_key       = var.node_pool_label_key
  node_pool_label_value     = var.xmtp_nodes_node_pool_label_value
  one_instance_per_k8s_node = var.one_instance_per_k8s_node
  ingress_class_name        = var.ingress_class_name
}

module "nodes_group2" {
  source     = "../../modules/xmtp-node"
  depends_on = [kubernetes_namespace.nodes, module.nodes_group1]
  count      = length(local.nodes_group2)

  name                      = local.nodes_group2[count.index].name
  namespace                 = var.nodes_namespace
  node_id                   = local.nodes_group2[count.index].node_id
  p2p_public_address        = local.nodes_group2[count.index].p2p_public_address
  persistent_peers          = local.nodes_group2[count.index].persistent_peers
  private_key               = var.node_keys[local.nodes_group2[count.index].name]
  container_image           = var.node_container_image
  storage_class_name        = var.xmtp_node_storage_class_name
  storage_request           = var.xmtp_node_storage_request
  cpu_request               = var.xmtp_node_cpu_request
  hostnames                 = [for hostname in var.hostnames : "${local.nodes_group2[count.index].name}.${hostname}"]
  p2p_port                  = 9001
  api_http_port             = 5555
  metrics_port              = 8009
  node_pool_label_key       = var.node_pool_label_key
  node_pool_label_value     = var.xmtp_nodes_node_pool_label_value
  one_instance_per_k8s_node = var.one_instance_per_k8s_node
  ingress_class_name        = var.ingress_class_name
}

module "e2e" {
  source     = "../../modules/xmtp-e2e"
  depends_on = [kubernetes_namespace.nodes]
  count      = var.enable_e2e ? 1 : 0

  namespace             = var.e2e_namespace
  container_image       = var.e2e_container_image
  api_url               = "http://nodes.${var.nodes_namespace}:5555"
  node_pool_label_key   = var.node_pool_label_key
  node_pool_label_value = var.e2e_node_pool_label_value
}

module "chat-app" {
  source = "../../modules/xmtp-chat-app"
  count  = var.enable_chat_app ? 1 : 0

  namespace             = var.chat_app_namespace
  node_pool_label_key   = var.node_pool_label_key
  node_pool_label_value = var.chat_app_node_pool_label_value
  api_url               = var.public_api_url
  hostnames             = [for hostname in var.hostnames : "chat.${hostname}"]
  ingress_class_name    = var.ingress_class_name
}
