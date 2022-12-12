variable "cluster_name" {}
variable "kubeconfig_path" {}
variable "http_node_port" { type = number }
variable "https_node_port" { type = number }
variable "node_pool_label_key" {}
variable "num_xmtp_nodes" { type = number }
variable "xmtp_nodes_pool_label_value" {}
variable "xmtp_utils_pool_label_value" {}
