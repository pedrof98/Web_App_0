from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models.measurements import TrafficMeasurement
from app.schemas.measurements import (
        MeasurementCreate,
        MeasurementRead,
        BatchMeasurementCreate,
        MeasurementUpdate
        )

router = APIRouter()

@router.post("/real-time", response_model=MeasurementRead)
def ingest_real_time_data(
        measurement_in: MeasurementCreate,
        db: Session = Depends(get_db)
        ):
    measurement = TrafficMeasurement(**measurement_in.model_dump())
    db.add(measurement)
    db.commit()
    db.refresh(measurement)
    return measurement


@router.post("/batch")
def ingest_batch_data(
        batch_in: BatchMeasurementCreate,
        db: Session = Depends(get_db)
        ):

    measurements_to_create = [
            TrafficMeasurement(**m.model_dump()) for m in batch_in.measurements
            ]
    db.bulk_save_objects(measurements_to_create)
    db.commit()

    return {"message": f"Ingested {len(measurements_to_create)} measurements successfully."}

@router.get("/", response_model=List[MeasurementRead])
def get_all_measurements(db: Session = Depends(get_db)):

    return db.query(TrafficMeasurement).all()


@router.get("/{measurement_id}", response_model=MeasurementRead)
def get_measurement_by_id(measurement_id: int, db: Session = Depends(get_db)):
    measurement = db.query(TrafficMeasurement).filter(TrafficMeasurement.id == measurement_id).first()
    if not measurement:
        raise HTTPException(status_code=404, detail="Measurement not found")
    return measurement


