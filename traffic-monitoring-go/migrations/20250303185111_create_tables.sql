

-- +goose Up
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	email VARCHAR(100) NOT NULL UNIQUE,
	hashed_password VARCHAR(255) NOT NULL,
	role VARCHAR(20) NOT NULL DEFAULT 'user'
);

CREATE TABLE stations (
	id SERIAL PRIMARY KEY,
	code VARCHAR(50) NOT NULL UNIQUE,
	name VARCHAR(100),
	city VARCHAR(50),
	latitude FLOAT,
	longitude FLOAT,
	date_of_installation DATE
);

CREATE TABLE sensors (
	id SERIAL PRIMARY KEY,
	sensor_id VARCHAR(50) NOT NULL UNIQUE,
	station_id INTEGER NOT NULL REFERENCES stations(id),
	measurement_type VARCHAR(50),
	status VARCHAR(20) NOT NULL DEFAULT 'active'
);


CREATE TABLE traffic_measurements (
	id SERIAL PRIMARY KEY,
	sensor_id INTEGER NOT NULL REFERENCES sensors(id),
	timestamp TIMESTAMP NOT NULL,
	speed FLOAT,
	vehicle_count INTEGER,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_events (
	id SERIAL PRIMARY KEY,
	date DATE NOT NULL,
	city VARCHAR(50) NOT NULL,
	event_type VARCHAR(50) NOT NULL,
	description TEXT,
	expected_congestion_level VARCHAR(10),
	station_id INTEGER REFERENCES stations(id)
);


-- +goose Down

DROP TABLE IF EXISTS user_events;
DROP TABLE IF EXISTS traffic_measurements;
DROP TABLE IF EXISTS sensors;
DROP TABLE IF EXISTS stations;
DROP TABLE IF EXISTS users;




























