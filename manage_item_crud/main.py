import asyncio
import json
import logging
import os
from typing import Dict
import threading
import time
from datetime import datetime

from fastapi import FastAPI, HTTPException, status
from aiokafka import AIOKafkaProducer
import uvicorn

from models import (CreateItemRequest, CreateItemResponse,
                    ChangeItemRequest, ChangeItemResponse,
                    DeleteItemRequest, DeleteItemResponse,
                    Item, OperationType)
from tracing import setup_tracer


from prometheus_client import Counter, Summary, start_http_server
from prometheus_client.core import CollectorRegistry


SERVICE_REGISTRY = CollectorRegistry()



ITEMS_CREATED = Counter('items_created_total', 'Total number of items created', registry=SERVICE_REGISTRY)
ITEMS_UPDATED = Counter('items_updated_total', 'Total number of items updated', registry=SERVICE_REGISTRY)
ITEMS_DELETED = Counter('items_deleted_total', 'Total number of items deleted', registry=SERVICE_REGISTRY)
REQUEST_LATENCY = Summary('request_latency_seconds', 'Latency of requests in seconds', registry=SERVICE_REGISTRY)

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("manage-item-crud")

BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
TOPIC = "item-events"
PARTITION_RAW = 0  

app = FastAPI(title="ManageItem CRUD service")
setup_tracer(app)
_kafka_producer: AIOKafkaProducer | None = None
_memory_store: Dict[int, Item] = {}




def json_datetime_serializer(obj):
    if isinstance(obj, datetime):
        return obj.isoformat()
    raise TypeError(f"Object of type {obj.__class__.__name__} is not JSON serializable")

@app.on_event("startup")
async def startup_event():
    
    metrics_thread = threading.Thread(target=start_http_server, args=(10668, '0.0.0.0', SERVICE_REGISTRY))
    metrics_thread.daemon = True
    metrics_thread.start()
    log.info("Prometheus metrics server started on port 10668")

    global _kafka_producer
    _kafka_producer = AIOKafkaProducer(
        bootstrap_servers=BOOTSTRAP_SERVERS,
        value_serializer=lambda v: json.dumps(v, default=json_datetime_serializer).encode("utf-8"),
    )
    await _kafka_producer.start()
    log.info("Kafka producer started")


@app.on_event("shutdown")
async def shutdown_event():
    if _kafka_producer:
        await _kafka_producer.stop()
        log.info("Kafka producer stopped")



async def _publish_event(payload: dict, partition: int = PARTITION_RAW) -> None:
    if not _kafka_producer:
        raise RuntimeError("Kafka producer not started")
    
    await _kafka_producer.send_and_wait(
        TOPIC,
        payload,
        partition=partition,
    )
    log.info("Published to partition %d: %s", partition, payload)



@app.post("/items", response_model=CreateItemResponse) 
async def create_item(req: CreateItemRequest):
    start_time = time.time()
    if req.operation_type != OperationType.CREATE:
        REQUEST_LATENCY.observe(time.time() - start_time)
        raise HTTPException(status.HTTP_400_BAD_REQUEST,
                            detail="operation_type must be 3 (create)")
    item = req.item
    _memory_store[item.id] = item
    await _publish_event({
        "operation_type": req.operation_type,
        "item_id": item.id,
        "item": item.model_dump()
    })
    ITEMS_CREATED.inc()
    REQUEST_LATENCY.observe(time.time() - start_time)
    return CreateItemResponse(status=200, item_id=item.id)


@app.put("/items/{item_id}", response_model=ChangeItemResponse)
async def change_item(item_id: int, req: ChangeItemRequest): 
    start_time = time.time()
    if req.operation_type != OperationType.CHANGE:
        REQUEST_LATENCY.observe(time.time() - start_time)
        raise HTTPException(status.HTTP_400_BAD_REQUEST,
                            detail="operation_type must be 2 (change)")
    if item_id not in _memory_store:
        REQUEST_LATENCY.observe(time.time() - start_time)
        raise HTTPException(status.HTTP_404_NOT_FOUND, detail="item not found")

    _memory_store[item_id] = req.item.model_copy(update={"id": item_id})
    await _publish_event({
        "operation_type": req.operation_type,
        "item_id": item_id,
        "item": req.item.model_dump()
    })
    ITEMS_UPDATED.inc()
    REQUEST_LATENCY.observe(time.time() - start_time)
    return ChangeItemResponse(status=200)


@app.delete("/items/{item_id}", response_model=DeleteItemResponse)
async def delete_item(item_id: int, req: DeleteItemRequest): 
    start_time = time.time()
    if req.operation_type != OperationType.DELETE:
        REQUEST_LATENCY.observe(time.time() - start_time)
        raise HTTPException(status.HTTP_400_BAD_REQUEST,
                            detail="operation_type must be 1 (delete)")
    if item_id not in _memory_store:
        REQUEST_LATENCY.observe(time.time() - start_time)
        raise HTTPException(status.HTTP_404_NOT_FOUND, detail="item not found")

    _memory_store.pop(item_id)
    await _publish_event({
        "operation_type": req.operation_type,
        "item_id": item_id,
        "item": None
    })
    ITEMS_DELETED.inc()
    REQUEST_LATENCY.observe(time.time() - start_time)
    return DeleteItemResponse(status=200)


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
