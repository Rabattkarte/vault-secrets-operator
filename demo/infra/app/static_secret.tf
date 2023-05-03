# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "kubernetes_secret" "static" {
  metadata {
    name      = "vso-static-demo"
    namespace = kubernetes_namespace.dev.metadata[0].name
  }
}

resource "kubernetes_manifest" "vault-static-secret" {
  manifest = {
    apiVersion = "secrets.hashicorp.com/v1alpha1"
    kind       = "VaultStaticSecret"
    metadata = {
      name      = "vso-static-demo"
      namespace = kubernetes_namespace.dev.metadata[0].name
    }
    spec = {
      namespace    = vault_auth_backend.default.namespace
      type         = "kv-v2"
      mount        = vault_kv_secret_v2.static.mount
      name         = vault_kv_secret_v2.static.name
      refreshAfter = "30s"
      destination = {
        create : false
        name : kubernetes_secret.static.metadata[0].name
      }
      # rolloutRestartTargets = [
      #   {
      #     kind = "Deployment"
      #     name = "vso-db-demo"
      #   }
      # ]
    }
  }
}

resource "vault_policy" "static" {
  namespace = local.namespace
  name      = "${local.auth_policy}-static"
  policy    = <<EOT
path "secret/*" {
  capabilities = ["read"]
}
EOT
}

resource "vault_kv_secret_v2" "static" {
  mount               = "secret"
  name                = "static"
  cas                 = 1
  delete_all_versions = true
  data_json = jsonencode(
    {
      foo = "bar"
    }
  )
}
