# hetzner library

This folder contains the Hetzner OpenAPI spec as YAML, downloaded from https://github.com/MaximilianKoestler/hcloud-openapi:

```
curl https://raw.githubusercontent.com/MaximilianKoestler/hcloud-openapi/master/openapi/hcloud.json | yq -p json -o yaml > hcloud.yaml
```

Recommended usage is with `openapi-typescript` and `openapi-fetch`:

```
bun add openapi-fetch
bun add -D openapi-typescript typescript
```

The docs also suggest setting `"noUncheckedIndexedAccess": true` in tsconfig compiler options.

Then generate types:

```
bunx openapi-typescript ./lib/hcloud/hcloud.yaml -o ./lib/hcloud/index.d.ts
```

Then use it:

```ts
import createClient from "openapi-fetch";
import type { paths } from "@lib/hcloud";
const client = createClient<paths>({
  headers: { Authorization: `Bearer ${token}` },
  baseUrl: "https://api.hetzner.cloud/v1",
});
const { data, error } = await client.GET("/datacenters");
```

## \*.json / \*.tsx

The JSON files are dumps of endpoints that don't change much, e.g. pricing, as of 2024-04-25.
The tsx files use the typings to export this data as typed data.
It's useful to import these into various things that might need it, e.g. the cluster creation flow.
