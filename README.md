# CMPS4191 Quiz 1

## Asynchronous Job + Worker + HTTP 202

| Key               | Value                                          |
| ----------------- | ---------------------------------------------- |
| **Student Name**  | [Andres Hung](https://github.com/andreshungbz) |
| **Student Email** | 2018118240@ub.edu.bz                           |
| **Course**        | CMPS4191 - Advanced Web Technologies           |
| **Due Date**      | August 30, 2026                                |

## Deliverable Links

- B) Slides: [TODO]()
- C) YouTube: [TODO]()

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
