import { components, paths } from "@/lib/hcloud";
import hetznerPricing from "@/lib/hcloud/pricing";
import createClient from "openapi-fetch";
type HetznerServer = components["schemas"]["server"];
type HetznerLoadBalancer = components["schemas"]["load_balancer"];
type HetznerVolume = components["schemas"]["volume"];
type HetznerPrimaryIp =
  components["schemas"]["get_primary_ip_response"]["primary_ip"];

const project = {
  hetznerApiToken:
    "SRqdiZKhmfm5PEVBzgBGxCaRuE92PfNNMFFBRPcD592EQmgULPFJa6M8szFqsGhx",
};
const cluster = {
  name: "quick-azure",
};

function serverHourlyPrice(type: string, location: string): number {
  const serverTypePricing = hetznerPricing.pricing.server_types.find(
    (server_type) => server_type.name === type
  );
  if (!serverTypePricing) {
    throw new Error(`unkown pricing for server type "${type}"`);
  }
  const serverTypeLocationPricing = serverTypePricing.prices.find(
    (price) => price.location === location
  );
  if (!serverTypeLocationPricing) {
    throw new Error(
      `unkown pricing for server type "${type}" in location "${location}"`
    );
  }
  return parseFloat(serverTypeLocationPricing.price_hourly.net);
}

function loadBalancerHourlyPrice(type: string, location: string): number {
  const typePricing = hetznerPricing.pricing.load_balancer_types.find(
    (lb_type) => lb_type.name === type
  );
  if (!typePricing) {
    throw new Error(`unkown pricing for load balancer type "${type}"`);
  }
  const typeLocationPricing = typePricing.prices.find(
    (price) => price.location === location
  );
  if (!typeLocationPricing) {
    throw new Error(
      `unkown pricing for load balancer type "${type}" in location "${location}"`
    );
  }
  return parseFloat(typeLocationPricing.price_hourly.net);
}

function volumeHourlyPrice(sizeGb: number): number {
  // hetzner only gives pricing for volumes on a monthly basis, so convert this to an hourly price that averages out
  // to the correct monthly price over the course of a year
  const daysPerMonth = 365.0 / 12.0;
  const hoursPerMonth = 24.0 * daysPerMonth;
  return (
    (parseFloat(hetznerPricing.pricing.volume.price_per_gb_month.net) *
      sizeGb) /
    hoursPerMonth
  );
}

function primaryIpHourlyPrice(type: string, location: string): number {
  const typePricing = hetznerPricing.pricing.primary_ips.find(
    (primaryIp) => primaryIp.type === type
  );
  if (!typePricing) {
    if (type === "ipv6") {
      return 0.0; // hetzner doesn't charge for primary ipv6 https://docs.hetzner.com/general/others/ipv4-pricing/#cloud
    }
    throw new Error(`unkown pricing for IP type "${type}"`);
  }
  const typeLocationPricing = typePricing.prices.find(
    (price) => price.location === location
  );
  if (!typeLocationPricing) {
    throw new Error(
      `unkown pricing for IP type "${type}" in location "${location}"`
    );
  }
  return parseFloat(typeLocationPricing.price_hourly.net);
}

async function main() {
  const client = createClient<paths>({
    headers: { Authorization: `Bearer ${project.hetznerApiToken}` },
    baseUrl: "https://api.hetzner.cloud/v1",
  });

  const servers: HetznerServer[] = [];
  for (let page = 1; ; page++) {
    const {
      error,
      response: _,
      data,
    } = await client.GET("/servers", {
      params: {
        query: {
          label_selector: `caph-cluster-${cluster.name}=owned`,
        },
        page,
        per_page: 100,
      },
    });
    if (error) {
      throw new Error(error);
    }
    if (data.servers.length > 0) {
      servers.push(...data.servers);
    }
    if (data.meta?.pagination.next_page === null) {
      break;
    }
  }

  console.log(`found ${servers.length} servers`);
  for (const server of servers) {
    console.log(
      `  type=${server.server_type.name} price_per_hour=${serverHourlyPrice(
        server.server_type.name,
        server.datacenter.location.name
      )}`
    );
  }

  const loadBalancers: HetznerLoadBalancer[] = [];
  for (let page = 1; ; page++) {
    const {
      error,
      response: _,
      data,
    } = await client.GET("/load_balancers", {
      params: {
        query: {
          label_selector: `caph-cluster-${cluster.name}=owned`,
        },
        page,
        per_page: 100,
      },
    });
    if (error) {
      throw new Error(error);
    }
    if (data.load_balancers.length > 0) {
      loadBalancers.push(...data.load_balancers);
    }
    if (data.meta?.pagination.next_page === null) {
      break;
    }
  }
  console.log(`found ${loadBalancers.length} load balancers`);
  for (const lb of loadBalancers) {
    console.log(
      `  type=${
        lb.load_balancer_type.name
      } price_per_hour=${loadBalancerHourlyPrice(
        lb.load_balancer_type.name,
        lb.location.name
      )}`
    );
  }

  // find all volumes attached to servers
  const volumes: HetznerVolume[] = [];
  const primaryIps: HetznerPrimaryIp[] = [];
  for (const server of servers) {
    const ipv4Id = server.public_net.ipv4?.id;
    if (ipv4Id !== null) {
      const {
        error,
        response: _,
        data,
      } = await client.GET("/primary_ips/{id}", {
        params: {
          path: {
            id: ipv4Id!,
          },
        },
      });
      if (error) {
        throw new Error(error);
      }
      primaryIps.push(data.primary_ip);
    }
    for (const volumeId of server.volumes || []) {
      const {
        error,
        response: _,
        data,
      } = await client.GET("/volumes/{id}", {
        params: {
          path: {
            id: volumeId,
          },
        },
      });
      if (error) {
        throw new Error(error);
      }
      volumes.push(data.volume);
    }
  }

  console.log(`found ${volumes.length} volumes`);
  for (const v of volumes) {
    console.log(`  size=${v.size} price_per_hour=${volumeHourlyPrice(v.size)}`);
  }

  console.log(`found ${primaryIps.length} primary IPs`);
  for (const ip of primaryIps) {
    console.log(
      `  ip=${ip.ip} type=${ip.type} labels=${JSON.stringify(
        ip.labels
      )} price_per_hour=${primaryIpHourlyPrice(
        ip.type,
        ip.datacenter.location.name
      )}`
    );
  }
}

export enum RoundingDirection {
  Up,
  Down,
}

export function roundToNearestHour(
  a: Date,
  direction: RoundingDirection
): Date {
  if (direction === RoundingDirection.Up) {
    a.setHours(a.getHours() + 1);
  }
  a.setMinutes(0);
  a.setSeconds(0);
  a.setMilliseconds(0);
  if (direction === RoundingDirection.Down) {
    a.setHours(a.getHours() - 1);
  }
  return a;
}

export function hoursBetween(a: Date, b: Date): number {
  return (b.getTime() - a.getTime()) / 1000 / 60 / 60;
}

main();
