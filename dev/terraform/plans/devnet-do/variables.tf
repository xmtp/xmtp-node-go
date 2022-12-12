variable "do_token" {}
variable "do_cluster_region" { default = "nyc1" }
variable "node_container_image" { default = "xmtp/xmtpd:dev" }
variable "e2e_container_image" { default = "xmtp/xmtpd-e2e:dev" }
variable "cluster_name" { default = "xmtp-devnet-nyc1" }
variable "nodes_namespace" { default = "xmtp-nodes" }
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
variable "cloudflare_api_token" { sensitive = true }
variable "cloudflare_zone_id" {}
variable "cert_manager_email" {}
variable "enable_e2e" { default = true }
variable "enable_chat_app" { default = true }
variable "enable_monitoring" { default = true }
