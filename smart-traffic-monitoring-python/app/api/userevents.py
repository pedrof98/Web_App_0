from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models.userevents import UserEvent
from app.schemas.userevents import UserEventRead, UserEventCreate, UserEventUpdate
from app.models.stations import Station
from app.core.auth import get_current_user
from app.models.users import User, UserRole



router = APIRouter()

@router.post("/", response_model=UserEventRead)
def create_user_event(
        event_in: UserEventCreate,
        db: Session = Depends(get_db),
        current_user: User = Depends(get_current_user)
        ):
    # Optional: verify station exists
    station = db.query(Station).filter(Station.id == event_in.station_id).first()
    if not station:
        raise HTTPException(status_code=400, detail="Station does not exist.")
    
    new_event = UserEvent(**event_in.model_dump())
    db.add(new_event)
    db.commit()
    db.refresh(new_event)
    return new_event


@router.get("/{event_id}", response_model=UserEventRead)
def get_event_by_id(
        event_id: int,
        db: Session = Depends(get_db)
        ):
    event = db.query(UserEvent).filter(UserEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Event not found.")
    return event


@router.put("/{event_id}", response_model=UserEventRead)
def update_event(
        event_id: int,
        event_in: UserEventUpdate,
        db: Session = Depends(get_db),
        current_user: User = Depends(get_current_user)
        ):
    event = db.query(UserEvent).filter(UserEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")

    for field, value in event_in.model_dump(exclude_unset=True).items():
        setattr(event, field, value)

    db.commit()
    db.refresh(event)
    return event

@router.get("/", response_model=List[UserEventRead])
def get_all_user_events(db: Session = Depends(get_db)):
    return db.query(UserEvent).all()

@router.delete("/{event_id}")
def delete_event(
        event_id: int,
        db: Session = Depends(get_db),
        current_user: User = Depends(get_current_user)
        ):
    event = db.query(UserEvent).filter(UserEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")

    db.delete(event)
    db.commit()
    return {"message": "User event deleted successfully"}

