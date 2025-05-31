import asyncio
import json
import logging
import os
import threading
import time

from aiokafka import AIOKafkaConsumer, AIOKafkaProducer
from aiokafka.structs import TopicPartition

# Import Prometheus client libraries
from prometheus_client import Counter, Summary, start_http_server
from prometheus_client.core import CollectorRegistry

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("manage-item-etl")

BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
TOPIC = "item-events"
PARTITION_RAW = 0      # оттуда читаем «сырые» события
PARTITION_PROCESSED = 1 # туда пишем «post-ETL» события

# Prometheus metrics
# Track messages processed by operation type
ITEMS_PROCESSED_CREATE = Counter('items_processed_create_total', 'Total number of CREATE operations processed')
ITEMS_PROCESSED_CHANGE = Counter('items_processed_change_total', 'Total number of CHANGE operations processed')
ITEMS_PROCESSED_DELETE = Counter('items_processed_delete_total', 'Total number of DELETE operations processed')
# Track processing time
PROCESSING_TIME = Summary('item_processing_seconds', 'Time spent processing an item')

# Create a registry for this service
SERVICE_REGISTRY = CollectorRegistry()

# Register metrics
SERVICE_REGISTRY.register(ITEMS_PROCESSED_CREATE)
SERVICE_REGISTRY.register(ITEMS_PROCESSED_CHANGE)
SERVICE_REGISTRY.register(ITEMS_PROCESSED_DELETE)
SERVICE_REGISTRY.register(PROCESSING_TIME)


async def process_and_forward(msg: dict, producer: AIOKafkaProducer):
    """
    Здесь помещаем «Т» (Transform) и «L» (Load) части ETL-конвейера:
    - enrich / очистка
    - запись в Postgres / ElasticSearch / MinIO / S3 и т. д.
    Сейчас просто печатаем.
    """
    start_time = time.time()
    op = msg["operation_type"]
    op_name = {1: "DELETE", 2: "CHANGE", 3: "CREATE"}.get(op, "?")
    log.info("ETL received %s for item %s", op_name, msg["item_id"])
    # TODO: T/L-логика здесь

    # далее ― отправляем в partition 1 для фасадных сервисов
    await producer.send_and_wait(
        TOPIC,
        msg,
        partition=PARTITION_PROCESSED
    )
    log.info("ETL: forwarded item %s to partition %d", msg["item_id"], PARTITION_PROCESSED)
    
    # Increment appropriate counter based on operation type
    if op == 3:  # CREATE
        ITEMS_PROCESSED_CREATE.inc()
    elif op == 2:  # CHANGE
        ITEMS_PROCESSED_CHANGE.inc()
    elif op == 1:  # DELETE
        ITEMS_PROCESSED_DELETE.inc()
    
    # Record processing time
    PROCESSING_TIME.observe(time.time() - start_time)


async def main():
    # Start Prometheus metrics server
    metrics_thread = threading.Thread(
        target=start_http_server, 
        args=(10667, '', SERVICE_REGISTRY)
    )
    metrics_thread.daemon = True
    metrics_thread.start()
    log.info("Prometheus metrics server started on port 10667")
    
    # 1) настроить consumer, привязать к partition RAW
    consumer = AIOKafkaConsumer(
        bootstrap_servers=BOOTSTRAP_SERVERS,
        value_deserializer=lambda b: json.loads(b.decode("utf-8")),
        auto_offset_reset="earliest",
        enable_auto_commit=True,
        group_id="manage-item-etl",
    )
    await consumer.start()
    consumer.assign([TopicPartition(TOPIC, PARTITION_RAW)])
    log.info("ETL consumer started on %s[%d]", TOPIC, PARTITION_RAW)

    # 2) настроить producer, чтобы форвардить дальше
    producer = AIOKafkaProducer(
        bootstrap_servers=BOOTSTRAP_SERVERS,
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
    )
    await producer.start()
    log.info("ETL producer started")

    try:
        async for record in consumer:
            await process_and_forward(record.value, producer)
    finally:
        await consumer.stop()
        await producer.stop()
        log.info("ETL shutdown complete")


if __name__ == "__main__":
    asyncio.run(main())
