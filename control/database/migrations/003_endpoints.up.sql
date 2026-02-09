-- +migrate Up

CREATE TABLE endpoints(

destination_reference TEXT NOT NULL,
source UUID NOT NULL,

FOREIGN KEY (source) REFERENCES client_user(client_id),
event_type VARCHAR(255),
endpoint_url VARCHAR(255),

PRIMARY KEY (destination_reference, event_type)
);