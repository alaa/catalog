# Catalog
Easy secrets integration with Vault and Marathon.

# Build

```
docker build --no-cache -t catalog:dev .
```

# Usage

Assuming that you have `vault` running

## Run catalog service:

```
docker run -it                             \
           -e VAULT_ADDR="$VAULT_ADDR"     \
           -e VAULT_TOKEN="$VAULT_TOKEN"   \
           -e MARATHON_URL="$MARATHON_URL" \
           -p 8080:8080                    \
           catalog:dev
```

## Create your service file:
```
{
  "id": "nginx",
  "cpus": 0.1,
  "mem": 128.0,
  "instances": 1,
  "container": {
    "type": "DOCKER",
    "docker": {
      "image": "nginx:latest",
      "forcePullImage": true,
      "network": "BRIDGE",
      "portMappings": [{
        "containerPort": 80,
        "hostPort": 0,
        "protocol": "tcp"
      }]
    }
  },
  "env": {
    "SERVICE_NAME": "nginx"
  },
  "healthChecks": [{
    "protocol": "HTTP",
    "path": "/",
    "intervalSeconds": 30
  }]
}
```

## Deploy the service using catalog:

```
#!/bin/bash
set -e

export CATALOG_ADDR="http://127.0.0.1:8080"
export VAULT_ENDPOINT="nginx"

curl --show-error --silent --fail       \
     -H "X-Vault-Token: ${VAULT_TOKEN}" \
     -X POST                            \
     -d@nginx.json                      \
     ${CATALOG_ADDR}/service/${VAULT_ENDPOINT}

```
