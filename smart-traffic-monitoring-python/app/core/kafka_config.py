from aiokafka import AIOKafkaProducer
import json
from typing import Any
import asyncio
from fastapi import Depends

class KafkaClient:
    def __init__(self, loop=None, bootstrap_servers='kafka:9092'):
        self.producer = None
        self.loop = loop
        self.bootstrap_servers = bootstrap_servers

    async def initialize(self):
        self.producer = AIOKafkaProducer(
                loop=self.loop,
                bootstrap_servers=self.bootstrap_servers,
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


async def get_kafka_client():
    kafka_client = KafkaClient(loop=asyncio.get_event_loop())
    await kafka_client.initialize()
    try:
        yield kafka_client
    finally:
        await kafka_client.close()