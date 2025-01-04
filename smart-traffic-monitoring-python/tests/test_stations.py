import pytest
from fastapi.testclient import TestClient
from sqlalchemy.orm import Session
import uuid

@pytest.fixture
def example_station_data():
    unique_code = f"STA-{uuid.uuid4().hex[:6].upper()}"
    return {
        "code": unique_code,
        "name": "Test Station",
        "city": "Test City",
        "latitude": 10.123,
        "longitude": 20.456,
        "date_of_installation": "2023-01-01"
    }

def test_create_station(client: TestClient, db_session: Session, example_station_data):
    # 1) Create a station
    response = client.post("/stations/", json=example_station_data)
    assert response.status_code == 200 # or 201, depending on how creation is handled
    data = response.json()
    assert "id" in data
    assert data["code"] == example_station_data["code"]
    station_id = data["id"]

    # 2) Retrieve the station
    get_response = client.get(f"/stations/{station_id}")
    assert get_response.status_code == 200
    get_data = get_response.json()
    assert get_data["name"] == example_station_data["name"]


def test_get_all_stations(client: TestClient, db_session: Session):

    station_1 = {
            "code": "STA-ABC",
            "name": "Station ABC",
            "city": "CityA",
            "latitude": 30.0,
            "longitude": 40.0,
            "date_of_installation": "2023-02-01"
    }

    station_2 = {
        "code": "STA-XYZ",
        "name": "Station XYZ",
        "city": "CityB",
        "latitude": 50.0,
        "longitude": 60.0,
        "date_of_installation": "2023-03-01"
    }
    client.post("/stations/", json=station_1)
    client.post("/stations/", json=station_2)


    # Retrieve all stations
    response = client.get("/stations/")
    assert response.status_code == 200
    stations = response.json()
    assert len(stations) >= 2 # expect atleast the 2 we just created



def test_update_station(client: TestClient, db_session: Session, example_station_data):
    response = client.post("/stations/", json=example_station_data)
    station_id = response.json()["id"]

    update_data = {"name": "Updated Station Name"}
    update_resp = client.put(f"/stations/{station_id}", json=update_data)

    if update_resp.status_code != 200:
        print(update_resp.json())

    assert update_resp.status_code == 200
    updated_station = update_resp.json()
    assert updated_station["name"] == "Updated Station Name"


def test_delete_station(client: TestClient, db_session: Session, example_station_data):
    create_resp = client.post("/stations/", json=example_station_data)
    station_id = create_resp.json()["id"]

    del_resp = client.delete(f"/stations/{station_id}")
    assert del_resp.status_code == 200
    assert del_resp.json()["message"] == "Station deleted successfully"

    # try to get the station to confirm deletion
    get_resp = client.get(f"/stations/{station_id}")
    assert get_resp.status_code == 404










