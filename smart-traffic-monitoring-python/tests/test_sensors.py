import pytest
from fastapi.testclient import TestClient
from sqlalchemy.orm import Session
import uuid



@pytest.fixture
def example_sensor_data():
    unique_code = f"SEN-{uuid.uuid4().hex[:6].upper()}"
    return {
            "sensor_id": unique_code,           
            "measurement_type": "Temperature",    
            "status": "active",                  
            "station_id": None                   
            }


def test_create_sensor(client: TestClient, 
                        db_session: Session, 
                        example_sensor_data, 
                        admin_token_header, 
                        create_station_fixture):
    # 1. Create a station first
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Sensor Station",
        "city": "Sensor City",
        "latitude": 25.0,
        "longitude": 35.0,
        "date_of_installation": "2023-01-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Update sensor_data with the station_id
    example_sensor_data["station_id"] = station_id

    # 3. Create a sensor
    response = client.post("/sensors/", json=example_sensor_data, headers=admin_token_header)
    assert response.status_code == 200, f"Failed to create sensor: {response.text}"
    data = response.json()
    assert "id" in data
    assert data["sensor_id"] == example_sensor_data["sensor_id"]
    assert data["station_id"] == station_id
    sensor_id = data["id"]

    # 4. Retrieve the sensor
    get_response = client.get(f"/sensors/{sensor_id}", headers=admin_token_header)
    assert get_response.status_code == 200
    get_data = get_response.json()
    assert get_data["status"] == example_sensor_data["status"]
    assert get_data["measurement_type"] == example_sensor_data["measurement_type"]
    assert get_data["station_id"] == station_id

def test_get_all_sensors(client: TestClient, 
                            db_session: Session, 
                            example_sensor_data, 
                            admin_token_header, 
                            create_station_fixture):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Multiple Sensors Station",
        "city": "Multi City",
        "latitude": 45.0,
        "longitude": 55.0,
        "date_of_installation": "2023-02-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Create multiple sensors
    sensor_1 = example_sensor_data.copy()
    sensor_1["sensor_id"] = f"SEN-{uuid.uuid4().hex[:6].upper()}"
    sensor_1["station_id"] = station_id

    sensor_2 = example_sensor_data.copy()
    sensor_2["sensor_id"] = f"SEN-{uuid.uuid4().hex[:6].upper()}"
    sensor_2["station_id"] = station_id

    response1 = client.post("/sensors/", json=sensor_1, headers=admin_token_header)
    assert response1.status_code == 200, f"Failed to create sensor 1: {response1.text}"
    response2 = client.post("/sensors/", json=sensor_2, headers=admin_token_header)
    assert response2.status_code == 200, f"Failed to create sensor 2: {response2.text}"

    # 3. Retrieve all sensors
    response = client.get("/sensors/", headers=admin_token_header)
    assert response.status_code == 200
    sensors = response.json()
    # At least the two sensors we just created should be present
    assert len(sensors) >= 2
    created_sensor_codes = {sensor_1["sensor_id"], sensor_2["sensor_id"]}
    retrieved_sensor_codes = {sensor["sensor_id"] for sensor in sensors}
    assert created_sensor_codes.issubset(retrieved_sensor_codes), "Created sensors not found in retrieved sensors"

def test_update_sensor(client: TestClient, 
                        db_session: Session, 
                        example_sensor_data, 
                        admin_token_header, 
                        create_station_fixture):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Update Sensor Station",
        "city": "Update City",
        "latitude": 60.0,
        "longitude": 70.0,
        "date_of_installation": "2023-03-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Create a sensor
    example_sensor_data["station_id"] = station_id
    response = client.post("/sensors/", json=example_sensor_data, headers=admin_token_header)
    assert response.status_code == 200, f"Failed to create sensor: {response.text}"
    sensor = response.json()
    sensor_id = sensor["id"]

    # 3. Update the sensor
    update_data = {
        "status": "Inactive",
        "measurement_type": "Humidity"
    }
    update_resp = client.put(f"/sensors/{sensor_id}", json=update_data, headers=admin_token_header)
    if update_resp.status_code != 200:
        print(update_resp.json())
    assert update_resp.status_code == 200, f"Failed to update sensor: {update_resp.text}"
    updated_sensor = update_resp.json()
    assert updated_sensor["status"] == update_data["status"]
    assert updated_sensor["measurement_type"] == update_data["measurement_type"]

def test_delete_sensor(client: TestClient, 
                        db_session: Session, 
                        example_sensor_data, 
                        admin_token_header, 
                        create_station_fixture):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Delete Sensor Station",
        "city": "Delete City",
        "latitude": 80.0,
        "longitude": 90.0,
        "date_of_installation": "2023-04-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Create a sensor
    example_sensor_data["station_id"] = station_id
    create_resp = client.post("/sensors/", json=example_sensor_data, headers=admin_token_header)
    assert create_resp.status_code == 200, f"Failed to create sensor: {create_resp.text}"
    sensor_id = create_resp.json()["id"]

    # 3. Delete the sensor
    del_resp = client.delete(f"/sensors/{sensor_id}", headers=admin_token_header)
    assert del_resp.status_code == 200, f"Failed to delete sensor: {del_resp.text}"
    assert del_resp.json()["message"] == "Sensor deleted successfully"

    # 4. Try to get the deleted sensor
    get_resp = client.get(f"/sensors/{sensor_id}", headers=admin_token_header)
    assert get_resp.status_code == 404, f"Deleted sensor still exists: {get_resp.json()}"





