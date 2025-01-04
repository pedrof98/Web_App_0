from fastapi import FastAPI
import uvicorn
from .api.api import router as traffic_router

app = FastAPI()

app.include_router(traffic_router)

@app.get("/")
def read_root():
    return {"message": "Welcome to the Smart Traffic Monitoring API"}


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8900)
