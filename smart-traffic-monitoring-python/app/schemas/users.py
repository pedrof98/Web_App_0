
from pydantic import BaseModel, EmailStr, ConfigDict
from typing import Optional
from app.models.users import UserRole

class UserBase(BaseModel):
    email: EmailStr
    role: UserRole


class UserCreate(UserBase):
    password: str


class UserRead(UserBase):
    id: int

    model_config = ConfigDict(from_attributes=True)


class UserUpdate(BaseModel):
    email: Optional[EmailStr] = None
    password: Optional[str] = None
    role: Optional[UserRole] = None


    model_config = ConfigDict(from_attributes=True)

