from pydantic import BaseModel
from typing import Optional

class SensorBase(BaseModel):
    sensor_id: str
    measurement_type: Optional[str] = None
    status: Optional[str] = "active"
    station_id: int


class SensorCreate(SensorBase):
    """If there's anything required on creation i will add it here"""
    pass

class SensorRead(SensorBase):
    id: int

    class Config:
        from_attributes = True

