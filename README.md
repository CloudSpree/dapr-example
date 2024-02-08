# DAPR local development example

This is an example repository for local development with DAPR targeting the Azure Container Apps environment, specifically using Service Bus queues. Therefore, the Redis Pub/Sub component has the somewhat unusual name `servicebus-queue`.

## Quick start

```bash
docker compose up --build
```

````
curl localhost:8080/send-email
````

## Azure infrastructure

If you're wondering how to configure the Azure counterpart, here are some examples. It does not include all the resources, but it's a good starting point, I believe.
Also, the role bound to the managed identity is not very fine-grained, but I'm sure you'll figure it out.

```terraform
resource "azurerm_servicebus_namespace" "this" {
  name                = "ctx-application-servicebus"
  location            = azurerm_resource_group.application.location
  resource_group_name = azurerm_resource_group.application.name
  sku                 = "Standard"
}

resource "azurerm_servicebus_queue" "emails" {
  namespace_id = azurerm_servicebus_namespace.this.id
  name                = "emails"
}

resource "azurerm_role_assignment" "application_servicebus_data_owner" {
  scope                = azurerm_servicebus_namespace.this.id
  role_definition_name = "Azure Service Bus Data Owner"
  // this is the managed identity used in Container Apps application
  principal_id         = azurerm_user_assigned_identity.container_apps.principal_id
}

// payload format: https://github.com/dapr/go-sdk/blob/main/service/common/type.go#L21
resource "azurerm_container_app_environment_dapr_component" "queue" {
  name                         = "servicebus-queue"
  container_app_environment_id = azurerm_container_app_environment.this.id
  component_type               = "pubsub.azure.servicebus.queues"
  version                      = "v1"
  metadata {
    name = "azureClientId"
    value = azurerm_user_assigned_identity.container_apps.client_id
  }

  metadata {
    name = "namespaceName"
    value = "${azurerm_servicebus_namespace.this.name}.servicebus.windows.net"
  }

  scopes = [
    "backend"
  ]
}

resource "azurerm_container_app" "backend" {
  name                         = "backend"
  container_app_environment_id = azurerm_container_app_environment.this.id
  resource_group_name          = azurerm_resource_group.application.name
  revision_mode                = "Single"

      identity {
        type = "SystemAssigned, UserAssigned"
        identity_ids = [
            azurerm_user_assigned_identity.container_apps.id,
        ]
      }

      registry {
        server = var.image_name_prefix
        identity = azurerm_user_assigned_identity.container_apps.id
      }



  template {
    min_replicas = 1
    max_replicas = 1

    container {
      name   = "app"
      image  = "some-image:some-tag"
      cpu    = 0.25
      memory = "0.5Gi"
    }
  }

  ingress {
    target_port = 8080
    external_enabled = true
    traffic_weight {
        latest_revision = true
        percentage = 100
    }
  }

  dapr {
    app_id = "backend"
    app_port = 8080
    app_protocol = "http"
  }
}
```
