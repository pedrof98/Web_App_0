from fastapi import APIRouter, Depends, HTTPException
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from app.core.auth import create_access_token, authenticate_user, get_user_by_email
from app.core.security import verify_password
from app.models.users import User
from app.database import get_db

router = APIRouter()

@router.post("/")
async def login(form_data: OAuth2PasswordRequestForm = Depends(), db: Session = Depends(get_db)):

    user = get_user_by_email(db, form_data.username)
    if not user or not verify_password(form_data.password, user.hashed_password):
        raise HTTPException(status_code=400, detail="Incorrect email or password")

    access_token = create_access_token(
            data={"sub": str(user.id), "role": user.role}
            )
    return {"access_token": access_token, "token_type": "bearer"}


