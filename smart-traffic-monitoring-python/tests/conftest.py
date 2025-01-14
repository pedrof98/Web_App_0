import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine, event
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool
from unittest.mock import AsyncMock, patch
from typing import Any

from app.core.security import get_password_hash
from app.database import get_db
from app.models.base import Base
from app.models.stations import Station
from app.models.userevents import UserEvent
from app.models.sensors import Sensor
from app.models.measurements import TrafficMeasurement
from app.models.users import User, UserRole
from app.core.auth import create_access_token

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


@pytest.fixture
def test_user(db_session):
    user = User(
            email="test@example.com",
            hashed_password=get_password_hash("testpassword"),
            role=UserRole.USER
            )
    db_session.add(user)
    db_session.commit()
    db_session.refresh(user)
    return user


@pytest.fixture
def test_admin(db_session):
    admin = User(
            email="admin@example.com",
            hashed_password=get_password_hash("adminpassword"),
            role=UserRole.ADMIN
            )
    db_session.add(admin)
    db_session.commit()
    db_session.refresh(admin)
    return admin


@pytest.fixture
def token_header(test_user):
    access_token = create_access_token(data={"sub": str(test_user.id), "role": test_user.role})
    return {"Authorization": f"Bearer {access_token}"}


@pytest.fixture
def admin_token_header(test_admin):
    access_token = create_access_token(data={"sub": str(test_admin.id), "role": test_admin.role})
    return {"Authorization": f"Bearer {access_token}"}



@pytest.fixture
def test_station(client: TestClient, admin_token_header):
    response = client.post(
            "/stations/",
            headers=admin_token_header,
            json={
                "code": "TEST-1",
                "name": "Test Station",
                "city": "Test City",
                "latitude": 10.0,
                "longitude": 20.0,
                "date_of_installation": "2023-01-01"
                }
            )
    return response.json()

@pytest.fixture
def test_sensor(client: TestClient, admin_token_header, test_station):
    response = client.post(
            "/sensors/",
            headers=admin_token_header,
            json={
                "sensor_id": "SEN-TEST1",
                "measurement_type": "traffic_counter",
                "status": "active",
                "station_id": test_station["id"]
                }
            )
    return response.json()


@pytest.fixture
def create_station_fixture(client: TestClient, admin_token_header):
    def _create_station(station_data: dict, headers: dict = admin_token_header) -> int:
        response = client.post("/stations/", headers=headers, json=station_data)
        assert response.status_code == 200, f"Failed to create station: {response.text}"
        station = response.json()
        return station["id"]
    return _create_station


@pytest.fixture
def create_sensor_fixture(client: TestClient, admin_token_header):
    def _create_sensor(sensor_data: dict, headers: dict = admin_token_header) -> int:
        response = client.post("/sensors/", headers=headers, json=sensor_data)
        assert response.status_code == 200, f"Failed to create sensor: {response.text}"
        sensor = response.json()
        return sensor["id"]
    return _create_sensor