terraform {
  required_providers {
    helm = {
      source = "hashicorp/helm"
    }
  }
}

resource "helm_release" "node" {
  name      = var.name
  namespace = var.namespace
  chart     = "${path.module}/helm/xmtp-node"

  set {
    name  = "name"
    value = var.name
  }
  set {
    name  = "nodeId"
    value = var.node_id
  }
  set {
    name  = "p2pPublicAddress"
    value = var.p2p_public_address
  }
  dynamic "set" {
    for_each = var.persistent_peers
    content {
      name  = "persistentPeers[${set.key}]"
      value = set.value
    }
  }
  set {
    name  = "nodeImage"
    value = var.container_image
  }
  set_sensitive {
    name  = "nodeKey"
    value = var.private_key
  }

  values = var.helm_values
}
