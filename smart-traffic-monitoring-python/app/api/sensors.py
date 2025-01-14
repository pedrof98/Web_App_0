from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from app.database import get_db
from app.models.sensors import Sensor
from app.schemas.sensors import SensorRead, SensorCreate, SensorUpdate
from app.models.users import User
from app.core.auth import get_current_user, check_admin_role

router = APIRouter()

@router.get("/", response_model=List[SensorRead])
def get_sensors(db: Session = Depends(get_db),
                current_user: User = Depends(get_current_user)):
    return db.query(Sensor).all()


@router.post("/", response_model=SensorRead)
def create_sensor(sensor_in: SensorCreate,
                   db: Session = Depends(get_db),
                   current_user: User = Depends(check_admin_role)):
    sensor = Sensor(**sensor_in.model_dump())
    db.add(sensor)
    db.commit()
    db.refresh(sensor)
    return sensor


@router.get("/{sensor_id}", response_model=SensorRead)
def get_sensor(sensor_id: int, db: Session = Depends(get_db)):
    sensor = db.query(Sensor).filter(Sensor.id == sensor_id).first()
    if not sensor:
        raise HTTPException(status_code=404, detail="Sensor not found")
    return sensor

@router.put("/{sensor_id}", response_model=SensorRead)
def update_sensor(sensor_id: int,
                sensor_in: SensorUpdate,
                db: Session = Depends(get_db),
                current_user: User = Depends(check_admin_role)):
    sensor = db.query(Sensor).filter(Sensor.id == sensor_id).first()
    if not sensor:
        raise HTTPException(status_code=404, detail="Sensor not found")

    for field, value in sensor_in.model_dump(exclude_unset=True).items():
        setattr(sensor, field, value)

    db.commit()
    db.refresh(sensor)
    return sensor


@router.delete("/{sensor_id}")
def delete_sensor(sensor_id: int,
                db: Session = Depends(get_db),
                current_user: User = Depends(check_admin_role)):
    sensor = db.query(Sensor).filter(Sensor.id == sensor_id).first()
    if not sensor:
        raise HTTPException(status_code=404, detail="Sensor not found")

    db.delete(sensor)
    db.commit()
    return {"message": "Sensor deleted successfully"}


