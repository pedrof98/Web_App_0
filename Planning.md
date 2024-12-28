# Domain & Use Case

Smart Traffic Monitoring System

Scenario: A platform that gathers real-time traffic information from sensors installed on highways and city streets.
Additionaly, transportation authorities and end-users can submit planned road closures, construction schedules, and even forecasted traffic patterns.
The purpose of this side project is to build the backend services for this platform, focusing on data ingestion, storage, and retrieval for analytics/visualization in a future UI.


# High-Level Requirements

1. Data ingestion from multiple sources:
    - Real-time sensors (sending data every few seconds/minutes)
    - Daily batch files from city authorities (e.g., records of traffic incidents for the past 24 hours)
    - User input (road closures, congestion forecasts, or planned events that might affect traffic)

2. CRUD operations:
    - Managing "traffic stations"
    - Managing "sensors" (porperties and location)

3. Data retrieval aggregation:
    - Ability to query current or historical traffic conditions for a specific station, date/time, or city zone.
    - Ability to compare "actual" vs. "forecasted" congestion levels (idea taken from a previous project)

4. Scalability and containerization:
    - Application should be Dockerized. Ideally, multiple services (API service, database, optional queue/stream service) run via Docker Compose.

5. Documentation and presentation:
    - Clear API documentation (Swagger/OpenAPI or similar)
    - Documentation explaining why I chose the architecture, frameworks and data stores I did.
    - Brief architectural slide deck that I can use to walk through the design (optional)


# Data Sources and Processing

## Real-time sensor data
    - Example JSON:
```json	
{
    "identifier": "sensor-1001-hwy",
    "station": "STA-001",
    "timestamp": "2025-04-01T08:30:00Z",
    "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
    },
    "metrics": [
        {
        "type": "speed",
        "value": 65,
        "unit": "mph"
        },
        {
        "type": "vehicle_count",
        "value": 32,
        "unit": "cars/minute"
        }
    ]
}
```
    - These events come in every few seconds from each sensor.
    The project should demonstrate how to handle frequent incoming data in a scalable or atleast organized manner.

## Daily Batch Uploads

    - City authorities send large JSON/CSV files containing incidents or aggregated data from the previous day (e.g., average speed, acidents, road closures). For this project, we should mock the daily batch, but demonstrate that the implemented system can ingest it though an API endpoint or background job.

## User-Submitted Data

    - An admin or user can log in to add planned events (like road closures, large public gatherings or weather impacts):

```json
    {
        "date": "2025-04-02",
        "city": "New York",
        "event_type": "Construction",
        "description": "Lane closures on I-95",
        "expected_congestion_level": "high"
    }
```

    - They might also provide forecasted traffic for a future date.


# Features and Functionality

1. Core CRUD
    - Traffic Stations:
        * Code (unique id)
        * Name/City/ Geo-coordinates
        * Date of installation
    
    - Sensors:
        * Sensor ID
        * Station Code they belong to
        * Measurement type (speed, vehicle_count, etc)
        * Status (active/inactive)

2. Data Ingestion Services

    - An API endpoint (or separate microservice) that reeives real-time sensor data asynchronously.
    - An endpoint or background job that processes daily batch data.

3. Reporting and Retrieval

    - Query endpoints that return current or average speed and vehicle counts in a given timeframe.
    - Ability to compare actual vs. forecasted congestion levels for a given date/time range.
    - Summaries: e.g., total incidents reported for the day, top congested highways, etc.

4. User management (optional)

    - Basic authentication or token-based auth (e.g., JWT) to protect admin endpoints (create/edit/delete stations)

5. Errors and Logging

    - The system should handle invalid inputs gracefully, logging errors in a structured way (e.g., JSON logs, or to a logging microservice if you want to get fancy)



# Technical Requirements

## Functional Requirements

1. API Endpoints

    - POST /stations - Create a new traffic station
    - GET /stations/:id - Retrieve station details
    - PUT /stations/:id - Update station properties
    - DELETE /stations/:id - Delete station
    - POST /sensors - Create a new sensor
    - GET /sensors/:id - Retrieve sensor details
    - PUT /sensors/:id - Update sensor properties
    - DELETE /sensors/:id - Delete sensor
    - POST /data/real-time - Ingest real-time sensor data
    - POST /data/batch - Ingest daily batch data (this might be ignored if we want to simplify)
    - GET /traffic - Query aggregated or raw traffic data (filter by city, date range, station, etc...)


2. Data Modeling

    - Tables or collections should be designed (either relational od non relational DBs) in a way that can handle the station/sensor relationships, the incoming measurements, and the user-submitted forecasts/events.

3. Documentation 

    - Must include an API reference. Using either OPENAPI/Swagger is recommended.


## Non-Functional Requirements

1. Performance

    - The system shoud handle frequent writes (real-time data). Consider how to scale if the data volume grows.


2. Scalability

    - Dockerize each service. Potentially creating separate containers for:
        - API
        - Database
        - Queue/Cache

3. Resilience

    - Basic Fault tolerance: the system should handle invalid data gracefully, return appropriate HTTP errors, etc.


4. Security

    - (Optional) JWT-based auth for any endpoints that modify system state.


5. Maintainability

    - Well-documented code, and testing should be developed to enable easy extension and maintenance.

6. Containerization

    - Either just dockerfiles or both dockerfiles and docker-compose files might be required if the system is going to be deployed separately or with multiple services per deployment.


# Possible future extensions/enhancements (only after main deployment is complete)

### Real-Time Analytics & Dashboards

- Integrate a streaming platform like Kafka to handle sensor data ingestion at scale.
- Provide near real-time dashboards showing city-wide traffic congestion.

### Advanced Forecasting

- Use machine learning or time-series analysis to predict traffic volumes.
- Integrate with external data sources (weather, special events) to enhance accuracy.

### Geo-Spatial Queries

- For station and sensor data, consider a GIS-capable database (PostGIS, MongoDB geospatial) for advanced distance/region queries.

### Microservices & CQRS

- Break the monolith into smaller microservices for ingestion, storage, analytics, etc.
- Use a command/query model to separate write-intensive services from read queries.

### Notifications

- Send alerts to users about sudden congestion spikes or accidents.

### CI/CD Pipeline

- Automate builds, tests, and deployments with GitHub Actions, Jenkins, or GitLab CI/CD.
- Container registry for versioned Docker images.
