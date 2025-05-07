#!/bin/bash

set -e

echo "1. 切換到 Minikube Docker 環境"
eval $(minikube docker-env)

echo "2. 建立 Docker images"
docker build -t api-gateway:latest -f api-gateway/Dockerfile .
docker build -t redirect-service:latest -f redirect-service/Dockerfile .
docker build -t shorten-service:latest -f shorten-service/Dockerfile .

echo "3. 建立 Kubernetes namespaces"
kubectl apply -f k8s/namespace.yaml
kubectl create namespace monitoring || true

echo "4. 部署 Redis & Postgres"
kubectl apply -f k8s/redis-deployment.yaml -n url-shortener
kubectl apply -f k8s/redis-service.yaml -n url-shortener
kubectl apply -f k8s/postgres-deployment.yaml -n url-shortener
kubectl apply -f k8s/postgres-service.yaml -n url-shortener

echo "5. 部署 shorten-service"
kubectl apply -f shorten-service/k8s/shorten-deployment.yaml -n url-shortener
kubectl apply -f shorten-service/k8s/shorten-service.yaml -n url-shortener

echo "6. 部署 redirect-service"
kubectl apply -f redirect-service/k8s/redirect-deployment.yaml -n url-shortener
kubectl apply -f redirect-service/k8s/redirect-service.yaml -n url-shortener

echo "7. 部署 api-gateway"
kubectl apply -f api-gateway/k8s/api-gateway-deployment.yaml -n url-shortener
kubectl apply -f api-gateway/k8s/api-gateway-service.yaml -n url-shortener

echo "8. 部署 api-gateway 的 ServiceMonitor"
kubectl apply -f api-gateway/k8s/api-gateway-monitor.yaml -n monitoring

echo "9. 部署 Prometheus & Grafana"
kubectl apply -f monitoring/prometheus-deployment.yaml -n monitoring
kubectl apply -f monitoring/prometheus-service.yaml -n monitoring
kubectl apply -f monitoring/grafana-deployment.yaml -n monitoring
kubectl apply -f monitoring/grafana-service.yaml -n monitoring

echo "✅ 完成！請用以下指令查看狀態："
echo "kubectl get pods -n url-shortener"
echo "kubectl get pods -n monitoring"