from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models.stations import Station                     
from app.schemas.stations import StationRead, StationCreate


router = APIRouter()

@router.get("/", response_model=List[StationRead])
def get_stations(db: Session = Depends(get_db)):
    return db.query(Station).all()


@router.post("/", response_model=StationRead)
def create_station(
        station_in: StationCreate,
        db: Session = Depends(get_db)
        ):
    station = Station(**station_in.dict())
    db.add(station)
    db.commit()
    db.refresh(station)
    return station


@router.get("/{station_id}", response_model=StationRead)
def get_station(
        station_id: int,
        db: Session = Depends(get_db)
        ):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")
    return station


@router.put("/{station_id}", response_model=StationRead)
def update_station(
        station_id: int,
        station_in: StationCreate,
        db: Session = Depends(get_db)
        ):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")

    for field, value in station_in.dict(exclude_unset=True).items():
        setattr(station, field, value)

    db.commit()
    db.refresh(station)
    return station


@router.delete("/{station_id}")
def delete_station(
        station_id: int,
        db: Session = Depends(get_db)
        ):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")

    db.delete(station)
    db.commit()
    return {"message": "Station deleted successfully"}


