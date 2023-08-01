# Strava Tile Proxy

This is a simple Go server that proxies my [personal strava heatmap](https://support.strava.com/hc/en-us/articles/216918467-Personal-Heatmaps) and the [global heatmap](https://www.strava.com/heatmap) on a public endpoint in order to use it as a [map source in Gaia GPS](https://help.gaiagps.com/hc/en-us/articles/115003639088-Import-a-Custom-Map-Source). It requires using manual credentials and web scraping, as heatmap data is not available through Strava's OAuth api.

![Gaia screenshot of tiles](https://user-images.githubusercontent.com/329222/123540346-cc45c200-d73e-11eb-839c-82f447b4d0d1.PNG)

## Configuration & Usage

The server is configured through environment variables, all of which are required:

* `ATHLETE_ID` - your strava athlete ID (https://www.strava.com/athletes/$ATHLETE_ID)
* `STRAVA_EMAIL` - your strava account's email
* `STRAVA_PASSWORD` - your strava account's password (social sign in is not supported)
* `REVEAL_PRIVACY_ZONES` - (bool) reveal [strava privacy zones](https://support.strava.com/hc/en-us/articles/115000173384-Privacy-Zones)
* `REVEAL_ONLY_ME_ACTIVITIES` - (bool) reveal activities only visible to you
* `REVEAL_FOLLOWER_ONLY_ACTIVITIES` - (bool) reveal activities visible to only your followers
* `REVEAL_PUBLIC_ACTIVITIES` - (bool) reveal activities that are public

Tiles are accessible at the url `/tiles/{z}/{x}/{y}`, where `z`, `x`, and `y` are standard TMS xyz coordinates, with a maximum `z` of 14 and a minimum of ~6. A query parameters can be used customize tiles:

* `color` (default: "red") - strava heat color ([supported options](https://github.com/apexskier/strava-tile-proxy/blob/411306d444c0f43f31d8d648a504ec56d2bb7b71/strava/consts.go#L17-L25))
* `sport` (default: "all") - strava sports ([supported options](https://github.com/apexskier/strava-tile-proxy/blob/b1f89caec30ffc7275a3df705cb42fb0c3ebd834/strava/consts.go#L32-L38))

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
