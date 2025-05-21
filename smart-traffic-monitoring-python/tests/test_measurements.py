import pytest
from fastapi.testclient import TestClient
from sqlalchemy.orm import Session
import uuid
from datetime import datetime, timedelta



@pytest.fixture
def example_measurement_data():
    """Generates example data for a single measurement"""
    return {
            "sensor_id": None,
            "timestamp": (datetime.utcnow()).isoformat(),
            "speed": 50.5,
            "vehicle_count": 10
            }


@pytest.fixture
def example_batch_measurement_data():
    """Generates example data for batch traffic measurements"""
    current_time = datetime.utcnow()
    measurements = [
            {
                "sensor_id": None,
                "timestamp": (current_time - timedelta(minutes=i)).isoformat(),
                "speed": 45.0 + i,
                "vehicle_count": 5 + i
                } for i in range(5)
            ]
    return {"measurements": measurements}


def test_ingest_real_time_Data(client: TestClient, 
                                admin_token_header, 
                                db_session: Session, 
                                example_measurement_data,
                                create_station_fixture,
                                create_sensor_fixture):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Real-Time Station",
        "city": "Realtime City",
        "latitude": 10.0,
        "longitude": 20.0,
        "date_of_installation": "2023-05-01"
    }
    station_id = create_station_fixture(station_data)

    sensor_data = {
            "sensor_id": f"SEN-{uuid.uuid4().hex[:6].upper()}",
            "measurement_type": "Temperature",
            "status": "active",
            "station_id": station_id
            }
    sensor_id = create_sensor_fixture(sensor_data)

    # update measurement data with sensor id
    example_measurement_data["sensor_id"] = sensor_id

    # ingest real-time measurement data
    response = client.post("/data/real-time", json=example_measurement_data, headers=admin_token_header)
    assert response.status_code == 200, f"Failed to ingest real-time data: {response.text}"
    data = response.json()
    assert "id" in data
    assert data["sensor_id"] == sensor_id
    assert data["speed"] == example_measurement_data["speed"]
    assert data["vehicle_count"] == example_measurement_data["vehicle_count"]
    assert "created_at" in data

    # Retrieve the ingested measurement
    measurement_id = data["id"]
    get_response = client.get(f"/data/{measurement_id}")
    assert get_response.status_code == 200, f"Failed to retrieve measurement: {get_response.text}"
    get_data = get_response.json()
    assert get_data["id"] == measurement_id
    assert get_data["sensor_id"] == sensor_id
    assert get_data["speed"] == example_measurement_data["speed"]
    assert get_data["vehicle_count"] == example_measurement_data["vehicle_count"]


def test_ingest_batch_data(client: TestClient, 
                            db_session: Session, 
                            example_batch_measurement_data, 
                            admin_token_header,
                            create_station_fixture,
                            create_sensor_fixture):
     # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Batch Station",
        "city": "Batch City",
        "latitude": 30.0,
        "longitude": 40.0,
        "date_of_installation": "2023-06-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Create a sensor linked to the station
    sensor_data = {
        "sensor_id": f"SEN-{uuid.uuid4().hex[:6].upper()}",
        "measurement_type": "Speed",
        "status": "active",
        "station_id": station_id
    }
    sensor_id = create_sensor_fixture(sensor_data)

    # Update all measurements with the sensor_id
    for measurement in example_batch_measurement_data["measurements"]:
        measurement["sensor_id"] = sensor_id

    # Ingest batch measurement data
    response = client.post("/data/batch", json=example_batch_measurement_data, headers=admin_token_header)
    assert response.status_code == 200, f"Failed to ingest batch data: {response.text}"
    data = response.json()
    assert "message" in data
    assert "Ingested 5 measurements successfully" in data["message"]

     # 5. Retrieve all measurements and verify
    get_response = client.get("/data/", headers=admin_token_header)
    assert get_response.status_code == 200, f"Failed to retrieve all measurements: {get_response.text}"
    measurements = get_response.json()
    # At least the 5 measurements we just created should be present
    assert len(measurements) >= 5
    # Verify that the ingested measurements exist
    ingested_timestamps = {m["timestamp"] for m in example_batch_measurement_data["measurements"]}
    retrieved_timestamps = {m["timestamp"] for m in measurements}
    assert ingested_timestamps.issubset(retrieved_timestamps), "Ingested batch measurements not found in retrieved measurements"


