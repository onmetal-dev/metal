// this file is 90% sonnet-3.5, 10% editing it to fix mistakes
import { z } from "zod";

// Helper schemas
const Reference = z.object({
  $ref: z.string().min(1, "$ref must be nonempty"),
});
type Reference = z.infer<typeof Reference>;

const ExampleXORExamples = z
  .object({
    example: z.any().optional(),
    examples: z.record(z.any()).optional(),
  })
  .refine((data) => !(data.example && data.examples), {
    message: "Example and examples are mutually exclusive",
  });

const SchemaXORContent = z
  .object({
    schema: z.any().optional(),
    content: z.any().optional(),
  })
  .describe(
    "Schema and content are mutually exclusive, at least one is required"
  )
  .refine((data) => !(data.schema && data.content), {
    message: "Schema and content are mutually exclusive",
  })
  .refine((data) => data.schema || data.content, {
    message: "Either schema or content is required",
  });

// Sub-schemas
const Contact = z.object({
  name: z.string().optional(),
  url: z.string().url().optional(),
  email: z.string().email().optional(),
});

const License = z.object({
  name: z.string(),
  url: z.string().url().optional(),
});

const Info = z.object({
  title: z.string(),
  description: z.string().optional(),
  termsOfService: z.string().url().optional(),
  contact: Contact.optional(),
  license: License.optional(),
  version: z.string(),
});

const ServerVariable = z.object({
  enum: z.array(z.string()).optional(),
  default: z.string(),
  description: z.string().optional(),
});

const Server = z.object({
  url: z.string(),
  description: z.string().optional(),
  variables: z.record(z.string(), ServerVariable).optional(),
});
// .catchall(z.any());

const ExternalDocumentation = z.object({
  description: z.string().optional(),
  url: z.string().url(),
});

const baseSchema = z.object({
  title: z.string().optional(),
  multipleOf: z.number().positive().optional(),
  maximum: z.number().optional(),
  exclusiveMaximum: z.boolean().optional(),
  minimum: z.number().optional(),
  exclusiveMinimum: z.boolean().optional(),
  maxLength: z.number().int().nonnegative().optional(),
  minLength: z.number().int().nonnegative().optional(),
  pattern: z.string().optional(),
  maxItems: z.number().int().nonnegative().optional(),
  minItems: z.number().int().nonnegative().optional(),
  uniqueItems: z.boolean().optional(),
  maxProperties: z.number().int().nonnegative().optional(),
  minProperties: z.number().int().nonnegative().optional(),
  required: z.array(z.string()).optional(),
  enum: z.array(z.any()).optional(),
  type: z
    .enum(["array", "boolean", "integer", "number", "object", "string"])
    .optional(),
  // not: z.union([Schema, Reference]).optional(),
  // allOf: z.array(z.union([Schema, Reference])).optional(),
  // oneOf: z.array(z.union([Schema, Reference])).optional(),
  // anyOf: z.array(z.union([Schema, Reference])).optional(),
  // items: z.union([Schema, Reference]).optional(),
  // properties: z.record(z.union([Schema, Reference])).optional(),
  // additionalProperties: z
  //   .union([Schema, Reference, z.boolean()])
  //   .optional(),
  description: z.string().optional(),
  format: z.string().optional(),
  default: z.any().optional(),
  nullable: z.boolean().optional(),
  discriminator: z
    .object({
      propertyName: z.string(),
      mapping: z.record(z.string(), z.any()).optional(),
    })
    .optional(),
  readOnly: z.boolean().optional(),
  writeOnly: z.boolean().optional(),
  example: z.any().optional(),
  externalDocs: ExternalDocumentation.optional(),
  deprecated: z.boolean().optional(),
  xml: z
    .object({
      name: z.string().optional(),
      namespace: z.string().url().optional(),
      prefix: z.string().optional(),
      attribute: z.boolean().optional(),
      wrapped: z.boolean().optional(),
    })
    .optional(),
});

