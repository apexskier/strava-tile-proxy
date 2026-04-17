# Strava Tile Proxy

This is a simple Go server that proxies my [personal strava heatmap](https://support.strava.com/hc/en-us/articles/216918467-Personal-Heatmaps) and the [global heatmap](https://www.strava.com/heatmap) on a public endpoint in order to use it as a [map source in Gaia GPS](https://help.gaiagps.com/hc/en-us/articles/115003639088-Import-a-Custom-Map-Source). It requires using manual credentials and web scraping, as heatmap data is not available through Strava's OAuth api.

(You can [follow me on Strava here](https://www.strava.com/athletes/14856714))

![Gaia screenshot of tiles](https://user-images.githubusercontent.com/329222/123540346-cc45c200-d73e-11eb-839c-82f447b4d0d1.PNG)

## Configuration & Usage

The server is configured through environment variables:

* `REVEAL_PRIVACY_ZONES` - (bool) reveal [strava privacy zones](https://support.strava.com/hc/en-us/articles/115000173384-Privacy-Zones)
* `REVEAL_ONLY_ME_ACTIVITIES` - (bool) reveal activities only visible to you
* `REVEAL_FOLLOWER_ONLY_ACTIVITIES` - (bool) reveal activities visible to only your followers
* `REVEAL_PUBLIC_ACTIVITIES` - (bool) reveal activities that are public
* `API_TOKEN` - (optional, string) if non-empty, must be present in the `api_token` query parameter on requests
* `STRAVA_REMEMBER_TOKEN` - (string) `strava_remember_token` cookie value, see [authentication below](#authentication)
* `STRAVA4_SESSION` - (string) `strava4_session` cookie value, see [authentication below](#authentication)

Tiles are accessible at the url `/[global|personal]/tiles/{z}/{x}/{y}`, where `z`, `x`, and `y` are standard TMS xyz coordinates, with a maximum `z` of 14 and a minimum of ~6. A query parameters can be used customize tiles:

* `color` (default: "orange" for personal, "blue" for global) - strava heat color ([supported options](./strava/heats.go))
* `sport` (default: "all") - strava sports ([supported options](./strava/sports.go))

### Authentication

Run `go run ./cmd/auth` to generate an `.env.auth` file, which will store the required cookie values you need for authentication. This requires Chrome installed and will run a Chrome instance for you to sign in on.

## Setup

This repo publishes a docker image you can use to run the proxy. I run using docker compose:

```yml
  strava-tile-proxy:
    container_name: strava-tile-proxy
    image: ghcr.io/apexskier/strava-tile-proxy/strava-tile-proxy:latest
    ports:
      - 43503:8080
    environment:
      - STRAVA_EMAIL
      - STRAVA_PASSWORD
      - REVEAL_PRIVACY_ZONES
      - REVEAL_ONLY_ME_ACTIVITIES
      - REVEAL_FOLLOWER_ONLY_ACTIVITIES
      - REVEAL_PUBLIC_ACTIVITIES
    restart: unless-stopped
 ```
