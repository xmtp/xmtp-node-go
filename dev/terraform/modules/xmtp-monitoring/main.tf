terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
  }
}

// Namespace
//

resource "kubernetes_namespace" "namespace" {
  metadata {
    name = var.namespace
  }
}

// Prometheus
// 

resource "helm_release" "prometheus" {
  depends_on = [kubernetes_namespace.namespace]
  name       = "prometheus"
  namespace  = var.namespace
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus"
}

// Grafana
// 

resource "kubernetes_config_map" "xmtp-dashboards" {
  depends_on = [kubernetes_namespace.namespace]
  metadata {
    name      = "xmtp-dashboards"
    namespace = var.namespace
  }
  data = {
    "xmtp-network-api.json" = "${file("${path.module}/grafana-dashboards/xmtp-network-api.json")}"
  }
}

resource "helm_release" "grafana" {
  name       = "grafana"
  namespace  = var.namespace
  depends_on = [kubernetes_config_map.xmtp-dashboards]
  repository = "https://grafana.github.io/helm-charts"
  chart      = "grafana"
  values = [
    "${file("${path.module}/grafana-helm-values.yaml")}"
  ]
}
