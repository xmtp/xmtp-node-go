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

resource "helm_release" "metrics-server" {
  name       = "metrics-server"
  namespace  = "kube-system"
  repository = "https://kubernetes-sigs.github.io/metrics-server/"
  chart      = "metrics-server"

  set {
    name  = "nodeSelector.${var.node_pool_label_key}"
    value = var.node_pool_label_value
  }

  values = [
    <<-EOF
    args:
      - --kubelet-insecure-tls
      - --kubelet-preferred-address-types=InternalIP
    EOF
  ]
}

resource "kubernetes_namespace" "namespace" {
  metadata {
    name = var.namespace
  }
}

resource "helm_release" "prometheus" {
  depends_on = [kubernetes_namespace.namespace]
  name       = "prometheus"
  namespace  = var.namespace
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus"
  set {
    name  = "server.nodeSelector.${var.node_pool_label_key}"
    value = var.node_pool_label_value
  }
  set {
    name  = "alertmanager.nodeSelector.${var.node_pool_label_key}"
    value = var.node_pool_label_value
  }
  set {
    name  = "kube-state-metrics.nodeSelector.${var.node_pool_label_key}"
    value = var.node_pool_label_value
  }
  set {
    name  = "prometheus-pushgateway.nodeSelector.${var.node_pool_label_key}"
    value = var.node_pool_label_value
  }
}

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
  set {
    name  = "nodeSelector.${var.node_pool_label_key}"
    value = var.node_pool_label_value
  }
}

resource "kubernetes_ingress_v1" "grafana" {
  count = length(var.grafana_hostnames) > 0 ? 1 : 0
  metadata {
    name      = "grafana"
    namespace = var.namespace
    annotations = {
      "cert-manager.io/cluster-issuer" = "cert-manager"
    }
  }
  spec {
    ingress_class_name = var.ingress_class_name
    tls {
      hosts       = var.grafana_hostnames
      secret_name = "grafana-tls"
    }
    dynamic "rule" {
      for_each = var.grafana_hostnames
      content {
        host = rule.value
        http {
          path {
            path = "/"
            backend {
              service {
                name = "grafana"
                port {
                  number = 80
                }
              }
            }
          }
        }
      }
    }
  }
}
