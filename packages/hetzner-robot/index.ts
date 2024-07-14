import Instructor from "@instructor-ai/instructor";
import chalk from "chalk";
import { createLLMClient } from "llm-polyglot";
import { z } from "zod";
import OpenAPISchema, {
  Components,
  Paths,
  mergeComponents,
} from "./oai3schema";
type OpenAPISchema = z.infer<typeof OpenAPISchema>;

const docs = Bun.file("firecrawl.md");
const lines = (await docs.text()).split("\n");

function getSection(name: "preface" | "server" | "server ordering"): string[] {
  switch (name) {
    case "preface":
      return lines.slice(144, 224);
    case "server":
      return lines.slice(225, 657);
    case "server ordering":
      return lines.slice(4290, 6177);
  }
}

const oaiClient = createLLMClient({
  provider: "openai",
  apiKey: process.env.OPENAI_API_KEY ?? undefined,
  organization: process.env.OPENAI_ORG_ID ?? undefined,
  baseURL: "https://oai.helicone.ai/v1",
  defaultHeaders: {
    "Helicone-Auth": `Bearer ${process.env.HELICONE_API_KEY}`,
  },
});

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

const client = Instructor({
  //  client: oaiClient,
  client: anthropicClient,
  mode: "TOOLS",
});

const componentsModel = z.object({
  components: Components.superRefine((components, ctx) => {
    // must have schemas and responses
    if (!components.schemas) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Components must include schemas",
      });
    }
    if (!components.responses) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Components must include responses",
      });
    }
  }),
});
type ComponentsModel = z.infer<typeof componentsModel>;

function componentRefs(components: Components): string[] {
  return Object.keys(components.schemas ?? {})
    .map((name) => `#/components/schemas/${name}`)
    .concat(
      Object.keys(components.responses ?? {}).map(
        (name) => `#/components/responses/${name}`
      )
    )
    .concat(
      Object.keys(components.parameters ?? {})
        .map((name) => `#/components/parameters/${name}`)
        .concat(
          Object.keys(components.requestBodies ?? {}).map(
            (name) => `#/components/requestBodies/${name}`
          )
        )
        .concat(
          Object.keys(components.securitySchemes ?? {}).map(
            (name) => `#/components/securitySchemes/${name}`
          )
        )
        .concat(
          Object.keys(components.examples ?? {}).map(
            (name) => `#/components/examples/${name}`
          )
        )
        .concat(
          Object.keys(components.links ?? {}).map(
            (name) => `#/components/links/${name}`
          )
        )
        .concat(
          Object.keys(components.callbacks ?? {}).map(
            (name) => `#/components/callbacks/${name}`
          )
        )
        .concat(
          Object.keys(components.headers ?? {}).map(
            (name) => `#/components/headers/${name}`
          )
        )
    );
}

async function generateComponents({
  componentsCreated,
  prompt,
  cache,
}: {
  componentsCreated: Components;
  prompt: string;
  cache: string;
}): Promise<ComponentsModel> {
  if (await Bun.file(cache).exists()) {
    return JSON.parse(await Bun.file(cache).text());
  }
  const messages = [
    {
      role: "system",
      content:
        "You are a world class generator of OpenAPI component definitions. You will be given a description of an API and you will generate the OpenAPI components to support the API: schemas, parameters, request bodies, and responses. Make sure to generate components for all of these. There should also be a securitySchema component if the API has an authentication mechanism." +
        (Object.keys(componentsCreated).length > 0
          ? ` You have already generated the following components: ${componentRefs(
              componentsCreated
            ).join(
              ", "
            )}. You do not need to generate them again. You can reference them in other schemas using $ref, e.g. $ref: #/components/schemas/Server. You can also use references among the components you generate your next response.`
          : ""),
    },
    {
      role: "user",
      content: prompt,
    },
  ];
  const response = await client.chat.completions.create({
    messages,
    model,
    response_model: {
      schema: componentsModel,
      name: "Components",
    },
    max_retries: 20,
  });
  await Bun.write(cache, JSON.stringify(response, null, 2));
  return response;
}

