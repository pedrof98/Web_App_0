from sqlalchemy import Column, Integer, Float, DateTime, ForeignKey
from sqlalchemy.orm import relationship
from datetime import datetime
from .base import Base

class TrafficMeasurement(Base):
    __tablename__ = "traffic_measurements"

    id = Column(Integer, primary_key=True)
    sensor_id = Column(Integer, ForeignKey("sensors.id"), nullable=False)
    timestamp = Column(DateTime, nullable=False)
    speed = Column(Float, nullable=True)
    vehicle_count = Column(Integer, nullable=True)
    created_at = Column(
            DateTime,
            nullable=False,
            default=datetime.utcnow
            )

    sensor = relationship("Sensor", back_populates="measurements")

