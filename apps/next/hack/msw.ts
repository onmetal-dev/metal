import type { paths } from "@lib/hcloud";
import { HttpResponse } from "msw";
import { setupServer } from "msw/node";
import createClient from "openapi-fetch";
import { createOpenApiHttp } from "openapi-msw";

const http = createOpenApiHttp<paths>({
  baseUrl: "https://fake.hetzner.cloud/v1",
});

const handlers = [
  http.get("/servers", () => {
    return HttpResponse.json({
      servers: [
        {
          id: 1,
          name: "test",
          backup_window: "123",
          created: "2023-01-01T00:00:00Z",
          datacenter: {
            description: "Test Datacenter",
            id: 1,
            location: {
              city: "Test City",
              country: "Test Country",
              description: "Test Location",
              id: 1,
              latitude: 0,
              longitude: 0,
              name: "Test Location",
              network_zone: "Test Zone",
            },
            name: "Test Datacenter",
            server_types: {
              supported: [],
              available: [],
              available_for_migration: [],
            },
          },
          image: {
            id: 1,
            name: "test-image",
            architecture: "x86",
            bound_to: null,
            created: "2023-01-01T00:00:00Z",
            created_from: { id: 1, name: "base-image" },
            deleted: null,
            deprecated: null,
            description: "A test image",
            disk_size: 10,
            image_size: 10,
            labels: {},
            os_flavor: "ubuntu",
            os_version: "20.04",
            protection: { delete: false },
            rapid_deploy: false,
            status: "available",
            type: "system",
          },
          included_traffic: 0,
          ingoing_traffic: 0,
          iso: null,
          labels: {},
          load_balancers: [],
          locked: false,
          outgoing_traffic: 0,
          placement_group: null,
          primary_disk_size: 0,
          private_net: [],
          protection: {
            delete: false,
            rebuild: false,
          },
          public_net: {
            floating_ips: [],
            ipv4: {
              blocked: false,
              dns_ptr: "test.dns",
              id: 1,
              ip: "0.0.0.0",
            },
            ipv6: {
              blocked: false,
              dns_ptr: [],
              ip: "0:0:0:0:0:0:0:0",
            },
          },
          rescue_enabled: false,
          server_type: {
            cores: 1,
            cpu_type: "dedicated",
            deprecated: false,
            description: "A test server type",
            disk: 10,
            id: 1,
            memory: 10,
            name: "test-server-type",
            prices: [],
            storage_type: "local",
          },
          status: "running",
          volumes: [],
        },
      ],
    });
  }),
];

export const server = setupServer(...handlers);

server.listen({
  onUnhandledRequest: "warn",
});

async function main() {
  const client = createClient<paths>({
    headers: { Authorization: `Bearer DONTMATTER` },
    baseUrl: "https://fake.hetzner.cloud/v1",
  });
  const { error, response: _, data } = await client.GET("/servers");
  if (error) {
    console.error(error);
  }
  console.log(data);
}

main();
