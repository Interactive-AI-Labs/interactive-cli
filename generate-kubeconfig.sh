#!/bin/bash

set -e

echo "🔧 Generating kubeconfig for Docker container..."

# Get current context
CONTEXT_NAME=$(kubectl config current-context 2>/dev/null || echo "")
if [ -z "$CONTEXT_NAME" ]; then
    echo "❌ Error: No current kubectl context found"
    echo "   Run 'kubectl config use-context <context-name>' first"
    exit 1
fi

echo "📍 Current context: $CONTEXT_NAME"

# Get the cluster name from the context
CLUSTER_NAME=$(kubectl config view -o jsonpath="{.contexts[?(@.name=='$CONTEXT_NAME')].context.cluster}")

if [ -z "$CLUSTER_NAME" ]; then
    echo "❌ Error: Could not find cluster for context $CONTEXT_NAME"
    exit 1
fi

echo "📍 Cluster: $CLUSTER_NAME"

# Get cluster details
CLUSTER_SERVER=$(kubectl config view -o jsonpath="{.clusters[?(@.name=='$CLUSTER_NAME')].cluster.server}")

if [ -z "$CLUSTER_SERVER" ]; then
    echo "❌ Error: Could not extract cluster server"
    exit 1
fi

# Get access token
echo "🔑 Getting access token..."
TOKEN=$(gcloud auth print-access-token 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
    echo "❌ Error: Could not get access token"
    echo "   Run 'gcloud auth login' first"
    exit 1
fi

# Create tmp directory
mkdir -p tmp

# Remove old kubeconfig if it exists
rm -f tmp/kubeconfig-docker

# Create token-based kubeconfig for GKE
cat > tmp/kubeconfig-docker << EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: $CLUSTER_SERVER
  name: docker-cluster
contexts:
- context:
    cluster: docker-cluster
    user: docker-user
  name: docker-context
current-context: docker-context
users:
- name: docker-user
  user:
    token: $TOKEN
EOF

echo "✅ Created tmp/kubeconfig-docker"
echo "⚠️  Note: Token expires in 1 hour. Re-run to refresh."
