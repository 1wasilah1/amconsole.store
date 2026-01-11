#!/bin/bash

# Deploy PS Rental Monitor to Kubernetes

echo "Deploying to Kubernetes..."

# Apply manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Wait for deployment
kubectl rollout status deployment/amconsole-app -n amconsole

echo "Deployment completed!"
echo "Access: https://amconsole.store"