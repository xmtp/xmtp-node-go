variable "node_container_image" {}
variable "e2e_container_image" {}
variable "nodes_namespace" {}
variable "monitoring_namespace" {}
variable "ingress_namespace" {}
variable "chat_app_namespace" {}
variable "e2e_namespace" {}
variable "nodes" {
  type = list(object({
    name               = string
    node_id            = string
    p2p_public_address = string
    persistent_peers   = list(string)
  }))
}
variable "node_keys" { type = map(string) }
variable "hostnames" { type = list(string) }
variable "cloudflare_api_token" {
  sensitive = true
  default   = ""
}
variable "cloudflare_zone_id" { default = "" }
variable "cert_manager_email" { default = "" }
variable "http_node_port" { default = 0 }
variable "https_node_port" { default = 0 }
variable "node_pool_label_key" {}
variable "ingress_node_pool_label_value" {}
variable "monitoring_node_pool_label_value" {}
variable "xmtp_nodes_node_pool_label_value" {}
variable "chat_app_node_pool_label_value" {}
variable "e2e_node_pool_label_value" {}
variable "ingress_service_type" {}
variable "xmtp_node_storage_class_name" {}
variable "xmtp_node_storage_request" {}
variable "xmtp_node_cpu_request" {}
variable "public_api_url" {}
variable "use_proxy_protocol" { type = bool }
variable "enable_e2e" { default = true }
variable "enable_chat_app" { default = true }
variable "enable_monitoring" { default = true }
variable "metrics_server_values" { default = [] }
variable "one_instance_per_k8s_node" { type = bool }
variable "ingress_class_name" { default = "traefik" }
