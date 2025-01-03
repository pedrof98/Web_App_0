from sqlalchemy import Column, Integer, String, ForeignKey
from sqlalchemy.orm import relationship
from .base import Base

class Sensor(Base):
    __tablename__ = "sensors"

    id = Column(Integer, primary_key=True)
    sensor_id = Column(String(50), unique=True, nullable=False)
    station_id = Column(Integer, ForeignKey("stations.id"), nullable=False)
    measurement_type = Column(String(50))
    status = Column(String(20), default="active")

    station = relationship("Station", back_populates="sensors")
    measurements = relationship("TrafficMeasurement", back_populates="sensor", cascade="all, delete-orphan")