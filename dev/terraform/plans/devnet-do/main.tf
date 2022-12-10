module "cluster" {
  source = "../../modules/xmtp-cluster-do"

  cluster_name         = var.cluster_name
  nodes_namespace      = var.nodes_namespace
  node_container_image = var.node_container_image
  e2e_container_image  = var.e2e_container_image
  nodes                = var.nodes
  node_keys            = var.node_keys

  do_token       = var.do_token
  cluster_region = var.do_cluster_region
}
