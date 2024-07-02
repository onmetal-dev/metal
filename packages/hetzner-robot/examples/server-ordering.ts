import Instructor from "@instructor-ai/instructor";
import { Mutex } from "async-mutex";
import chalk from "chalk";
import flatCache from "flat-cache";
import { createLLMClient } from "llm-polyglot";
import createClient from "openapi-fetch";
import path from "path";
import { table } from "table";
import { z } from "zod";
import type { components, paths } from "../gen/hetzner-robot";

const cache = flatCache.load("anthropic", path.resolve("./.flat-cache"));

const username = process.env.HETZNER_ROBOT_USERNAME!;
const password = process.env.HETZNER_ROBOT_PASSWORD!;
const authorizedKey = process.env.AUTHORIZED_KEY!; // ssh-keygen -E md5 -lf .ssh/hetzner.pub

const client = createClient<paths>({
  baseUrl: "https://robot-ws.your-server.de",
  headers: {
    Authorization: `Basic ${Buffer.from(`${username}:${password}`).toString(
      "base64"
    )}`,
  },
});

const { data, response: _ } = await client.GET("/order/server/product");
if (!data) {
  throw new Error("No data returned about available servers to order");
}

// things worth filtering on:
// things worth displaying: id, name, location (array), prices (array for each location), description
// things worth parsing:
// - Would be nice to parse the description for # of CPUs and RAM
// const productsAll = data.products
// let productsFiltered = []

console.log(chalk.green(`Found ${data.length} products`));
console.log(chalk.green("Filtering out products without IPv4"));
let filteredProducts = data.filter(({ product }) => {
  return product!.orderable_addons?.some(
    (addon) => addon.id === "primary_ipv4"
  );
});
console.log(chalk.green(`Found ${filteredProducts.length} products with IPv4`));

const preferredDist = "Ubuntu 22.04.2 LTS base";
console.log(chalk.green(`Filtering out products without ${preferredDist}`));
filteredProducts = filteredProducts.filter(({ product }) => {
  return product!.dist?.some((dist) => dist === preferredDist);
});

console.log(
  chalk.green(`Found ${filteredProducts.length} products with ${preferredDist}`)
);

const anthropicClient = createLLMClient({
  provider: "anthropic",
  apiKey: process.env.ANTHROPIC_API_KEY ?? undefined,
  baseURL: "https://anthropic.helicone.ai",
  defaultHeaders: {
    "Helicone-Auth": `Bearer ${process.env.HELICONE_API_KEY}`,
  },
});

// const model = "gpt-4o";
const model = "claude-3-5-sonnet-20240620";
const max_tokens = 4096;

const instructor = Instructor({
  //  client: oaiClient,
  client: anthropicClient,
  mode: "TOOLS",
});

const Spec = z.object({
  cpu: z.number().describe("Number of CPU cores"),
  ram: z.number().describe("Amount of RAM in GB"),
  storage: z.number().describe("Amount of storage in TB"),
});
type Spec = z.infer<typeof Spec>;

const mutex = new Mutex();

async function parseDescription(description: string): Promise<Spec> {
  // cache key is base64 encoded description
  const cacheKey = Buffer.from(description).toString("base64");
  const cached = cache.getKey(cacheKey) as Spec;
  if (cached) {
    return cached;
  }
  return await mutex.runExclusive(async () => {
    const response = await instructor.chat.completions.create({
      messages: [
        {
          role: "system",
          content:
            "You are a helpful assistant that parses a description of a server and returns the number of CPU cores and the amount of RAM in GB.",
        },
        {
          role: "user",
          content: `Description: ${description}`,
        },
      ],
      model,
      response_model: {
        schema: Spec,
        name: "Spec",
      },
      max_retries: 20,
      max_tokens,
    });
    cache.setKey(cacheKey, response);
    cache.save(true);
    return response;
  });
}

// print out table of id, name, locations, monthly price in each location, description
function highestMonthlyPrice(
  product: components["schemas"]["Product"]
): number {
  return product
    .prices!.map((priceForLoc) => parseFloat(priceForLoc.price!.gross!))
    .reduce((a, b) => Math.max(a, b), 0);
}
console.log(
  table([
    [
      "id",
      "name",
      "locations",
      "monthly price",
      "description",
      "CPU",
      "RAM (GB)",
      "Storage (TB)",
      "€/CPU",
      "€/RAM",
      "€/TB",
    ],
    ...(await Promise.all(
      filteredProducts
        .sort(
          (a, b) =>
            highestMonthlyPrice(a.product!) - highestMonthlyPrice(b.product!)
        )
        .map(async ({ product }) => {
          const specs = await parseDescription(product!.description!.join(" "));
          const highestMonthlyP = highestMonthlyPrice(product!);
          return [
            product!.id,
            product!.name,
            product!.location?.sort().join(", "),
            product?.prices
              ?.sort((a, b) => a.location!.localeCompare(b.location!))
              ?.map(
                (priceForLoc) =>
                  `${priceForLoc.location}: ${parseFloat(
                    priceForLoc.price?.gross!
                  ).toFixed(2)}`
              )
              .join(", "),
            product!.description,
            specs.cpu,
            specs.ram,
            specs.storage,
            `€ ${(highestMonthlyP / specs.cpu).toFixed(2)}`,
            `€ ${(highestMonthlyP / specs.ram).toFixed(2)}`,
            `€ ${(highestMonthlyP / specs.storage).toFixed(2)}`,
          ];
        })
    )),
  ])
);

// let's buy an EX44 (TODO: this doesn't work)
const { data: transaction, response } = await client.POST(
  "/order/server/transaction",
  {
    body: {
      product_id: "EX44",
      location: "FSN1",
      addon: ["primary_ipv4"],
      dist: preferredDist,
      authorized_key: [authorizedKey],
      lang: "en",
      test: "true",
    },
    bodySerializer(body) {
      const bodyCopy = JSON.parse(JSON.stringify(body));
      bodyCopy["authorized_key[]"] = bodyCopy["authorized_key"];
      delete bodyCopy["authorized_key"];
      bodyCopy["addon[]"] = bodyCopy["addon"];
      delete bodyCopy["addon"];
      const params = new URLSearchParams(JSON.parse(JSON.stringify(body)));
      console.log(params.toString());
      return params.toString();
    },
  }
);
console.log(transaction);
console.log(response.status);
console.log(response.statusText);
console.log(await response.text());