export type Schema = z.infer<typeof baseSchema> & {
  not?: Schema | Reference;
  allOf?: (Schema | Reference)[];
  oneOf?: (Schema | Reference)[];
  anyOf?: (Schema | Reference)[];
  items?: Schema | Reference;
  properties?: Record<string, Schema | Reference>;
  additionalProperties?: Schema | Reference | boolean;
};

const Schema: z.ZodType<Schema> = baseSchema
  .extend({
    not: z.lazy(() => z.union([Schema, Reference])).optional(),
    allOf: z.lazy(() => z.array(z.union([Schema, Reference]))).optional(),
    oneOf: z.lazy(() => z.array(z.union([Schema, Reference]))).optional(),
    anyOf: z.lazy(() => z.array(z.union([Schema, Reference]))).optional(),
    items: z.lazy(() => z.union([Schema, Reference])).optional(),
    properties: z
      .lazy(() => z.record(z.string(), z.union([Schema, Reference])))
      .optional()
      .superRefine((properties, ctx) => {
        if (!properties) {
          return;
        }
        if (Object.keys(properties).length === 0) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Properties must have at least one property",
          });
        }
      }),
    additionalProperties: z
      .lazy(() => z.union([Schema, Reference, z.boolean()]))
      .optional(),
  })
  .superRefine((schema, ctx) => {
    // cannot have a completely empty schema
    if (Object.keys(schema).length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Schema must have at least one property",
      });
    }
    if (
      schema.type === "array" &&
      (!schema.items || Object.keys(schema.items).length === 0)
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Array schema must have an non-empty items property",
      });
    } else if (schema.type === "object" && !schema.properties) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Object schema must have a properties",
      });
    }
  });

const Example = z.object({
  summary: z.string().optional(),
  description: z.string().optional(),
  value: z.any().optional(),
  externalValue: z.string().url().optional(),
});

const MediaType = z
  .object({
    schema: z.union([Schema, Reference]).optional(),
    example: z.any().optional(),
    examples: z.record(z.string(), z.union([Example, Reference])).optional(),
    encoding: z.record(z.any() /* TODO recursive type */).optional(),
  })
  .and(ExampleXORExamples);

export const SimplifiedMediaTypeForResponses = z.union([
  z.object({
    schema: Reference,
  }),
  z.object({
    schema: z.object({
      type: z.literal("array"),
      items: Reference,
    }),
  }),
]);

const Header = z
  .object({
    description: z.string().optional(),
    required: z.boolean().optional(),
    deprecated: z.boolean().optional(),
    allowEmptyValue: z.boolean().optional(),
    style: z.enum(["simple"]).optional(),
    explode: z.boolean().optional(),
    allowReserved: z.boolean().optional(),
    schema: z.union([Schema, Reference]).optional(),
    content: z.record(MediaType).optional(),
    example: z.any().optional(),
    examples: z.record(z.string(), z.union([Example, Reference])).optional(),
  })
  .and(ExampleXORExamples)
  .and(SchemaXORContent);

// TODO recursive type Encoding
// const HeaderOrReference = z.lazy(() => z.union([Header, Reference]));
//   const Encoding = z
//     .object({
//       contentType: z.string().optional(),
//       headers: z.record(HeaderOrReference).optional(),
//       style: z
//         .enum(["form", "spaceDelimited", "pipeDelimited", "deepObject"])
//         .optional(),
//       explode: z.boolean().optional(),
//       allowReserved: z.boolean().default(false).optional(),
//     })
//     .catchall(z.unknown());

const Link = z
  .object({
    operationId: z.string().optional(),
    operationRef: z.string().optional(),
    parameters: z.record(z.any()).optional(),
    requestBody: z.any().optional(),
    description: z.string().optional(),
    server: Server.optional(),
  })
  .refine((data) => !(data.operationId && data.operationRef), {
    message: "Operation Id and Operation Ref are mutually exclusive",
  });

