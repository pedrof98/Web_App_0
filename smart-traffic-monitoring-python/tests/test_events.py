
import pytest
from fastapi.testclient import TestClient
from sqlalchemy.orm import Session
import uuid
from datetime import date, datetime, timedelta

# Assuming the 'client' fixture is defined in conftest.py

@pytest.fixture
def example_event_data():
    """Generates example data for a single user event."""
    return {
        "date": date.today().isoformat(),
        "city": "Test City",
        "event_type": "Accident",
        "description": "Minor accident on highway.",
        "expected_congestion_level": "High",
        "station_id": None  # To be set in the test after creating a station
    }

@pytest.fixture
def example_batch_event_data():
    """Generates example data for batch user events."""
    current_date = date.today()
    events = [
        {
            "date": (current_date - timedelta(days=i)).isoformat(),
            "city": f"City {i}",
            "event_type": "Roadwork" if i % 2 == 0 else "Concert",
            "description": f"Description for event {i}",
            "expected_congestion_level": "Medium" if i % 2 == 0 else "Low",
            "station_id": None  # To be set in the test after creating a station
        } for i in range(5)
    ]
    return {"events": events}

def create_station(client: TestClient, station_data: dict) -> int:
    """
    Helper function to create a station and return its ID.
    """
    response = client.post("/stations/", json=station_data)
    assert response.status_code == 200, f"Failed to create station: {response.text}"
    station = response.json()
    return station["id"]

def test_create_user_event(client: TestClient, db_session: Session, example_event_data):
    # 1. Create a station first
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Event Station",
        "city": "Event City",
        "latitude": 15.0,
        "longitude": 25.0,
        "date_of_installation": "2023-01-01"
    }
    station_id = create_station(client, station_data)

    # 2. Update event_data with the station_id
    example_event_data["station_id"] = station_id

    # 3. Create a user event
    response = client.post("/events/", json=example_event_data)
    assert response.status_code == 200, f"Failed to create user event: {response.text}"
    data = response.json()
    assert "id" in data
    assert data["city"] == example_event_data["city"]
    assert data["event_type"] == example_event_data["event_type"]
    assert data["station_id"] == station_id
    event_id = data["id"]

    # 4. Retrieve the user event
    get_response = client.get(f"/events/{event_id}")
    assert get_response.status_code == 200, f"Failed to retrieve user event: {get_response.text}"
    get_data = get_response.json()
    assert get_data["description"] == example_event_data["description"]
    assert get_data["expected_congestion_level"] == example_event_data["expected_congestion_level"]
    assert get_data["station_id"] == station_id

def test_get_all_user_events(client: TestClient, db_session: Session, example_event_data):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Multiple Events Station",
        "city": "Multiple City",
        "latitude": 35.0,
        "longitude": 45.0,
        "date_of_installation": "2023-02-01"
    }
    station_id = create_station(client, station_data)

    # 2. Create multiple user events
    event_1 = example_event_data.copy()
    event_1["city"] = "City A"
    event_1["event_type"] = "Concert"
    event_1["station_id"] = station_id

    event_2 = example_event_data.copy()
    event_2["city"] = "City B"
    event_2["event_type"] = "Parade"
    event_2["station_id"] = station_id

    response1 = client.post("/events/", json=event_1)
    assert response1.status_code == 200, f"Failed to create user event 1: {response1.text}"
    response2 = client.post("/events/", json=event_2)
    assert response2.status_code == 200, f"Failed to create user event 2: {response2.text}"

    # 3. Retrieve all user events
    response = client.get("/events/")
    assert response.status_code == 200, f"Failed to retrieve all user events: {response.text}"
    events = response.json()
    # At least the two events we just created should be present
    assert len(events) >= 2, "Less than 2 user events retrieved."
    created_event_ids = {response1.json()["id"], response2.json()["id"]}
    retrieved_event_ids = {event["id"] for event in events}
    assert created_event_ids.issubset(retrieved_event_ids), "Created user events not found in retrieved events."

def test_update_user_event(client: TestClient, db_session: Session, example_event_data):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Update Event Station",
        "city": "Update City",
        "latitude": 55.0,
        "longitude": 65.0,
        "date_of_installation": "2023-03-01"
    }
    station_id = create_station(client, station_data)

    # 2. Create a user event
    example_event_data["station_id"] = station_id
    response = client.post("/events/", json=example_event_data)
    assert response.status_code == 200, f"Failed to create user event: {response.text}"
    event = response.json()
    event_id = event["id"]

    # 3. Update the user event
    update_data = {
        "city": "Updated City",
        "event_type": "Road Closure",
        "description": "Road closure due to maintenance.",
        "expected_congestion_level": "Low"
    }
    update_resp = client.put(f"/events/{event_id}", json=update_data)
    if update_resp.status_code != 200:
        print(update_resp.json())
    assert update_resp.status_code == 200, f"Failed to update user event: {update_resp.text}"
    updated_event = update_resp.json()
    assert updated_event["city"] == update_data["city"]
    assert updated_event["event_type"] == update_data["event_type"]
    assert updated_event["description"] == update_data["description"]
    assert updated_event["expected_congestion_level"] == update_data["expected_congestion_level"]

def test_delete_user_event(client: TestClient, db_session: Session, example_event_data):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Delete Event Station",
        "city": "Delete City",
        "latitude": 75.0,
        "longitude": 85.0,
        "date_of_installation": "2023-04-01"
    }
    station_id = create_station(client, station_data)

    # 2. Create a user event
    example_event_data["station_id"] = station_id
    create_resp = client.post("/events/", json=example_event_data)
    assert create_resp.status_code == 200, f"Failed to create user event: {create_resp.text}"
    event_id = create_resp.json()["id"]

    # 3. Delete the user event
    del_resp = client.delete(f"/events/{event_id}")
    assert del_resp.status_code == 200, f"Failed to delete user event: {del_resp.text}"
    assert del_resp.json()["message"] == "User event deleted successfully"

    # 4. Try to get the deleted user event
    get_resp = client.get(f"/events/{event_id}")
    assert get_resp.status_code == 404, f"Deleted user event still exists: {get_resp.json()}"
    assert get_resp.json()["detail"] == "Event not found."
