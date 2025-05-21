from fastapi import APIRouter
from app.api import stations, sensors, data, userevents, auth, users

router = APIRouter()
router.include_router(stations.router, prefix="/stations", tags=["stations"])
router.include_router(sensors.router, prefix="/sensors", tags=["sensors"])
router.include_router(data.router, prefix="/data", tags=["data"])
router.include_router(userevents.router, prefix="/events", tags=["events"])
router.include_router(auth.router, prefix="/token", tags=["token"])
router.include_router(users.router, prefix="/users", tags=["users"])
