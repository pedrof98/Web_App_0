import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine, event
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool
from unittest.mock import AsyncMock, patch
from typing import Any

from app.database import get_db
from app.models.base import Base
from app.models.stations import Station
from app.models.userevents import UserEvent
from app.models.sensors import Sensor
from app.models.measurements import TrafficMeasurement

TEST_DATABASE_URL = "sqlite:///:memory:"

engine = create_engine(
    TEST_DATABASE_URL,
    connect_args={"check_same_thread": False},
    poolclass=StaticPool
)
TestingSessionLocal = sessionmaker(
    autocommit=False,
    autoflush=False,
    bind=engine
)

class MockKafkaClient:
    def __init__(self, loop=None):
        self.producer = AsyncMock()
        self.loop = loop

    async def initialize(self):
        return

    async def close(self):
        return

    async def send_message(self, topic: str, value: Any, key: str = None):
        return 

@pytest.fixture(scope="session", autouse=True)
def create_test_db():
    Base.metadata.create_all(bind=engine)
    yield
    Base.metadata.drop_all(bind=engine)

@pytest.fixture
def db_session():
    connection = engine.connect()
    transaction = connection.begin()

    session = TestingSessionLocal(bind=connection)

    @event.listens_for(session, "after_transaction_end")
    def restart_savepoint(session, transaction):
        if transaction.nested and not transaction._parent.nested:
            connection.begin_nested()

    yield session

    session.close()
    transaction.rollback()
    connection.close()

@pytest.fixture
def mock_kafka():
    with patch("app.core.kafka_config.KafkaClient", return_value=MockKafkaClient()) as mock:
        yield mock

@pytest.fixture
def app_with_mock_kafka(mock_kafka):
    from app.main import app  
    return app

@pytest.fixture
def client(app_with_mock_kafka, db_session):
    def override_get_db():
        try:
            yield db_session
        finally:
            pass

    app_with_mock_kafka.dependency_overrides[get_db] = override_get_db
    with TestClient(app_with_mock_kafka) as c:
        yield c
    app_with_mock_kafka.dependency_overrides.clear()
