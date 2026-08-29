# CMPS4191 Quiz 2

## Asynchronous Job + Worker + HTTP 202

| Key               | Value                                          |
| ----------------- | ---------------------------------------------- |
| **Student Name**  | [Andres Hung](https://github.com/andreshungbz) |
| **Student Email** | 2018118240@ub.edu.bz                           |
| **Course**        | CMPS4191 - Advanced Web Technologies           |
| **Due Date**      | August 30, 2026                                |

## Deliverable Links

- B) Slides: [https://docs.google.com/presentation/d/1884qi9FXfv65lyBc8eOx94k_JUAKvcrWkMm7Zqt5ZAQ/edit?usp=sharing](https://docs.google.com/presentation/d/1884qi9FXfv65lyBc8eOx94k_JUAKvcrWkMm7Zqt5ZAQ/edit?usp=sharing)
- C) YouTube Demo: [https://youtu.be/GYW6C-\_avok](https://youtu.be/GYW6C-_avok)

## Running the Application

### Docker Compose

```
docker compose up
```

### Manual Method

#### Prerequisites

- awk
- curl
- go
- golang-migrate
- jq
- make
- PostgreSQL

#### Database Setup

```
CREATE ROLE gatekeeper WITH LOGIN PASSWORD 'password';
CREATE DATABASE gatekeeper;
ALTER DATABASE gatekeeper OWNER TO gatekeeper;
```

#### Application Setup

```
cp .envrc.example .envrc
make db/migrations/up
make run
```

#### Making the Measurement Script Executable and Running It

```
chmod +x measure_async.sh
./measure_async.sh
```
