from fastapi import FastAPI

from app.api.api import router as traffic_router
import asyncio

app = FastAPI()

app.include_router(traffic_router)

@app.get("/")
def read_root():
    return {"message": "Welcome to the Smart Traffic Monitoring API"}


def start():
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8900, loop="asyncio")

if __name__ == "__main__":
    start()
