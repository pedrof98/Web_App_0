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

def test_create_station(client: TestClient, 
                        db_session: Session, 
                        example_station_data, 
                        admin_token_header, 
                        create_station_fixture):
    # 1) Create a station
    station_id = create_station_fixture(example_station_data)
    assert station_id is not None

    # 2) Retrieve the station
    response = client.get(f"/stations/{station_id}", headers=admin_token_header)
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == example_station_data["name"]

def test_get_all_stations(client: TestClient, 
                            db_session: Session, 
                            admin_token_header, 
                            create_station_fixture):
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
    create_station_fixture(station_1)
    create_station_fixture(station_2)

    # Retrieve all stations
    response = client.get("/stations/", headers=admin_token_header)
    assert response.status_code == 200
    stations = response.json()
    assert len(stations) >= 2  # expect at least the 2 we just created

def test_update_station(client: TestClient, 
                        db_session: Session, 
                        example_station_data, 
                        admin_token_header, 
                        create_station_fixture):

    station_id = create_station_fixture(example_station_data)

    update_data = {"name": "Updated Station Name"}
    update_resp = client.put(f"/stations/{station_id}", json=update_data, headers=admin_token_header)

    if update_resp.status_code != 200:
        print(update_resp.json())

    assert update_resp.status_code == 200
    updated_station = update_resp.json()
    assert updated_station["name"] == "Updated Station Name"

def test_delete_station(client: TestClient, 
                        db_session: Session, 
                        example_station_data, 
                        admin_token_header, 
                        create_station_fixture):
    station_id = create_station_fixture(example_station_data)

    del_resp = client.delete(f"/stations/{station_id}", headers=admin_token_header)
    assert del_resp.status_code == 200
    assert del_resp.json()["message"] == "Station deleted successfully"

    # try to get the station to confirm deletion
    get_resp = client.get(f"/stations/{station_id}", headers=admin_token_header)
    assert get_resp.status_code == 404









