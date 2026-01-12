#!/bin/bash

# Script untuk mendapatkan kubeconfig dari VPS
# Jalankan di VPS setelah k3s terinstall

echo "📋 Getting kubeconfig for GitHub Actions..."

# Copy kubeconfig
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config

# Fix server IP
sed -i 's/127.0.0.1/72.61.69.116/g' ~/.kube/config

# Encode to base64 for GitHub Secrets
echo "🔐 Base64 encoded kubeconfig (copy this to GitHub Secrets as KUBE_CONFIG):"
echo "----------------------------------------"
cat ~/.kube/config | base64 -w 0
echo ""
echo "----------------------------------------"

echo "✅ Done! Add the base64 string above to GitHub repository secrets as 'KUBE_CONFIG'"