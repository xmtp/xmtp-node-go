terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    helm = {
      source = "hashicorp/helm"
    }
    cloudflare = {
      source = "cloudflare/cloudflare"
    }
  }
}

locals {
  nodes_domains = [for val in setproduct(var.nodes_subdomains, var.hostnames) : "${val.0}.${val.1}"]
  utils_domains = [for val in setproduct(var.utils_subdomains, var.hostnames) : "${val.0}.${val.1}"]
}

resource "kubernetes_namespace" "namespace" {
  metadata {
    name = var.namespace
  }
}

resource "helm_release" "traefik" {
  name       = "traefik"
  namespace  = var.namespace
  depends_on = [kubernetes_namespace.namespace]
  repository = "https://traefik.github.io/charts"
  chart      = "traefik"

  dynamic "set" {
    for_each = var.http_node_port != 0 ? [1] : []
    content {
      name  = "ports.web.nodePort"
      value = var.http_node_port
    }
  }
  dynamic "set" {
    for_each = var.https_node_port != 0 ? [1] : []
    content {
      name  = "ports.websecure.nodePort"
      value = var.https_node_port
    }
  }
  dynamic "set" {
    for_each = var.ingress_class_name != "" ? ["kubernetesCRD", "kubernetesIngress"] : []
    content {
      name  = "providers.${set.value}.ingressClass"
      value = var.ingress_class_name
    }
  }

  values = [
    <<-EOF
    nodeSelector:
      ${var.node_pool_label_key}: ${var.node_pool_label_value}
    service:
      type: ${var.service_type}
    EOF
    ,
    <<-EOF
    service:
      annotations:
        service.beta.kubernetes.io/do-loadbalancer-enable-proxy-protocol: "${var.use_proxy_protocol}"
    ports:
      web:
        proxyProtocol:
          trustedIPs: ["127.0.0.1/32","10.120.0.0/16"]
      websecure:
        proxyProtocol:
          trustedIPs: ["127.0.0.1/32","10.120.0.0/16"]
    EOF
  ]
}

resource "kubernetes_secret" "cloudflare" {
  metadata {
    name      = "cloudflare"
    namespace = var.namespace
  }
  data = {
    api-token = var.cloudflare_api_token
  }
}

module "cert-manager" {
  source     = "terraform-iaac/cert-manager/kubernetes"
  count      = var.cert_manager_email != "" ? 1 : 0
  depends_on = [kubernetes_secret.cloudflare]

  namespace_name       = var.namespace
  create_namespace     = false
  cluster_issuer_email = var.cert_manager_email
  solvers = [
    {
      dns01 = {
        cloudflare = {
          email = var.cert_manager_email
          apiTokenSecretRef = {
            name = "cloudflare"
            key  = "api-token"
          }
        },
        selector = {
          dnsZones = var.hostnames
        }
      },
    }
  ]

  additional_set = [
    {
      name  = "nodeSelector.${var.node_pool_label_key}"
      value = var.node_pool_label_value
    },
    {
      name  = "webhook.nodeSelector.${var.node_pool_label_key}"
      value = var.node_pool_label_value
    },
    {
      name  = "cainjector.nodeSelector.${var.node_pool_label_key}"
      value = var.node_pool_label_value
    },
    {
      name  = "startupapicheck.nodeSelector.${var.node_pool_label_key}"
      value = var.node_pool_label_value
    },
  ]
}

data "kubernetes_service" "traefik" {
  depends_on = [helm_release.traefik]
  metadata {
    name      = "traefik"
    namespace = var.namespace
  }
}

locals {
  // If you are using a nested subdomain like xmtp.example.com, then the the CF
  // default/free universal cert won't cover the *.xmtp.example.com domains, so
  // we can't used proxied in this case, but normally would. If you are in this
  // scenario, you can upgrade in CF and get "Total TLS" under
  // SSL / TLS -> Edige Certificates.
  cloudflare_proxied = false
}

resource "cloudflare_record" "naked_domains" {
  depends_on = [helm_release.traefik]
  count      = var.cloudflare_api_token != "" ? length(var.hostnames) : 0
  zone_id    = var.cloudflare_zone_id
  name       = var.hostnames[count.index]
  value      = data.kubernetes_service.traefik.status.0.load_balancer.0.ingress.0.ip
  type       = "A"
  proxied    = local.cloudflare_proxied
}

resource "cloudflare_record" "nodes_domains" {
  depends_on = [helm_release.traefik]
  count      = var.cloudflare_api_token != "" ? length(local.nodes_domains) : 0
  zone_id    = var.cloudflare_zone_id
  name       = local.nodes_domains[count.index]
  value      = data.kubernetes_service.traefik.status.0.load_balancer.0.ingress.0.ip
  type       = "A"
  proxied    = local.cloudflare_proxied
}

resource "cloudflare_record" "utils_domains" {
  depends_on = [helm_release.traefik]
  count      = var.cloudflare_api_token != "" ? length(local.utils_domains) : 0
  zone_id    = var.cloudflare_zone_id
  name       = local.utils_domains[count.index]
  value      = data.kubernetes_service.traefik.status.0.load_balancer.0.ingress.0.ip
  type       = "A"
  proxied    = local.cloudflare_proxied
}

resource "kubernetes_service" "nodes" {
  metadata {
    name      = "nodes"
    namespace = var.nodes_namespace
  }
  spec {
    selector = {
      "app.kubernetes.io/name" = "xmtp-node"
    }
    port {
      name        = "http"
      port        = var.node_api_http_port
      target_port = var.node_api_http_port
    }
  }
}

resource "kubernetes_ingress_v1" "nodes" {
  metadata {
    name      = "nodes"
    namespace = var.nodes_namespace
    annotations = {
      "cert-manager.io/cluster-issuer" = "cert-manager"
    }
  }
  spec {
    ingress_class_name = var.ingress_class_name
    tls {
      hosts       = var.hostnames
      secret_name = "nodes-tls"
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
                name = kubernetes_service.nodes.metadata.0.name
                port {
                  number = kubernetes_service.nodes.spec.0.port.0.port
                }
              }
            }
          }
        }
      }
    }
  }
}
