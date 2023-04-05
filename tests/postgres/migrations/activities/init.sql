CREATE TABLE activities (
    id       SERIAL PRIMARY KEY,
    type     VARCHAR(255),
    name     VARCHAR(255)
);
INSERT INTO activities (type, name) VALUES ('movement', 'Migrate South');
INSERT INTO activities (type, name) VALUES ('movement', 'Migrate North');