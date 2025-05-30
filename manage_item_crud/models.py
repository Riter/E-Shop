from enum import IntEnum
from uuid import UUID, uuid4
from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime


class OperationType(IntEnum):
    DELETE = 1
    CHANGE = 2
    CREATE = 3


class Item(BaseModel):
    id: int
    name: str
    description: str
    price: float
    category: str
    created_at: datetime
    images: List[str]


# --- HTTP DTOs --------------------------------------------------------------

class CreateItemRequest(BaseModel):
    operation_type: OperationType
    item: Item


class CreateItemResponse(BaseModel):
    status: int
    item_id: int


class ChangeItemRequest(BaseModel):
    operation_type: OperationType
    item: Item


class ChangeItemResponse(BaseModel):
    status: int


class DeleteItemRequest(BaseModel):
    operation_type: OperationType
    item_id: int


class DeleteItemResponse(BaseModel):
    status: int
