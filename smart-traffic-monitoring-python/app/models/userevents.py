from sqlalchemy import Column, Integer, String, Date, Text
from app.models.base import Base

class UserEvent(Base):
    __tablename__ = "user_events"

    id = Column(Integer, primary_key=True)
    date = Column(Date, nullable=False)
    city = Column(String(50), nullable=False)
    event_type = Column(String(50), nullable=False)
    description = Column(Text, nullable=True)
    expected_congestion_level = Column(String(10), nullable=True)

