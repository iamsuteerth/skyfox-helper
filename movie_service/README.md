# Movie Service

A containerized Golang application that provides movie information via RESTful API endpoints.

## Overview

Movie Service is part of the Skyfox Helper repository. It provides information about movies including titles, ratings, plot summaries, and other metadata. The service is designed to be lightweight, scalable, and easily deployable using containers.

## Features

- RESTful API for accessing movie data
- Fetch all movies in the database
- Get specific movie details by IMDB ID
- API key authentication for secure access
- Health check endpoint
- Structured JSON responses
- Containerized for easy deployment

## API Endpoints

### Health Check
```
GET /mshealth
```
Returns the health status of the service.

### Get All Movies
```
GET /movies
```
Returns a list of all available movies.

### Get Movie by ID
```
GET /movies/:id
```
Returns details for a specific movie by its IMDB ID.

## API Responses

### Successful Response (200 OK)
When requesting a specific movie:
```json
{
    "Title": "A Quiet Place",
    "Year": "2018",
    "Rated": "PG-13",
    "Released": "06 Apr 2018",
    "Runtime": "90 min",
    "Genre": "Drama, Horror, Sci-Fi",
    "Director": "John Krasinski",
    "Writer": "Bryan Woods (screenplay by), Scott Beck (screenplay by), John Krasinski (screenplay by), Bryan Woods (story by), Scott Beck (story by)",
    "Actors": "Emily Blunt, John Krasinski, Millicent Simmonds, Noah Jupe",
    "Plot": "In a post-apocalyptic world, a family is forced to live in silence while hiding from monsters with ultra-sensitive hearing.",
    "Language": "English, American Sign Language",
    "Country": "USA",
    "Awards": "Nominated for 1 Oscar. Another 34 wins & 108 nominations.",
    "Poster": "https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg",
    "Ratings": [
        {
            "Source": "Internet Movie Database",
            "Value": "7.5/10"
        },
        {
            "Source": "Rotten Tomatoes",
            "Value": "95%"
        },
        {
            "Source": "Metacritic",
            "Value": "82/100"
        }
    ],
    "Metascore": "82",
    "imdbRating": "7.5",
    "imdbVotes": "379,472",
    "imdbID": "tt6644200",
    "Type": "movie",
    "DVD": "N/A",
    "BoxOffice": "N/A",
    "Production": "N/A",
    "Website": "N/A",
    "Response": "True"
}
```

### Error Responses

#### Not Found (404)
```json
{
    "error": "Movie with requested ID not found",
    "status": "NOT_FOUND"
}
```

#### Forbidden (403) - Missing API Key
```json
{
    "message": "API key is required",
    "status": "FORBIDDEN"
}
```

#### Forbidden (403) - Invalid API Key
```json
{
    "message": "Invalid API key",
    "status": "FORBIDDEN"
}
```

## Configuration

The service can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Port on which the server listens | 4567 |
| API_KEY | API key for authentication (if empty, authentication is disabled) | "" |
| MOVIES_DATA_PATH | Path to the JSON file containing movie data | "data/movies.json" |
| LOG_LEVEL | Logging level (debug/info) | info |
| APP_VERSION | Application version for health check | "dev" |

## Project Structure

```
.
├── data
│   └── movies.json       # Movie database
├── Dockerfile            # Container configuration
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── internal
│   ├── models
│   │   └── movies.go     # Data models
│   └── services
│       └── movie_service.go  # Business logic
└── server
    └── main.go           # Application entry point
```

## Running Locally

1. Clone the repository
2. Navigate to the movie_service directory
3. Run with Go:
   ```
   go run server/main.go
   ```

## Running with Docker

```bash
# Build the image
docker build -t movie-service .

# Run the container
docker run -p 4567:4567 -e API_KEY=your_api_key iamsuteerth/movie-service:amd64
```

## Authentication

Secure your API by setting the `API_KEY` environment variable. Clients must include this key in the `x-api-key` header when making requests.

