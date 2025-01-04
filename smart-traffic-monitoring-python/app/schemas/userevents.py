from pydantic import BaseModel, ConfigDict
from typing import Optional 
from datetime import date

class UserEventBase(BaseModel):
    date: date
    city: str
    event_type: str
    description: Optional[str] = None
    expected_congestion_level: Optional[str] = None
    station_id: int


class UserEventCreate(UserEventBase):
    "complete later on if needed"
    pass


class UserEventUpdate(BaseModel):
    date: Optional[date] = None
    city: Optional[str] = None
    event_type: Optional[str] = None
    description: Optional[str] = None
    expected_congestion_level: Optional[str] = None
    station_id: Optional[int] = None

    model_config = ConfigDict(from_attributes=True)

class UserEventRead(UserEventBase):
    id: int

    model_config = ConfigDict(from_attributes=True)


