Запуск через консоль:
 
PYTHONPATH=src uv run python scripts/analyze_dataset.py \
  --dataset data/demo/dataset.csv \
  --images-dir data/demo/images \
  --output-dir outputs/demo



Запуск через fastapi:

uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8001