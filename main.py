from fastapi import APIRouter, FastAPI

app = FastAPI()
router = APIRouter()

@router.get("/")
async def hello():
    return {"hello": "hello"}

app.include_router(router)


