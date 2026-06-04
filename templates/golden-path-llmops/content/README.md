# ${{ values.name }}

${{ values.description }}

## Stack
- Python · FastAPI · LiteLLM Gateway
- Traçabilité : OpenTelemetry · DxP LLMOps
- FinOps : suivi tokens par service

## Commandes
```bash
pip install -r requirements.txt
uvicorn app:app --reload
```

## Variables d'environnement
- `LITELLM_URL` : URL de la gateway LiteLLM
- `LITELLM_API_KEY` : clé API LiteLLM
