from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional

class MeasurementBase(BaseModel):
    sensor_id: int
    timestamp: datetime
    speed: Optional[float] = None
    vehicle_count: Optional[int] = None


class MeasurementCreate(MeasurementBase):
    """optional but might add in the future"""
    pass


class MeasurementRead(MeasurementBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True

class BatchMeasurementCreate(BaseModel):
    """
    For batch ingstion, in case of receiving multiple measurments in a single request.
    Each measurement adheres to MeasurementCreate
    """
    measurements: list[MeasurementCreate]