def test_get_all_measurements(client: TestClient, 
                                db_session: Session, 
                                example_measurement_data, 
                                admin_token_header,
                                create_station_fixture,
                                create_sensor_fixture):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Get All Measurements Station",
        "city": "GetAll City",
        "latitude": 50.0,
        "longitude": 60.0,
        "date_of_installation": "2023-07-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Create a sensor linked to the station
    sensor_data = {
        "sensor_id": f"SEN-{uuid.uuid4().hex[:6].upper()}",
        "measurement_type": "Vehicle Count",
        "status": "active",
        "station_id": station_id
    }
    sensor_id = create_sensor_fixture(sensor_data)

    # 3. Create multiple measurements
    measurement_1 = example_measurement_data.copy()
    measurement_1["sensor_id"] = sensor_id
    measurement_1["timestamp"] = (datetime.utcnow() - timedelta(minutes=10)).isoformat()
    measurement_1["speed"] = 60.0
    measurement_1["vehicle_count"] = 15

    measurement_2 = example_measurement_data.copy()
    measurement_2["sensor_id"] = sensor_id
    measurement_2["timestamp"] = (datetime.utcnow() - timedelta(minutes=5)).isoformat()
    measurement_2["speed"] = 55.5
    measurement_2["vehicle_count"] = 12

    # Ingest measurements
    response1 = client.post("/data/real-time", json=measurement_1, headers=admin_token_header)
    assert response1.status_code == 200, f"Failed to ingest measurement 1: {response1.text}"
    response2 = client.post("/data/real-time", json=measurement_2, headers=admin_token_header)
    assert response2.status_code == 200, f"Failed to ingest measurement 2: {response2.text}"

    # 4. Retrieve all measurements
    get_response = client.get("/data/", headers=admin_token_header)
    assert get_response.status_code == 200, f"Failed to retrieve all measurements: {get_response.text}"
    measurements = get_response.json()
    # At least the two measurements we just created should be present
    assert len(measurements) >= 2
    # Verify that the ingested measurements exist
    created_measurement_ids = {response1.json()["id"], response2.json()["id"]}
    retrieved_measurement_ids = {m["id"] for m in measurements}
    assert created_measurement_ids.issubset(retrieved_measurement_ids), "Ingested measurements not found in retrieved measurements"

def test_get_measurement_by_id(client: TestClient, 
                                db_session: Session, 
                                example_measurement_data, 
                                admin_token_header,
                                create_station_fixture,
                                create_sensor_fixture):
    # 1. Create a station
    station_data = {
        "code": f"STA-{uuid.uuid4().hex[:6].upper()}",
        "name": "Get Measurement Station",
        "city": "GetMeasurement City",
        "latitude": 70.0,
        "longitude": 80.0,
        "date_of_installation": "2023-08-01"
    }
    station_id = create_station_fixture(station_data)

    # 2. Create a sensor linked to the station
    sensor_data = {
        "sensor_id": f"SEN-{uuid.uuid4().hex[:6].upper()}",
        "measurement_type": "Speed",
        "status": "active",
        "station_id": station_id
    }
    sensor_id = create_sensor_fixture(sensor_data)

    # 3. Create a measurement
    example_measurement_data["sensor_id"] = sensor_id
    example_measurement_data["timestamp"] = (datetime.utcnow()).isoformat()
    response = client.post("/data/real-time", json=example_measurement_data, headers=admin_token_header)
    assert response.status_code == 200, f"Failed to ingest measurement: {response.text}"
    measurement = response.json()
    measurement_id = measurement["id"]

    # 4. Retrieve the measurement by ID
    get_response = client.get(f"/data/{measurement_id}", headers=admin_token_header)
    assert get_response.status_code == 200, f"Failed to retrieve measurement by ID: {get_response.text}"
    get_data = get_response.json()
    assert get_data["id"] == measurement_id
    assert get_data["sensor_id"] == sensor_id
    assert get_data["speed"] == example_measurement_data["speed"]
    assert get_data["vehicle_count"] == example_measurement_data["vehicle_count"]
    assert "created_at" in get_data

    # 5. Attempt to retrieve a non-existent measurement
    non_existent_id = measurement_id + 999
    get_non_existent = client.get(f"/data/{non_existent_id}", headers=admin_token_header)
    assert get_non_existent.status_code == 404, f"Expected 404 for non-existent measurement, got {get_non_existent.status_code}"
    assert get_non_existent.json()["detail"] == "Measurement not found"














