from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from app.core.auth import check_admin_role
from app.core.security import get_password_hash
from app.models.users import User, UserRole
from app.database import get_db
from app.schemas.users import UserCreate, UserRead

router = APIRouter()

@router.post("/", response_model=UserRead, status_code=status.HTTP_201_CREATED)
async def create_user(
        user_in: UserCreate,
        db: Session = Depends(get_db),
        current_user: User = Depends(check_admin_role)
        ):
        existing_user = db.query(User).filter(User.email == user_in.email).first()
        if existing_user:
                raise HTTPException(
                        status_code=status.HTTP_400_BAD_REQUEST,
                        detail="Email already registered"
                        )


        db_user = User(
                email=user_in.email,
                hashed_password=get_password_hash(user_in.password),
                role=user_in.role
                )
        db.add(db_user)
        db.commit()
        db.refresh(db_user)
        return db_user

