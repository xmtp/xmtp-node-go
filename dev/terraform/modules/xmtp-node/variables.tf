variable "name" {}
variable "namespace" {}
variable "node_id" {}
variable "p2p_public_address" {}
variable "private_key" { sensitive = true }
variable "persistent_peers" { type = list(string) }
variable "container_image" {}
variable "helm_values" { default = [] }
variable "hostnames" { type = list(string) }
variable "storage_class_name" {}
variable "storage_request" {}
variable "cpu_request" {}
variable "p2p_port" { type = number }
variable "api_http_port" { type = number }
variable "metrics_port" { type = number }
variable "node_pool_label_key" {}
variable "node_pool_label_value" {}
variable "one_instance_per_k8s_node" { type = bool }
variable "ingress_class_name" {}
