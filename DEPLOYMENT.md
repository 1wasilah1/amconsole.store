# VPS Deployment Guide

## Prerequisites
- VPS dengan IP: 72.61.69.116
- Domain amconsole.store pointing ke IP tersebut
- Root access ke VPS

## Step 1: Setup VPS
```bash
# SSH ke VPS
ssh root@72.61.69.116

# Upload dan jalankan setup script
chmod +x vps-setup.sh
./vps-setup.sh
```

## Step 2: Get Kubeconfig untuk GitHub Actions
```bash
# Di VPS, jalankan:
chmod +x get-kubeconfig.sh
./get-kubeconfig.sh

# Copy output base64 ke GitHub repository secrets sebagai 'KUBE_CONFIG'
```

## Step 3: Setup GitHub Secrets
Di GitHub repository, tambahkan secrets:
- `KUBE_CONFIG`: Base64 encoded kubeconfig dari step 2
- `DOCKER_USERNAME`: Docker Hub username
- `DOCKER_PASSWORD`: Docker Hub password/token

## Step 4: Deploy
Push ke branch `main` untuk trigger auto-deploy, atau manual:
```bash
./deploy.sh
```

## Troubleshooting
```bash
# Check cluster status
./troubleshoot.sh

# Check specific issues
kubectl describe ingress amconsole-ingress -n amconsole
kubectl logs -n traefik-system deployment/traefik
```

## Architecture
```
Internet → Domain (amconsole.store) → VPS (72.61.69.116)
    ↓
k3s Cluster
    ├── MetalLB (LoadBalancer)
    ├── Traefik (Ingress + SSL)
    └── Application (amconsole)
```

## URLs
- **Production**: https://amconsole.store
- **Admin**: https://amconsole.store/admin

## Key Components
1. **k3s**: Lightweight Kubernetes
2. **MetalLB**: LoadBalancer untuk bare metal
3. **Traefik**: Ingress controller + SSL termination
4. **Let's Encrypt**: Automatic SSL certificates