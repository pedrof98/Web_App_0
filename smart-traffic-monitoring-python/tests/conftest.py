import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine, event
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from app.main import app
from app.database import get_db
from app.models.base import Base
from app.models.stations import Station
from app.models.userevents import UserEvent
from app.models.sensors import Sensor
from app.models.measurements import TrafficMeasurement

"""
1) Create an in-memory SQLite engine for testing, or a separate Postgres DB

    But for quick demonstration, let's do an in-memory SQLite:
"""

TEST_DATABASE_URL = "sqlite:///:memory:"

engine = create_engine(TEST_DATABASE_URL, connect_args={"check_same_thread": False}, poolclass=StaticPool)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# 2) Create the DB schema once before all tests
@pytest.fixture(scope="session", autouse=True)
def create_test_db():
    Base.metadata.create_all(bind=engine)
    yield
    Base.metadata.drop_all(bind=engine)


# 3) Override the 'get_db' dependency to use our test session
@pytest.fixture
def db_session():
    # Create a new database session for a test
    connection = engine.connect()
    transaction = connection.begin()

    # bind individual session to the connection
    session = TestingSessionLocal(bind=connection)

    # begin nested transaction
    nested = connection.begin_nested()

    # Listen for the 'rollback' event to start a new savepoint after a rollback
    @event.listens_for(session, "after_transaction_end")
    def restart_savepoint(session, transaction):
        if transaction.nested and not transaction._parent.nested:
            connection.begin_nested()
    
    yield session
    
    session.close()
    transaction.rollback()
    connection.close()


# 4) Provide a TestClient that uses the override
@pytest.fixture
def client(db_session):
    def override_get_db():
        try:
            yield db_session
        finally:
            pass


    app.dependency_overrides[get_db] = override_get_db
    with TestClient(app) as c:
        yield c
    # Reset overrides
    app.dependency_overrides.clear()






