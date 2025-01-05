from aiokafka import AIOKafkaProducer
import json
from typing import Any
import asyncio

class KafkaClient:
    def __init__(self, loop=None):
        self.producer = None
        self.loop = loop

    async def initialize(self):
        self.producer = AIOKafkaProducer(
                loop=self.loop,
                bootstrap_servers='kafka:9092',
                value_serializer=lambda v: json.dumps(v).encode('utf-8')
                )
        await self.producer.start()
    
    async def close(self):
        await self.producer.close()


    async def send_message(self, topic: str, value: Any, key: str = None):
        try:
            await self.producer.send_and_await(topic, value, key=key.encode() if key else None)
        except Exception as e:
            print(f"Error sending message to Kafka: {e}")
            raise
