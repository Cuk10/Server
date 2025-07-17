# Chirpy Backend

## Description

This project is a Go-based backend server for a simple social media application called Chirpy. It provides API endpoints for user authentication, chirp management, and more.

## Features

*   User authentication and management (creation, login, update).
*   Chirp creation, retrieval, and deletion.
*   JWT-based authentication with refresh tokens.
*   Database interaction using PostgreSQL.
*   Admin endpoints for metrics and resetting the environment.
*   Rate limiting.

## Getting Started

### Prerequisites

*   Go
*   PostgreSQL

### Installation

1.  Clone the repository.
2.  Set up your PostgreSQL database and update the `DB_URL` environment variable in the `.env` file.
3.  Build project:

    ```bash
    go build -o server
    ```

### Running the Server

```bash
./server
```

The server will start on port 8080.

## API Endpoints

*   `GET /api/healthz`: Health check endpoint.
*   `POST /admin/reset`: Resets the environment (for development).
*   `POST /api/users`: Creates a new user.
*   `PUT /api/users`: Updates an existing user.
*   `POST /api/login`: Logs in a user.
*   `POST /api/refresh`: Refreshes a JWT token.
*   `POST /api/revoke`: Revokes a refresh token.
*   `GET /api/chirps`: Retrieves all chirps.
*   `POST /api/chirps`: Creates a new chirp.
*   `GET /api/chirps/{chirpID}`: Retrieves a specific chirp.
*   `DELETE /api/chirps/{chirpID}`: Deletes a specific chirp.

## Environment Variables

*   `DB_URL`: PostgreSQL database connection URL.
*   `PLATFORM`: Platform the application is running on.
*   `SECRET`: Secret key for JWT signing.
*   `POLKA_KEY`: API key for Polka webhooks.

