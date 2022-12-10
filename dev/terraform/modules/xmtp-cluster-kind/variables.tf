variable "node_container_image" {
  type = string
}

variable "e2e_container_image" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "kubeconfig_path" {
  type = string
}

variable "nodes_namespace" {
  type = string
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
