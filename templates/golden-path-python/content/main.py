from fastapi import FastAPI
import os

app = FastAPI(title="${{ values.name }}", description="${{ values.description }}")

@app.get("/")
def root():
    return {"service": "${{ values.name }}", "status": "ok", "platform": "DxP"}

@app.get("/health")
def health():
    return {"status": "ok"}
