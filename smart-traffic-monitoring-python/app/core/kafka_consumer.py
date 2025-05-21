import asyncio
from aiokafka import AIOKafkaConsumer
import json

async def process_message(message):
    #add processing logic
    print(f"Processing message: {message}")

async def start_consumer():
    consumer = AIOKafkaConsumer(
            'traffic-measurements',
            bootstrap_servers='kafka:9092',
            value_deserializer=lambda m: json.loads(m.decode('utf-8'))
            )

    await consumer.start()
    try:
        async for msg in consumer:
            await process_message(msg.value)
    finally:
        await consumer.stop()


if __name__ == "__main__":
    asyncio.run(start_consumer())