export const Response = z.object({
  description: z.string(),
  headers: z.record(z.string(), z.union([Header, Reference])).optional(),
  content: z.record(z.string(), SimplifiedMediaTypeForResponses).optional(),
  //    content: z.record(z.string(), MediaType).optional(),
  links: z.record(z.string(), z.union([Link, Reference])).optional(),
});

const ParameterLocation = z.discriminatedUnion("in", [
  z
    .object({
      in: z.literal("path"),
      style: z.enum(["matrix", "label", "simple"]).default("simple"),
      required: z.literal(true),
    })
    .describe("Parameter in path"),

  z
    .object({
      in: z.literal("query"),
      style: z
        .enum(["form", "spaceDelimited", "pipeDelimited", "deepObject"])
        .default("form"),
    })
    .describe("Parameter in query"),

  z
    .object({
      in: z.literal("header"),
      style: z.literal("simple").default("simple"),
    })
    .describe("Parameter in header"),

  z
    .object({
      in: z.literal("cookie"),
      style: z.literal("form").default("form"),
    })
    .describe("Parameter in cookie"),
]);

const Parameter = z
  .object({
    name: z.string(),
    in: z.enum(["path", "query", "header", "cookie"]),
    description: z.string().optional(),
    required: z.boolean().default(false).optional(),
    deprecated: z.boolean().default(false).optional(),
    allowEmptyValue: z.boolean().default(false).optional(),
    style: z.string().optional(),
    explode: z.boolean().optional(),
    allowReserved: z.boolean().default(false).optional(),
    schema: z.union([Schema, Reference]).optional(),
    content: z
      .record(z.string(), MediaType)
      .refine((val) => Object.keys(val!).length === 1, {
        message: "content must have exactly one property",
      })
      .optional(),

    example: z.any().optional(),
    examples: z.record(z.string(), z.union([Example, Reference])).optional(),
  })
  .and(ExampleXORExamples)
  .and(SchemaXORContent)
  .and(ParameterLocation);

const RequestBody = z.object({
  description: z.string().optional(),
  content: z.record(z.string(), MediaType),
  required: z.boolean().optional(),
});

export const Responses = z.record(z.string(), z.union([Response, Reference]));

const Operation = z.object({
  tags: z.array(z.string()).optional(),
  summary: z.string().optional(),
  description: z.string().optional(),
  externalDocs: ExternalDocumentation.optional(),
  operationId: z.string().optional(),
  parameters: z.array(z.union([Parameter, Reference])).optional(),
  requestBody: z.union([RequestBody, Reference]).optional(),
  responses: Responses,
  // TODO: callbacks has some recursive stuff going on
  // callbacks: z
  //   .record(z.string(), z.union([z.lazy(() => Callback), Reference]))
  //   .optional(),
  deprecated: z.boolean().optional(),
  security: z.array(z.record(z.array(z.string()))).optional(),
  servers: z.array(Server).optional(),
});

function nonEmptyResponses(operation: any, ctx: any) {
  if (!operation || !operation.responses) {
    return;
  }
  Object.entries(operation.responses).forEach(([status, res]) => {
    const response = res as z.infer<typeof Response>;
    // if response is a ref, we're done
    if ("$ref" in response) {
      return;
    }
    if (
      !response.headers &&
      (!response.content || Object.keys(response.content).length === 0)
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: `At least one of headers or content must be defined in a response: ${JSON.stringify(
          response,
          null,
          2
        )}`,
      });
    }
  });
}

export const PathItem = z.object({
  $ref: z.string().optional(),
  summary: z.string().optional(),
  description: z.string().optional(),
  servers: z.array(Server).optional(),
  parameters: z.array(z.union([Parameter, Reference])).optional(),
  get: Operation.optional().superRefine(nonEmptyResponses),
  put: Operation.optional().superRefine(nonEmptyResponses),
  post: Operation.optional().superRefine(nonEmptyResponses),
  delete: Operation.optional(), // allow empty responses for delete requests
  options: Operation.optional(),
  head: Operation.optional(),
  patch: Operation.optional(),
  trace: Operation.optional(),
});

