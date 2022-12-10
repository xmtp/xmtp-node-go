variable "kubeconfig_path" {
  type    = string
  default = ".xmtp/kubeconfig.yaml"
}

variable "node_container_image" {
  type    = string
  default = "xmtp/xmtpd:dev"
}

variable "e2e_container_image" {
  type    = string
  default = "xmtp/xmtpd-e2e:dev"
}

variable "cluster_name" {
  type    = string
  default = "xmtp-devnet-local"
}

variable "nodes_namespace" {
  type    = string
  default = "xmtp-nodes"
}

variable "nodes" {
  type = list(object({
    name               = string
    node_id            = string
    p2p_public_address = string
    persistent_peers   = list(string)
  }))
}

variable "node_keys" {
  type = map(string)
}
