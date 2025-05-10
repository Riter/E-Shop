import asyncio
import json
import logging
import os
from typing import Dict
from uuid import UUID

from fastapi import FastAPI, HTTPException, status
from aiokafka import AIOKafkaProducer
import uvicorn

from models import (CreateItemRequest, CreateItemResponse,
                    ChangeItemRequest, ChangeItemResponse,
                    DeleteItemRequest, DeleteItemResponse,
                    Item, OperationType)

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("manage-item-crud")

BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
TOPIC = "item-events"
PARTITION_RAW = 0  # partition для «сырых» событий

app = FastAPI(title="ManageItem CRUD service")
_kafka_producer: AIOKafkaProducer | None = None
_memory_store: Dict[UUID, Item] = {}

class UUIDEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, UUID):
            # Convert UUID to string
            return str(obj)
        return super().default(obj)


# ---------- FastAPI event hooks --------------------------------------------
@app.on_event("startup")
async def startup_event():
    global _kafka_producer
    _kafka_producer = AIOKafkaProducer(
        bootstrap_servers=BOOTSTRAP_SERVERS,
        value_serializer=lambda v: json.dumps(v, cls=UUIDEncoder).encode("utf-8"),
    )
    await _kafka_producer.start()
    log.info("Kafka producer started")


@app.on_event("shutdown")
async def shutdown_event():
    if _kafka_producer:
        await _kafka_producer.stop()
        log.info("Kafka producer stopped")


# ---------- Helper ----------------------------------------------------------
async def _publish_event(payload: dict, partition: int = PARTITION_RAW) -> None:
    if not _kafka_producer:
        raise RuntimeError("Kafka producer not started")
    # отправляем в заданную партицию
    await _kafka_producer.send_and_wait(
        TOPIC,
        payload,
        partition=partition,
    )
    log.info("Published to partition %d: %s", partition, payload)


# ---------- Endpoints -------------------------------------------------------
@app.post("/items", response_model=CreateItemResponse)
async def create_item(req: CreateItemRequest):
    if req.operation_type != OperationType.CREATE:
        raise HTTPException(status.HTTP_400_BAD_REQUEST,
                            detail="operation_type must be 3 (create)")
    item = req.item
    _memory_store[item.id] = item
    await _publish_event({
        "operation_type": req.operation_type,
        "item_id": str(item.id),
        "item": item.model_dump()
    })
    return CreateItemResponse(status=200, item_id=item.id)


@app.put("/items/{item_id}", response_model=ChangeItemResponse)
async def change_item(item_id: UUID, req: ChangeItemRequest):
    if req.operation_type != OperationType.CHANGE:
        raise HTTPException(status.HTTP_400_BAD_REQUEST,
                            detail="operation_type must be 2 (change)")
    if item_id not in _memory_store:
        raise HTTPException(status.HTTP_404_NOT_FOUND, detail="item not found")

    _memory_store[item_id] = req.item.model_copy(update={"id": item_id})
    await _publish_event({
        "operation_type": req.operation_type,
        "item_id": str(item_id),
        "item": req.item.model_dump()
    })
    return ChangeItemResponse(status=200)


@app.delete("/items/{item_id}", response_model=DeleteItemResponse)
async def delete_item(item_id: UUID, req: DeleteItemRequest):
    if req.operation_type != OperationType.DELETE:
        raise HTTPException(status.HTTP_400_BAD_REQUEST,
                            detail="operation_type must be 1 (delete)")
    if item_id not in _memory_store:
        raise HTTPException(status.HTTP_404_NOT_FOUND, detail="item not found")

    _memory_store.pop(item_id)
    await _publish_event({
        "operation_type": req.operation_type,
        "item_id": str(item_id),
        "item": None
    })
    return DeleteItemResponse(status=200)


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
