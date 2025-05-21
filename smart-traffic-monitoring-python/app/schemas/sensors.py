from pydantic import BaseModel, ConfigDict
from typing import Optional

class SensorBase(BaseModel):
    sensor_id: str
    measurement_type: Optional[str] = None
    status: Optional[str] = "active"
    station_id: int


class SensorCreate(SensorBase):
    """If there's anything required on creation i will add it here"""
    pass


class SensorUpdate(BaseModel):
    sensor_id: Optional[str] = None
    measurement_type: Optional[str] = None
    status: Optional[str] = None
    station_id: Optional[int] = None

    model_config = ConfigDict(from_attributes=True)
class SensorRead(SensorBase):
    id: int

    model_config = ConfigDict(from_attributes=True)

