from pydantic import BaseModel, ConfigDict
from typing import Optional
from datetime import date

class StationBase(BaseModel):
    code: str
    name: Optional[str]
    city: Optional[str]
    latitude: Optional[float]
    longitude: Optional[float]
    date_of_installation: Optional[date]


class StationCreate(StationBase):
    # this is for anything required upon creation
    pass


class StationUpdate(BaseModel):
    # this is for anything required upon update
    code: Optional[str] = None
    name: Optional[str] = None
    city: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float]  = None
    date_of_installation: Optional[date] = None

    model_config = ConfigDict(from_attributes=True)

class StationRead(StationBase):
    id: int

    model_config = ConfigDict(from_attributes=True)

