# ${{ values.name }}

${{ values.description }}

## Stack
- Langage : ${{ values.language }}
- CI/CD : Tekton → Harbor → ArgoCD
- Plateforme : DxP

## Démarrage rapide
```bash
# Cloner le repo
git clone <repo-url>
cd ${{ values.name }}

# Lancer en local
docker build -t ${{ values.name }} .
docker run -p 8080:8080 ${{ values.name }}
```
