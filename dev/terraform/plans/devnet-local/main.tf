module "cluster" {
  source = "../../modules/xmtp-cluster-kind"

  cluster_name         = var.cluster_name
  nodes_namespace      = var.nodes_namespace
  node_container_image = var.node_container_image
  e2e_container_image  = var.e2e_container_image
  nodes                = var.nodes
  node_keys            = var.node_keys

  kubeconfig_path = startswith(var.kubeconfig_path, "/") ? var.kubeconfig_path : abspath(var.kubeconfig_path)
}
