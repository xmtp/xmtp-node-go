variable "namespace" {}
variable "node_pool_label_key" {}
variable "node_pool_label_value" {}
variable "http_node_port" { default = 0 }
variable "https_node_port" { default = 0 }
variable "cert_manager_email" { default = "" }
variable "service_type" {}
variable "use_proxy_protocol" {}
variable "hostnames" { type = list(string) }
variable "nodes_subdomains" { type = list(string) }
variable "utils_subdomains" { type = list(string) }
variable "node_api_http_port" { type = number }
variable "cloudflare_api_token" {
  sensitive = true
  default   = ""
}
variable "cloudflare_zone_id" { default = "" }
variable "nodes_namespace" {}
variable "ingress_class_name" {}
