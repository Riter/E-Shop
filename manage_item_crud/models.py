from enum import IntEnum
from uuid import UUID, uuid4
from pydantic import BaseModel, Field
from typing import Optional


class OperationType(IntEnum):
    DELETE = 1
    CHANGE = 2
    CREATE = 3


class Item(BaseModel):
    id: Optional[UUID] = Field(default_factory=uuid4)
    name: str
    description: Optional[str] = None
    price: float
    category: str


# --- HTTP DTOs --------------------------------------------------------------

class CreateItemRequest(BaseModel):
    operation_type: OperationType
    item: Item


class CreateItemResponse(BaseModel):
    status: int
    item_id: UUID


class ChangeItemRequest(BaseModel):
    operation_type: OperationType
    item: Item


class ChangeItemResponse(BaseModel):
    status: int


class DeleteItemRequest(BaseModel):
    operation_type: OperationType
    item_id: UUID


class DeleteItemResponse(BaseModel):
    status: int
