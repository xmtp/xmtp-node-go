variable "name" {
  type = string
}

variable "namespace" {
  type = string
}

variable "node_id" {
  type = string
}

variable "p2p_public_address" {
  type = string
}

variable "private_key" {
  type      = string
  sensitive = true
}

variable "persistent_peers" {
  type = list(string)
}

variable "container_image" {
  type = string
}

variable "helm_values" {
  type    = list(string)
  default = []
}
