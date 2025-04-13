-- +goose Up
-- Create V2X base message table
CREATE TABLE v2x_messages (
	id SERIAL PRIMARY KEY,
	protocol VARCHAR(20) NOT NULL,
	message_type VARCHAR(50) NOT NULL,
	raw_data BYTEA,
	timestamp TIMESTAMP NOT NULL,
	received_at TIMESTAMP NOT NULL,
	rssi SMALLINT,
	source_id VARCHAR(100),
	latitude DOUBLE PRECISION,
	longitude DOUBLE PRECISION,
	elevation DOUBLE PRECISION,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_v2x_messages_timestamp ON v2x_messages(timestamp);
CREATE INDEX idx_v2x_messages_protocol ON v2x_messages(protocol);
CREATE INDEX idx_v2x_messages_message_type ON v2x_messages(message_type);
CREATE INDEX idx_v2x_messages_location ON v2x_messages(latitude, longitude);


-- Create BSM-specific table
CREATE TABLE basic_safety_messages (
	id SERIAL PRIMARY KEY,
	v2x_message_id INTEGER NOT NULL REFERENCES v2x_messages(id) ON DELETE CASCADE,
	temporary_id INTEGER NOT NULL,
	message_count SMALLINT,
	sec_mark SMALLINT,
	speed REAL,
	heading REAL,
	lateral_accel REAL,
	longitudinal_accel REAL,
	vertical_accel REAL,
	yaw_rate REAL,
	brake_applied BOOLEAN,
	traction_control BOOLEAN,
	abs BOOLEAN,
	stability_control BOOLEAN,
	brake_boost BOOLEAN,
	auxiliary_brakes BOOLEAN,
	width REAL,
	length REAL,
	height REAL,
	vehicle_class SMALLINT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_bsm_v2x_message_id ON basic_safety_messages(v2x_message_id);
CREATE INDEX idx_bsm_temporary_id ON basic_safety_messages(temporary_id);

-- Create SPAT table
CREATE TABLE signal_phase_and_timing (
	id SERIAL PRIMARY KEY,
	v2x_message_id INTEGER NOT NULL REFERENCES v2x_messages(id) ON DELETE CASCADE,
	intersection_id INTEGER NOT NULL,
	msg_count SMALLINT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_spat_v2x_message_id ON signal_phase_and_timing(v2x_message_id);
CREATE INDEX idx_spat_intersection_id ON signal_phase_and_timing(intersection_id);

-- Create phase states table for SPAT
CREATE TABLE phase_states (
	id SERIAL PRIMARY KEY,
	spat_message_id INTEGER NOT NULL REFERENCES signal_phase_and_timing(id) ON DELETE CASCADE,
	phase_id SMALLINT NOT NULL,
	light_state VARCHAR(20) NOT NULL,
	start_time SMALLINT,
	min_end_time SMALLINT,
	max_end_time SMALLINT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_phase_states_spat_id ON phase_states(spat_message_id);

-- Create roadside alerts table
CREATE TABLE roadside_alerts (
    id SERIAL PRIMARY KEY,
    v2x_message_id INTEGER NOT NULL REFERENCES v2x_messages(id) ON DELETE CASCADE,
    alert_type INTEGER NOT NULL,
    description TEXT,
    priority SMALLINT,
    radius SMALLINT,
    duration SMALLINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rsa_v2x_message_id ON roadside_alerts(v2x_message_id);
CREATE INDEX idx_rsa_alert_type ON roadside_alerts(alert_type);

-- Create C-V2X specific message table
CREATE TABLE cv2x_messages (
    id SERIAL PRIMARY KEY,
    v2x_message_id INTEGER NOT NULL REFERENCES v2x_messages(id) ON DELETE CASCADE,
    interface_type VARCHAR(20) NOT NULL,
    qos_info SMALLINT,
    plmn_info VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cv2x_v2x_message_id ON cv2x_messages(v2x_message_id);

-- Create security info table
CREATE TABLE v2x_security_info (
    id SERIAL PRIMARY KEY,
    v2x_message_id INTEGER NOT NULL REFERENCES v2x_messages(id) ON DELETE CASCADE,
    signature_valid BOOLEAN NOT NULL,
    certificate_id VARCHAR(100),
    trust_level SMALLINT,
    validation_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_security_v2x_message_id UNIQUE (v2x_message_id)
);

CREATE INDEX idx_security_v2x_message_id ON v2x_security_info(v2x_message_id);

-- Create anomaly detection table
CREATE TABLE v2x_anomaly_detections (
    id SERIAL PRIMARY KEY,
    v2x_message_id INTEGER NOT NULL REFERENCES v2x_messages(id) ON DELETE CASCADE,
    anomaly_type VARCHAR(50) NOT NULL,
    confidence_score REAL NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_anomaly_v2x_message_id ON v2x_anomaly_detections(v2x_message_id);

-- +goose Down
DROP TABLE IF EXISTS v2x_anomaly_detections;
DROP TABLE IF EXISTS v2x_security_info;
DROP TABLE IF EXISTS cv2x_messages;
DROP TABLE IF EXISTS roadside_alerts;
DROP TABLE IF EXISTS phase_states;
DROP TABLE IF EXISTS signal_phase_and_timing;
DROP TABLE IF EXISTS basic_safety_messages;
DROP TABLE IF EXISTS v2x_messages;





























