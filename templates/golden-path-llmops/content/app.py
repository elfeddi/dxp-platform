import os
from fastapi import FastAPI
from pydantic import BaseModel
import httpx

app = FastAPI(
    title="${{ values.name }}",
    description="${{ values.description }}"
)

LITELLM_URL = os.getenv("LITELLM_URL", "http://litellm.llmops.svc.cluster.local:4000")
LITELLM_KEY = os.getenv("LITELLM_API_KEY", "")

class PromptRequest(BaseModel):
    prompt: str
    model: str = "gpt-4o-mini"

@app.post("/generate")
async def generate(req: PromptRequest):
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{LITELLM_URL}/v1/chat/completions",
            headers={"Authorization": f"Bearer {LITELLM_KEY}"},
            json={
                "model": req.model,
                "messages": [{"role": "user", "content": req.prompt}],
            },
            timeout=30,
        )
        return resp.json()

@app.get("/health")
def health():
    return {"status": "ok", "service": "${{ values.name }}"}
