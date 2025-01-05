from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
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
from app.core.kafka_config import KafkaClient, get_kafka_client
import asyncio

router = APIRouter()
kafka_client = None

@router.on_event("startup")
async def startup_event():
    pass # initialization handled by dependency injection

@router.on_event("shutdown")
async def shutdown_event():
    pass # cleanup handled by dependency injection




@router.post("/real-time", response_model=MeasurementRead)
async def ingest_real_time_data(
    measurement_in: MeasurementCreate,
    kafka: KafkaClient = Depends(get_kafka_client),
    db: Session = Depends(get_db)
    ):

    measurement = TrafficMeasurement(**measurement_in.model_dump())
    db.add(measurement)
    db.commit()
    db.refresh(measurement)

    # send to kafka
    await kafka.send_message(
            topic="traffic-measurements",
            value=measurement_in.model_dump(),
            key=str(measurement.sensor_id)
            )

    
    return measurement


@router.post("/batch")
async def ingest_batch_data(
        batch_in: BatchMeasurementCreate,
        kafka: KafkaClient = Depends(get_kafka_client),
        db: Session = Depends(get_db)
        ):

    measurements_to_create = []
    for m in batch_in.measurements:
        measurement = TrafficMeasurement(**m.model_dump())
        measurements_to_create.append(measurement)

        await kafka.send_message(
                topic="traffic-measurements",
                value=m.model_dump(),
                key=str(m.sensor_id)
                )

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


