# Strava Tile Proxy

This is a simple Go server that proxies my [personal strava heatmap](https://www.google.com/search?q=strava+personal+heatmap&oq=strava+personal+heatmap&aqs=chrome..69i57j69i60l2&sourceid=chrome&ie=UTF-8) on a public endpoint in order to use it as a [map source in Gaia GPS](https://www.google.com/search?q=gaia+gps+cusotm+map+source&oq=gaia+gps+cusotm+map+source&aqs=chrome..69i57j33i10i22i29i30&sourceid=chrome&ie=UTF-8).

The server is configured through environment variables, all of which are required:

* `STRAVA_EMAIL` - your strava account's email
* `STRAVA_PASSWORD` - your strava account's password (social sign in is not supported)
* `REVEAL_PRIVACY_ZONES` - (bool) reveal [strava privacy zones](https://support.strava.com/hc/en-us/articles/115000173384-Privacy-Zones)
* `REVEAL_ONLY_ME_ACTIVITIES` - (bool) reveal activities only visible to you
* `REVEAL_FOLLOWER_ONLY_ACTIVITIES` - (bool) reveal activities visible to only your followers
* `REVEAL_PUBLIC_ACTIVITIES` - (bool) reveal activities that are public

Tiles are accessible at the url `/tiles/{z}/{x}/{y}`, where `z`, `x`, and `y` are standard TMS xyz coordinates. A query parameters can be used customize tiles:

* `color` (default: "hot") - strava heat color ([supported options](https://github.com/apexskier/strava-tile-proxy/blob/b1f89caec30ffc7275a3df705cb42fb0c3ebd834/strava/consts.go#L9-L16))
* `sport` (default: "all") - strava sports ([supported options](https://github.com/apexskier/strava-tile-proxy/blob/b1f89caec30ffc7275a3df705cb42fb0c3ebd834/strava/consts.go#L32-L38))