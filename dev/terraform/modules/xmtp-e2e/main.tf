terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
}

resource "kubernetes_deployment" "runner" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels = {
      app = var.name
    }
  }
  spec {
    replicas = 1
    selector {
      match_labels = {
        app = var.name
      }
    }
    template {
      metadata {
        labels = {
          app = var.name
        }
        annotations = {
          "prometheus.io/scrape" = "true"
          "prometheus.io/path"   = "/metrics"
          "prometheus.io/port"   = "9009"
          "prometheus.io/scheme" = "http"
        }
      }
      spec {
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
