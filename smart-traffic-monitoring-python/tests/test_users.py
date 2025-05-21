
from fastapi.testclient import TestClient
from app.models.users import User, UserRole



def test_login_success(client: TestClient, test_user):
    response = client.post(
            "/token",
            data={"username": "test@example.com", "password": "testpassword"}
            )
    assert response.status_code == 200
    assert "access_token" in response.json()


def test_login_wrong_password(client: TestClient, test_user):
    response = client.post(
            "/token",
            data={"username": "test@example.com", "password": "wrongpassword"}
            )
    assert response.status_code == 400


def test_protected_route(client: TestClient, token_header):
    response = client.get("/stations/", headers=token_header)
    assert response.status_code == 200


def test_protected_route_no_token(client: TestClient):
    response = client.get("/stations/")
    assert response.status_code == 401


def test_admin_route(client: TestClient, admin_token_header):
    response = client.post("/stations/", headers=admin_token_header, json={
        "code": "TEST-1",
        "name": "Test Station",
        "city": "Test City",
        "latitude": 10.0,
        "longitude": 20.0,
        "date_of_installation": "2023-01-01"
        })
    assert response.status_code == 200


def test_admin_route_with_user_token(client: TestClient, token_header):
    response = client.post("/stations/", headers=token_header, json={
        "code": "TEST-1",
        "name": "Test Station",
        "city": "Test City",
        "latitude": 10.0,
        "longitude": 20.0,
        "date_of_installation": "2023-01-01"
        })
    assert response.status_code == 403


def test_create_user(client: TestClient, admin_token_header):
    response = client.post(
        "/users/",
        headers=admin_token_header,
        json={
            "email": "testuser@example.com",
            "password": "newpassword",
            "role": "user"
        }
    )
    if response.status_code != 201:
        print(f"Response JSON: {response.json()}")
    assert response.status_code == 201, f"Expected 200, got {response.status_code}. Response: {response.text}"