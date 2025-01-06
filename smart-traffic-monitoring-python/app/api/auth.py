from fastapi import APIRouter, Depends, HTTPException
from fastapi.security import OAuth2PasswordRequestForm
from app.core.auth import create_access_token

router = APIRouter()

@router.post("/")
async def login(form_data: OAuth2PasswordRequestForm = Depends()):
    user = authenticate_user(form_data.username, form_data.password)
    if not user:
        raise HTTPException(status_code=400, detail="Incorrect email or password")

    access_token = create_access_toekn(
            data={"sub": str(user.id), "role": user.role}
            )
    return {"access_token": access_token, "token_type": "bearer"}

