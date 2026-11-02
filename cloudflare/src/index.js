import { Container, getContainer } from "@cloudflare/containers";

const PREFIX = "/strava-tile-proxy";

// Worker secrets, passed to the container as environment variables; see
// README.md.
const ENV_NAMES = [
  "STRAVA_REMEMBER_TOKEN",
  "STRAVA4_SESSION",
  "API_TOKEN",
  "REVEAL_PRIVACY_ZONES",
  "REVEAL_ONLY_ME_ACTIVITIES",
  "REVEAL_FOLLOWER_ONLY_ACTIVITIES",
  "REVEAL_PUBLIC_ACTIVITIES",
];

export class StravaTileProxy extends Container {
  defaultPort = 8080;
  // Map clients fetch tiles in bursts; stay warm through a browsing session.
  sleepAfter = "15m";

  constructor(ctx, env) {
    super(ctx, env);
    this.envVars = Object.fromEntries(
      ENV_NAMES.filter((name) => env[name] !== undefined).map((name) => [
        name,
        env[name],
      ]),
    );
  }
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname !== PREFIX && !url.pathname.startsWith(`${PREFIX}/`)) {
      return new Response("Not Found", { status: 404 });
    }
    url.pathname = url.pathname.slice(PREFIX.length) || "/";
    const forwarded = new Request(url, request);
    forwarded.headers.set(
      "X-Forwarded-For",
      request.headers.get("CF-Connecting-IP") ?? "",
    );
    return getContainer(env.APP).fetch(forwarded);
  },
};