async function generatePaths({
  componentsCreated,
  prompt,
  cache,
}: {
  componentsCreated: Components;
  prompt: string;
  cache: string;
}): Promise<Paths> {
  if (await Bun.file(cache).exists()) {
    return JSON.parse(await Bun.file(cache).text()).paths;
  }
  const messages = [
    {
      role: "system",
      content:
        "You are a world class generator of OpenAPI path definitions. You will be given a description of an API and you will generate the OpenAPI path definitions for the API. In OpenAPI terms, paths are endpoints (resources), such as /users or /reports/summary/, that your API exposes, and operations are the HTTP methods used to manipulate these paths, such as GET, POST or DELETE. Here's an example of a path definition in JSON: " +
        `
{
  "paths": {
    "/users/{id}": {
      "get": {
        "tags": [
          "Users"
        ],
        "summary": "Gets a user by ID.",
        "description": "A detailed description of the operation. Use markdown for rich text representation, such as **bold**, *italic*, and [links](https://swagger.io).",
        "operationId": "getUserById",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "User ID",
            "required": true,
            "schema": {
              "type": "integer",
              "format": "int64"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful operation",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/User"
                }
              }
            }
          }
        },
        "externalDocs": {
          "description": "Learn more about user operations provided by this API.",
          "url": "http://api.example.com/docs/user-operations/"
        }
      }
    }
  }
}
` +
        (Object.keys(componentsCreated).length > 0
          ? ` You also have the following components to use: ${componentRefs(
              componentsCreated
            ).join(
              ", "
            )}. You can reference them in your path definitions using $ref, e.g. $ref: #/components/schemas/Server`
          : ""),
    },
    {
      role: "user",
      content: prompt,
    },
  ];
  const response = await client.chat.completions.create({
    messages,
    model,
    response_model: {
      schema: OpenAPISchema.pick({
        paths: true,
      }),
      name: "OpenAPI",
    },
    max_retries: 20,
  });
  await Bun.write(cache, JSON.stringify(response, null, 2));
  return response.paths;
}

const componentsAndPaths = z.object({
  components: Components.superRefine((components, ctx) => {
    // must have schemas and responses
    if (!components.schemas) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Components must include schemas",
      });
    }
    if (!components.responses) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Components must include responses",
      });
    }
  }),
  paths: Paths,
});

type ComponentsAndPathsModel = z.infer<typeof componentsAndPaths>;

async function generateComponentsAndPaths({
  componentsCreated,
  prompt,
  cache,
}: {
  componentsCreated: Components;
  prompt: string;
  cache: string;
}): Promise<ComponentsAndPathsModel> {
  if (await Bun.file(cache).exists()) {
    return JSON.parse(await Bun.file(cache).text());
  }
  const messages = [
    {
      role: "system",
      content:
        "You are a world class generator of OpenAPI path and component definitions. You will be given a description of an API and you will generate the OpenAPI path and component definitions for the API. In OpenAPI terms, paths are endpoints (resources), such as /users or /reports/summary/, that your API exposes, and operations are the HTTP methods used to manipulate these paths, such as GET, POST or DELETE. Components are the various schemas for responses, parameters, request bodies, etc. Here's an example of a path and component definition in JSON: " +
        `
{
  "paths": {
    "/users/{id}": {
      "get": {
        "tags": [
          "Users"
        ],
        "summary": "Gets a user by ID.",
        "description": "A detailed description of the operation. Use markdown for rich text representation, such as **bold**, *italic*, and [links](https://swagger.io).",
        "operationId": "getUserById",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "User ID",
            "required": true,
            "schema": {
              "type": "integer",
              "format": "int64"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful operation",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/User"
                }
              }
            }
          }
        },
        "externalDocs": {
          "description": "Learn more about user operations provided by this API.",
          "url": "http://api.example.com/docs/user-operations/"
        }
      }
    }
  }
}
` +
        (Object.keys(componentsCreated).length > 0
          ? ` You also have the following components to use as via $ref if necessary: ${JSON.stringify(
              componentsCreated
            )}. You can reference them in your path definitions using $ref, e.g. $ref: #/components/schemas/Server. All the new components you create must not conflict with these.`
          : ""),
    },
    {
      role: "user",
      content: prompt,
    },
  ];
  const response = await client.chat.completions.create({
    messages,
    model,
    response_model: {
      schema: componentsAndPaths.superRefine((candp, ctx) => {
        // double check that the new components do not conflict with the existing ones
        const newComponents = componentRefs(candp.components);
        const existingComponents = componentRefs(componentsCreated);
        const conflicts = newComponents.filter((component) =>
          existingComponents.includes(component)
        );
        if (conflicts.length > 0) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: `New component definitions conflict with existing components. Please rename these or omit them if you can reuse the existing component definition with the same name: ${conflicts.join(
              ", "
            )}`,
          });
        }
      }),
      name: "OpenAPICP",
    },
    max_tokens,
    max_retries: 20,
  });
  await Bun.write(cache, JSON.stringify(response, null, 2));
  return response;
}

