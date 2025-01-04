from sqlalchemy import Column, Integer, String, Date, Float
from sqlalchemy.orm import relationship
from .base import Base

class Station(Base):
    __tablename__ = "stations"

    id = Column(Integer, primary_key=True, index=True)
    code = Column(String(50), unique=True, nullable=False)
    name = Column(String(100))
    city = Column(String(50))
    latitude = Column(Float)
    longitude = Column(Float)
    date_of_installation = Column(Date)

    sensors = relationship("Sensor", back_populates="station", cascade="all, delete-orphan")
    events = relationship("UserEvent", back_populates="station", cascade="all, delete-orphan")

