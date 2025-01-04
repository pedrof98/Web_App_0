from pydantic import BaseModel, Field, ConfigDict
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


class MeasurementUpdate(MeasurementBase):
    sensor_id: Optional[int]
    timestamp: Optional[datetime]
    speed: Optional[float]
    vehicle_count: Optional[int]

    model_config = ConfigDict(from_attributes=True)


class MeasurementRead(MeasurementBase):
    id: int
    created_at: datetime

    model_config = ConfigDict(from_attributes=True)

class BatchMeasurementCreate(BaseModel):
    """
    For batch ingstion, in case of receiving multiple measurments in a single request.
    Each measurement adheres to MeasurementCreate
    """
    measurements: list[MeasurementCreate]

