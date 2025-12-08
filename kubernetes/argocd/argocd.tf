resource "kubernetes_namespace_v1" "argocd" {
  metadata {
    name = var.argocd_namespace
  }
}

resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = var.argocd_chart_version
  namespace        = kubernetes_namespace_v1.argocd.metadata[0].name
  create_namespace = false
  wait             = true
  timeout          = 600

  values = [
    yamlencode({
      global = {
        domain = "" # Configure if you have a domain
      }

      configs = {
        params = {
          "server.insecure" = true # Set to false if using TLS
        }
      }

      # Run on system nodes with tolerations
      controller = {
        tolerations = [{
          key      = "node-type"
          operator = "Equal"
          value    = "system"
          effect   = "NoSchedule"
        }]
        nodeSelector = {
          "node-type" = "system"
        }
      }

      server = {
        tolerations = [{
          key      = "node-type"
          operator = "Equal"
          value    = "system"
          effect   = "NoSchedule"
        }]
        nodeSelector = {
          "node-type" = "system"
        }
        service = {
          type = "ClusterIP"
        }
      }

      repoServer = {
        tolerations = [{
          key      = "node-type"
          operator = "Equal"
          value    = "system"
          effect   = "NoSchedule"
        }]
        nodeSelector = {
          "node-type" = "system"
        }
      }

      applicationSet = {
        tolerations = [{
          key      = "node-type"
          operator = "Equal"
          value    = "system"
          effect   = "NoSchedule"
        }]
        nodeSelector = {
          "node-type" = "system"
        }
      }

      notifications = {
        tolerations = [{
          key      = "node-type"
          operator = "Equal"
          value    = "system"
          effect   = "NoSchedule"
        }]
        nodeSelector = {
          "node-type" = "system"
        }
      }

      redis = {
        tolerations = [{
          key      = "node-type"
          operator = "Equal"
          value    = "system"
          effect   = "NoSchedule"
        }]
        nodeSelector = {
          "node-type" = "system"
        }
      }

      dex = {
        enabled = false
      }
    })
  ]
}