export const Paths = z.record(
  z.string().regex(/^\/.+/).describe("Path must be a relative URL"),
  PathItem
);
export type Paths = z.infer<typeof Paths>;

const SecurityScheme = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("apiKey"),
    name: z.string(),
    in: z.enum(["header", "query", "cookie"]),
    description: z.string().optional(),
  }),
  z.object({
    scheme: z.string(),
    bearerFormat: z.string().optional(),
    description: z.string().optional(),
    type: z.literal("http"),
  }),
  z.object({
    type: z.literal("oauth2"),
    flows: z.object({
      implicit: z
        .object({
          authorizationUrl: z.string().url(),
          refreshUrl: z.string().url().optional(),
          scopes: z.record(z.string()),
        })
        .optional(),
      password: z
        .object({
          tokenUrl: z.string().url(),
          refreshUrl: z.string().url().optional(),
          scopes: z.record(z.string()),
        })
        .optional(),
      clientCredentials: z
        .object({
          tokenUrl: z.string().url(),
          refreshUrl: z.string().url().optional(),
          scopes: z.record(z.string()),
        })
        .optional(),
      authorizationCode: z
        .object({
          authorizationUrl: z.string().url(),
          tokenUrl: z.string().url(),
          refreshUrl: z.string().url().optional(),
          scopes: z.record(z.string()),
        })
        .optional(),
    }),
    description: z.string().optional(),
  }),
  z.object({
    type: z.literal("openIdConnect"),
    openIdConnectUrl: z.string().url(),
    description: z.string().optional(),
  }),
]);

export const Components = z
  .object({
    schemas: z.record(z.string(), z.union([Schema, Reference])).optional(),
    responses: z.record(z.string(), z.union([Response, Reference])).optional(),
    parameters: z
      .record(z.string(), z.union([Parameter, Reference]))
      .optional(),
    examples: z.record(z.string(), z.union([Example, Reference])).optional(),
    requestBodies: z
      .record(z.string(), z.union([RequestBody, Reference]))
      .optional(),
    headers: z.record(z.string(), z.union([Header, Reference])).optional(),
    securitySchemes: z
      .record(z.string(), z.union([SecurityScheme, Reference]))
      .optional(),
    links: z.record(z.string(), z.union([Link, Reference])).optional(),
    callbacks: z
      .record(z.string(), z.union([z.lazy(() => Callback), Reference]))
      .optional(),
  })
  .strict();

export type Components = z.infer<typeof Components>;

export function mergeComponents(...components: Components[]): Components {
  const merged: Components = {};
  for (const component of components) {
    merged.schemas = { ...merged.schemas, ...component.schemas };
    merged.responses = { ...merged.responses, ...component.responses };
    merged.parameters = { ...merged.parameters, ...component.parameters };
    merged.requestBodies = {
      ...merged.requestBodies,
      ...component.requestBodies,
    };
    merged.headers = { ...merged.headers, ...component.headers };
    merged.securitySchemes = {
      ...merged.securitySchemes,
      ...component.securitySchemes,
    };
    merged.links = { ...merged.links, ...component.links };
    merged.callbacks = { ...merged.callbacks, ...component.callbacks };
  }
  return merged;
}

const Callback = z.object({}).catchall(z.record(z.string(), PathItem));

const SecurityRequirement = z.record(z.string(), z.array(z.string()));

const Tag = z.object({
  name: z.string(),
  description: z.string().optional(),
  externalDocs: ExternalDocumentation.optional(),
});

export const OpenAPISchema = z.object({
  openapi: z.string().regex(/^3\.0\.\d(-.+)?$/),
  info: Info,
  externalDocs: ExternalDocumentation.optional(),
  servers: z.array(Server).optional(),
  security: SecurityRequirement.optional(),
  tags: z.array(Tag).optional(),
  paths: Paths,
  components: Components.optional(),
});

export default OpenAPISchema;
