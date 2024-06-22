import { $ } from "bun";

$`curl https://raw.githubusercontent.com/MaximilianKoestler/hcloud-openapi/master/openapi/hcloud.json | yq -p json -o yaml > hcloud.yaml`;
$`bunx openapi-typescript ./hcloud.yaml -o ./index.d.ts`;
