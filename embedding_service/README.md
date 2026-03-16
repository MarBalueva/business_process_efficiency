# Embedding Service (Local)

## Run locally

```bash
cd embedding_service
python -m venv .venv
. .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -r requirements.txt
uvicorn app:app --host 0.0.0.0 --port 8001
```

## Endpoints

- `GET /health`
- `POST /embed`

Request:

```json
{
  "texts": ["Согласование заявки", "Утверждение запроса"]
}
```

Response:

```json
{
  "vectors": [[0.01, -0.02, ...], [...]],
  "dim": 1024
}
```
