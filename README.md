# chripy
A web server built in Go. This server allows users to register an account, and create and fetch Chirps. Data is persisted on disk and requires authorization, and authentication to access.

## APIs
### /app/
This api serves static files stored on the server
### GET /api/healthz
This api returns the status of the server
### GET /admin/metrics
This api returns the amount of times */app/* has been hit

### POST /api/chirps
This api allows a user to create a Chirp
Expected Input
`
{
  "body": "I'm the one who knocks!"
}
`
Expected Headers
`
{
  "Authorization": "Bearer {accessToken}"
}
`
Expected Response
`
{
  "id": 5,
  "body": "I'm the one who knocks!",
  "author_id": 1
}
`