output "argocd_namespace" {
  description = "ArgoCD namespace"
  value       = kubernetes_namespace_v1.argocd.metadata[0].name
}

output "argocd_server_service" {
  description = "ArgoCD server service name"
  value       = "argocd-server"
}

output "cluster_name" {
  description = "EKS cluster name (from cluster state)"
  value       = local.cluster_name
}

output "region" {
  description = "AWS region (from cluster state)"
  value       = local.region
}
