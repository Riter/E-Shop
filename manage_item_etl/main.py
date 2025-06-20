import asyncio
import json
import logging
import os
import threading
import time

from aiokafka import AIOKafkaConsumer, AIOKafkaProducer
from aiokafka.structs import TopicPartition


from prometheus_client import Counter, Summary, start_http_server
from prometheus_client.core import CollectorRegistry

from models import Item, OperationType

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("manage-item-etl")

BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
TOPIC = "item-events"
PARTITION_RAW = 0      
PARTITION_PROCESSED = 1 



ITEMS_PROCESSED_CREATE = Counter('items_processed_create_total', 'Total number of CREATE operations processed')
ITEMS_PROCESSED_CHANGE = Counter('items_processed_change_total', 'Total number of CHANGE operations processed')
ITEMS_PROCESSED_DELETE = Counter('items_processed_delete_total', 'Total number of DELETE operations processed')

PROCESSING_TIME = Summary('item_processing_seconds', 'Time spent processing an item')


SERVICE_REGISTRY = CollectorRegistry()


SERVICE_REGISTRY.register(ITEMS_PROCESSED_CREATE)
SERVICE_REGISTRY.register(ITEMS_PROCESSED_CHANGE)
SERVICE_REGISTRY.register(ITEMS_PROCESSED_DELETE)
SERVICE_REGISTRY.register(PROCESSING_TIME)


async def process_and_forward(msg: dict, producer: AIOKafkaProducer):
    """
    Process item events through the ETL pipeline:
    - Extract: Data already extracted from Kafka
    - Transform: Validate, enrich, and format data
    - Load: Forward to processed partition and could write to databases
    """
    start_time = time.time()
    
    try:
        
        op = msg.get("operation_type")
        if op not in [OperationType.DELETE, OperationType.CHANGE, OperationType.CREATE]:
            log.error(f"Invalid operation type: {op}")
            return
            
        op_name = {
            OperationType.DELETE: "DELETE", 
            OperationType.CHANGE: "CHANGE", 
            OperationType.CREATE: "CREATE"
        }.get(op)
        
        log.info("ETL processing %s operation for item %s", op_name, msg.get("item_id"))

        
        if op in [OperationType.CREATE, OperationType.CHANGE]:
            
            if "item" not in msg:
                log.error(f"Missing item data for {op_name} operation")
                return
                
            
            item_data = msg["item"]
            
            
            if "category" in item_data:
                item_data["category"] = item_data["category"].lower()
                
            
            if "price" in item_data and item_data["price"] < 0:
                item_data["price"] = 0.0

            
            log.info(f"Transformed item {item_data.get('id')}: category normalized, price validated")
                
        elif op == OperationType.DELETE:
            
            if "item_id" not in msg:
                log.error("Missing item_id for DELETE operation")
                return
        
        
        await producer.send_and_wait(
            TOPIC,
            msg,
            partition=PARTITION_PROCESSED
        )
        log.info("ETL: forwarded item %s to partition %d", 
                 msg.get("item_id") or msg.get("item", {}).get("id"), 
                 PARTITION_PROCESSED)
        
        
        if op == OperationType.CREATE:
            ITEMS_PROCESSED_CREATE.inc()
        elif op == OperationType.CHANGE:
            ITEMS_PROCESSED_CHANGE.inc()
        elif op == OperationType.DELETE:
            ITEMS_PROCESSED_DELETE.inc()
        
    except Exception as e:
        log.error(f"Error processing message: {str(e)}")
    finally:
        
        PROCESSING_TIME.observe(time.time() - start_time)


async def main():
    
    metrics_thread = threading.Thread(
        target=start_http_server, 
        args=(10667, '0.0.0.0', SERVICE_REGISTRY)
    )
    metrics_thread.daemon = True
    metrics_thread.start()
    log.info("Prometheus metrics server started on port 10667")
    
    
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
