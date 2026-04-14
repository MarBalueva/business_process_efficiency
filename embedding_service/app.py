from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

MODEL_NAME = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"

app = FastAPI(title="Embedding Service", version="1.0.0")
model = None


class EmbedRequest(BaseModel):
    texts: list[str]


@app.on_event("startup")
def load_model():
    global model
    print(f"Loading embedding model: {MODEL_NAME}", flush=True)
    model = SentenceTransformer(MODEL_NAME)
    print("Embedding model loaded", flush=True)


@app.get("/health")
def health():
    return {"ok": True, "model": MODEL_NAME, "loaded": model is not None}


@app.post("/embed")
def embed(req: EmbedRequest):
    if not req.texts:
        return {"vectors": [], "dim": 0}

    if model is None:
        raise HTTPException(status_code=503, detail="model is not loaded yet")

    try:
        vectors = model.encode(req.texts, normalize_embeddings=True).tolist()
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc

    dim = len(vectors[0]) if vectors else 0
    return {"vectors": vectors, "dim": dim}