async function runInstructor() {
  const oaiSchema: OpenAPISchema = {
    openapi: "3.0.0",
    info: {
      title: "Hetzner Robot API",
      description: "API for managing servers and orders on Hetzner.",
      termsOfService: "https://www.hetzner.com/legal/terms-of-service",
      contact: {
        name: "Hetzner Online GmbH",
        url: "https://www.hetzner.com",
        email: "info@hetzner.com",
      },
      license: {
        name: "Apache 2.0",
        url: "https://www.apache.org/licenses/LICENSE-2.0",
      },
      version: "1.0.0",
    },
    servers: [
      {
        url: "https://robot-ws.your-server.de",
        description: "Hetzner Robot API Server",
      },
    ],
    tags: [
      { name: "Servers", description: "Manage servers" },
      { name: "Server Ordering", description: "Order server products" },
    ],
    paths: {},
    components: {},
  };
  let componentsCreated: Components = {};

  // server API
  const serverComponents = await generateComponents({
    componentsCreated,
    prompt: `I will describe the API to you using snippets from the API docs. Here is the preface plus the APIs for servers. Generate components for these sections, placing the URI of the components generated in the componentsCreated array:
   ${getSection("preface").join("\n")}
   ${getSection("server").join("\n")}`,
    cache: "cache/cache-server-components.json",
  });

  console.log("components created", componentRefs(serverComponents.components));
  componentsCreated = mergeComponents(
    componentsCreated,
    serverComponents.components
  );
  oaiSchema.components = mergeComponents(
    oaiSchema.components as Components,
    serverComponents.components
  );

  const serverPaths = await generatePaths({
    componentsCreated,
    prompt: `I will describe the API to you using snippets from the API docs. Here is the preface plus the APIs for servers. Generate paths for these sections:
---
${getSection("preface").join("\n")}
${getSection("server").join("\n")}
---
Make sure to include endpoint definitions for each of these operations:
GET /server
GET /server/{server-number}
POST /server/{server-number}
GET /server/{server-number}/cancellation
POST /server/{server-number}/cancellation
DELETE /server/{server-number}/cancellation
`,
    cache: "cache/cache-server-paths.json",
  });
  for (const path in serverPaths) {
    oaiSchema.paths[path] = serverPaths[path];
  }

  // server ordering api
  const paths = [
    {
      path: "/order/server/product",
      operations: ["get"],
      additionalPrompt: `The response to GET /order/server/product is an array of objects, and each object has a single "product" key. Do not miss this important detail in the nested structure.`,
    },
    {
      path: "/order/server/product/{product-id}",
      operations: ["get"],
    },
    {
      path: "/order/server/transaction",
      operations: ["get"],
    },
    {
      path: "/order/server/transaction",
      operations: ["post"],
      additionalPrompt:
        "Property names should not contain brackets [], spaces, or the word @deprecated. If you see those in the docs, they are simply annotations describing the nature of the property.",
    },
    {
      path: "/order/server/transaction/{id}",
      operations: ["get"],
    },
    {
      path: "/order/server_market/product",
      operations: ["get"],
    },
    {
      path: "/order/server_market/product/{product-id}",
      operations: ["get"],
    },
    {
      path: "/order/server_market/transaction",
      operations: ["get"],
    },
    {
      path: "/order/server_market/transaction",
      operations: ["post"],
    },
    {
      path: "/order/server_market/transaction/{id}",
      operations: ["get"],
    },
    {
      path: "/order/server_addon/{server-number}/product",
      operations: ["get"],
    },
    {
      path: "/order/server_addon/transaction",
      operations: ["get"],
    },
    {
      path: "/order/server_addon/transaction",
      operations: ["post"],
    },
    {
      path: "/order/server_addon/transaction/{id}",
      operations: ["get"],
    },
  ];

  for (const path of paths) {
    console.log(
      chalk.green(
        `Generating path and components for [${path.operations.join(", ")}] ${
          path.path
        }`
      )
    );
    const serverOrdering = await generateComponentsAndPaths({
      componentsCreated,
      prompt: `I will describe the API to you using snippets from the API docs. Here is the section for server ordering. Generate components (schemas for responses, parameters, request bodies, etc.) and a path definition (with operations for each method) for this section:
---
${getSection("server ordering").join("\n")}
---
For now let's focus on the endpoint definitions for just:
${path.operations
  .map((operation) => `${operation.toUpperCase()} ${path.path}`)
  .join("\n")}
---
${
  path.additionalPrompt
    ? `Important note for this part of the API: ${path.additionalPrompt}`
    : ""
}
  `,
      cache: `cache/cache-server-ordering-components-paths-${path.operations
        .map(
          (operation) =>
            operation.toUpperCase() + "-" + path.path.replaceAll("/", "-")
        )
        .join("-")}.json`,
    });
    for (const path in serverOrdering.paths) {
      oaiSchema.paths[path] = serverOrdering.paths[path];
    }
    componentsCreated = mergeComponents(
      componentsCreated,
      serverOrdering.components
    );
    oaiSchema.components = mergeComponents(
      oaiSchema.components,
      serverOrdering.components
    );
  }

  Bun.write("hetzner-robot-openapi.json", JSON.stringify(oaiSchema, null, 2));
}

runInstructor().catch(console.error);
