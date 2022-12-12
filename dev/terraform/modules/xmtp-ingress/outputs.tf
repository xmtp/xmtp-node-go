output "namespace" {
  value = kubernetes_namespace.namespace.metadata.0.name
}

output "external_ip" {
  value = one(data.kubernetes_service.traefik.status.0.load_balancer.0.ingress[*].ip)
}
