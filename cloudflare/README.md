# Cloudflare Containers deployment

Runs the proxy as a [Cloudflare Container](https://developers.cloudflare.com/containers/)
behind a Worker at `camlittle.com/strava-tile-proxy/*`: one `lite` instance
that sleeps after 15 idle minutes, so the first tile after that waits for it
to start. The Worker strips the `/strava-tile-proxy` prefix before forwarding.

Configuration is Worker secrets, passed to the container as the environment
variables documented in the top-level README (names in `src/index.js`). Set
them from a JSON object of name/value pairs:

```sh
npx wrangler@4 secret bulk secrets.json
```

Deploy with Docker running; `wrangler deploy` builds `../Dockerfile` and
pushes it to Cloudflare's registry:

```sh
npm ci
npx wrangler@4 deploy
```
