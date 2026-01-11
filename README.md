# PS Rental Monitor

Aplikasi monitoring rental PlayStation dengan Go dan PostgreSQL.

## Deployment

### Kubernetes
```bash
./deploy.sh
```

### Manual Deploy
```bash
kubectl apply -f k8s/
```

## Access
- **Production**: https://amconsole.store
- **Admin**: https://amconsole.store/admin
- **Public**: https://amconsole.store/

## Features
- Admin login dan management
- Real-time monitoring dengan WebSocket
- Auto-deploy dengan GitHub Actions
- HTTPS dengan Let's Encrypt
- Kubernetes deployment

## Development
```bash
go run main.go
```