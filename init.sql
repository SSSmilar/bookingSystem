CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       role VARCHAR(50) DEFAULT 'user',
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE rooms (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(100) NOT NULL,
                       capacity INT NOT NULL,
                       description TEXT
);

CREATE TABLE bookings (
                          id SERIAL PRIMARY KEY,
                          room_id INT REFERENCES rooms(id),
                          user_id INT REFERENCES users(id),
                          title VARCHAR(200) NOT NULL,
                          start_time TIMESTAMP NOT NULL,
                          end_time TIMESTAMP NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          CONSTRAINT check_times CHECK (start_time < end_time)
);
INSERT INTO rooms (name, capacity, description)
VALUES ('Переговорка #1 (Mars)', 10, 'Проектор, Маркерная доска, Кондиционер');