This is a [Next.js](https://nextjs.org/) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

## Developing

- `git clone --recurse-submodules`
- `bun install`
- `cp .env.example .env` and follow the directions in the file to populate the values
- `bun dev`

The app mostly generates traces instead of logs, and the `bun dev` command will spin up a local opentelemetry setup to collect the traces.
Navigate to the following URLs to check them out:

- Jaeger at [http://0.0.0.0:16686](http://0.0.0.0:16686)
- Zipkin at [http://0.0.0.0:9411](http://0.0.0.0:9411)
- Prometheus at [http://0.0.0.0:9090](http://0.0.0.0:9090)

## Deploying to shared dev environment

- `railway link` and choose the dev environment and website.
- `railway up` => launches your local code for the website in the dev env
- or `railway up -s worker` => launches your local code for the worker in the dev env
- You might also need to `railway shell -s website` and `bun run db:push` in order for the dev environment's configured database (the dev tembo instance's `metalprod` schema) to be updated
