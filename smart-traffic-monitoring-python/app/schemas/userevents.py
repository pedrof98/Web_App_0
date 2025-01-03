from pydantic import BaseModel
from typing import Optional 
from datetime import date

class UserEventBase(BaseModel):
    date: date
    city: str
    event_type: str
    description: Optional[str] = None
    expected_congestion_level: Optional[str] = None


class UserEventCreate(UserEventBase):
    "complete later on if needed"
    pass


class UserEventRead(UserEventBase):
    id: int

    class Config:
        from_attributes = True


