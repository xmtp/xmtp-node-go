locals {
  labels = {
    "app.kubernetes.io/name" = var.name
  }
}

resource "kubernetes_namespace" "namespace" {
  metadata {
    name   = var.namespace
    labels = local.labels
  }
}

resource "kubernetes_deployment" "runner" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels    = local.labels
  }
  spec {
    replicas = 1
    selector {
      match_labels = local.labels
    }
    template {
      metadata {
        labels = local.labels
        annotations = {
          "prometheus.io/scrape" = "true"
          "prometheus.io/path"   = "/metrics"
          "prometheus.io/port"   = "9009"
          "prometheus.io/scheme" = "http"
        }
      }
      spec {
        node_selector = {
          (var.node_pool_label_key) = var.node_pool_label_value
        }
        container {
          name  = "runner"
          image = var.container_image
          port {
            name           = "metrics"
            container_port = 9009
          }
          port {
            name           = "health"
            container_port = 6062
          }
          env {
            name  = "XMTP_API_URL"
            value = var.api_url
          }
          env {
            name  = "E2E_CONTINUOUS"
            value = "yes"
          }
          readiness_probe {
            http_get {
              path = "/healthz"
              port = "health"
            }
            success_threshold = 1
            failure_threshold = 3
            period_seconds    = 3
          }
        }
      }
    }
  }
}
