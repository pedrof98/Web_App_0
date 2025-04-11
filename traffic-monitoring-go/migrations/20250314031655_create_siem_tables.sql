-- +goose Up
CREATE TABLE log_sources (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	type VARCHAR(50) NOT NULL,
	description TEXT,
	enabled BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE security_events (
	id SERIAL PRIMARY KEY,
	timestamp TIMESTAMP NOT NULL,
	source_ip VARCHAR(45),
	source_port INTEGER,
	destination_ip VARCHAR(45),
	destination_port INTEGER,
	protocol VARCHAR(20),
	action VARCHAR(20),
	status VARCHAR(20),
	user_id INTEGER REFERENCES users(id),
	device_id VARCHAR(100),
	log_source_id INTEGER NOT NULL REFERENCES log_sources(id),
	severity VARCHAR(20) NOT NULL,
	category VARCHAR(30) NOT NULL,
	message TEXT NOT NULL,
	raw_data TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_security_events_timestamp ON security_events(timestamp);
CREATE INDEX idx_security_events_severity ON security_events(severity);
CREATE INDEX idx_security_events_category ON security_events(category);


CREATE TABLE rules (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL UNIQUE,
	description TEXT,
	condition TEXT NOT NULL,
	severity VARCHAR(20) NOT NULL,
	status VARCHAR(20) NOT NULL,
	created_by INTEGER NOT NULL REFERENCES users(id),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE alerts (
	id SERIAL PRIMARY KEY,
	rule_id INTEGER NOT NULL REFERENCES rules(id),
	security_event_id INTEGER NOT NULL REFERENCES security_events(id),
	timestamp TIMESTAMP NOT NULL,
	severity VARCHAR(20) NOT NULL,
	status VARCHAR(20) NOT NULL,
	assigned_to INTEGER REFERENCES users(id),
	resolution TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_alerts_timestamp ON alerts(timestamp);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_status ON alerts(status);

-- +goose Down
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS security_events;
DROP TABLE IF EXISTS log_sources;


