# DxP — Engineering Platform as a Service

## Structure du repo

| Dossier | Description |
|---------|-------------|
| `stack0/` | CLI Go C1→C6 — cœur de DxP |
| `dxp-portal/` | Backstage — Developer Portal |
| `infrastructure/scripts/` | Scripts d installation (dxp-install.sh) |
| `infrastructure/ingress/` | Ressources Ingress HTTPS |
| `infrastructure/configs/` | Values Helm par composant |
| `infrastructure/crds/` | CRDs Kubernetes DxP |
| `templates/` | Golden Paths (DevOps, DataOps, MLOps) |
| `workloads/` | Apps déployées via DxP |
| `docs/` | Documentation architecture et sessions |

## Démarrage rapide

```bash
cd infrastructure/scripts
cp .env.example .env
vi .env  # remplir les valeurs
./dxp-install.sh
```

## URLs d accès (après installation)

| Service | URL |
|---------|-----|
| ArgoCD | https://DXP_IP:9443/argocd |
| Grafana | https://DXP_IP:9443/grafana |
| Harbor | http://DXP_IP:9091 |
| Tekton | http://DXP_IP:9295 |
| Dex SSO | https://DXP_IP:32000 |
| Airflow | http://DXP_IP:9294 |
