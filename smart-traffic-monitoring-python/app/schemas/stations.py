from pydantic import BaseModel
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

class StationRead(StationBase):
    id: int

    class Config:
        from_attributes = True

