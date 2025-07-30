from fastapi import APIRouter, FastAPI
from fastapi.staticfiles import StaticFiles

app = FastAPI()
router = APIRouter()

@router.post("/")
async def hello():
    return {"hello": "hello"}

app.include_router(router)
app.mount("/", StaticFiles(directory="frontend", html=True), name="frontend")
