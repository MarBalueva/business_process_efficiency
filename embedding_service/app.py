from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

MODEL_NAME = "BAAI/bge-m3"

app = FastAPI(title="Embedding Service", version="1.0.0")
model = SentenceTransformer(MODEL_NAME)


class EmbedRequest(BaseModel):
    texts: list[str]


@app.get("/health")
def health():
    return {"ok": True, "model": MODEL_NAME}


@app.post("/embed")
def embed(req: EmbedRequest):
    if not req.texts:
        return {"vectors": [], "dim": 0}

    try:
        vectors = model.encode(req.texts, normalize_embeddings=True).tolist()
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc

    dim = len(vectors[0]) if vectors else 0
    return {"vectors": vectors, "dim": dim}

