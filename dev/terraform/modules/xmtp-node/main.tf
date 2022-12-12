terraform {
  required_providers {
    helm = {
      source = "hashicorp/helm"
    }
  }
}

locals {
  labels = {
    "app.kubernetes.io/name" = "xmtp-node"
    "node.xmtp.dev/name" : var.name
  }
}

resource "kubernetes_ingress_v1" "ingress" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels    = local.labels
  }
  spec {
    ingress_class_name = var.ingress_class_name
    tls {
      hosts = var.hostnames
    }
    dynamic "rule" {
      for_each = var.hostnames
      content {
        host = rule.value
        http {
          path {
            path = "/"
            backend {
              service {
                name = var.name
                port {
                  number = var.api_http_port
                }
              }
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "service" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels    = local.labels
    annotations = {
      "prometheus.io/scrape" = "true"
      "prometheus.io/path"   = "/metrics"
      "prometheus.io/port"   = var.metrics_port
      "prometheus.io/scheme" = "http"
    }
  }
  spec {
    selector = local.labels
    port {
      name = "api"
      port = var.api_http_port
    }
    port {
      name = "p2p"
      port = var.p2p_port
    }
    port {
      name = "metrics"
      port = var.metrics_port
    }
  }
}

resource "kubernetes_secret" "secret" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels    = local.labels
  }
  data = {
    XMTP_NODE_KEY = var.private_key
  }
}

resource "kubernetes_stateful_set" "statefulset" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels    = local.labels
  }
  spec {
    selector {
      match_labels = local.labels
    }
    service_name = var.name
    replicas     = 1
    template {
      metadata {
        labels = local.labels
      }
      spec {
        termination_grace_period_seconds = 10
        node_selector = {
          (var.node_pool_label_key) = var.node_pool_label_value
        }
        dynamic "affinity" {
          for_each = var.one_instance_per_k8s_node ? [1] : []
          content {
            pod_anti_affinity {
              required_during_scheduling_ignored_during_execution {
                label_selector {
                  match_expressions {
                    key      = "app.kubernetes.io/name"
                    operator = "In"
                    values   = [local.labels["app.kubernetes.io/name"]]
                  }
                }
                topology_key = "kubernetes.io/hostname"
              }
            }
          }
        }
        container {
          name  = "node"
          image = var.container_image
          port {
            name           = "http-api"
            container_port = var.api_http_port
          }
          port {
            name           = "p2p"
            container_port = var.p2p_port
          }
          volume_mount {
            name       = "data"
            mount_path = "/data"
          }
          env_from {
            secret_ref {
              name = var.name
            }
          }
          command = concat(
            [
              "xmtpd",
              "--p2p-port=${var.p2p_port}",
              "--metrics",
              "--metrics-address=0.0.0.0",
              "--metrics-port=${var.metrics_port}",
              "--api.http-port=${var.api_http_port}",
              "--log-encoding=json",
              "--data-path=/data",
            ],
            [for peer in var.persistent_peers : "--p2p-persistent-peer=${peer}"],
          )
          readiness_probe {
            http_get {
              path = "/healthz"
              port = "http-api"
            }
            success_threshold = 1
            failure_threshold = 3
            period_seconds    = 10
          }
          resources {
            requests = {
              "cpu" = var.cpu_request
            }
          }
        }
      }
    }
    volume_claim_template {
      metadata {
        name   = "data"
        labels = local.labels
      }
      spec {
        access_modes = [
          "ReadWriteOnce"
        ]
        storage_class_name = var.storage_class_name
        resources {
          requests = {
            "storage" = var.storage_request
          }
        }
      }
    }
  }
}
