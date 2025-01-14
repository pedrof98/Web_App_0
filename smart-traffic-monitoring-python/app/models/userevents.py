from sqlalchemy import Column, Integer, String, Date, Text, ForeignKey
from sqlalchemy.orm import relationship
from .base import Base

class UserEvent(Base):
    __tablename__ = "user_events"

    id = Column(Integer, primary_key=True)
    date = Column(Date, nullable=False)
    city = Column(String(50), nullable=False)
    event_type = Column(String(50), nullable=False)
    description = Column(Text, nullable=True)
    expected_congestion_level = Column(String(10), nullable=True)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=True)

    station = relationship("Station", back_populates="events")