import {
  Application,
  ApplicationConfig,
  BuildArtifact,
} from "@/app/server/db/schema";
import { describe, expect, it } from "bun:test";
import { Testing } from "cdk8s";
import { readFileSync } from "fs";
import { join } from "path";
import {
  ArgoDeploymentOptions,
  ArgoRollout,
  Rollout,
  parseRollout,
} from "./rollout";

describe("parseRollout", () => {
  it("correctly parses JSON input", () => {
    const filePath = join(__dirname, "testdata", "rollout.json");
    const rolloutJson = readFileSync(filePath, "utf-8");
    const rollout: Rollout = parseRollout(rolloutJson);
    expect(rollout).toBeDefined();
    expect(rollout).toHaveProperty("spec");
    expect(rollout).toHaveProperty("status");
  });
});

describe("ArgoDeployment", () => {
  const app: Application = {
    name: "raf-test",
    id: "1",
    creatorId: "1",
    createdAt: new Date(),
    updatedAt: new Date(),
    teamId: "1",
  };

  const appConfig: ApplicationConfig = {
    id: "1",
    createdAt: new Date(),
    updatedAt: new Date(),
    teamId: "1",
    applicationId: "1",
    source: {
      type: "upload",
    },
    builder: {
      type: "nixpacks",
    },
    external: [
      {
        name: "https",
        portName: "http",
        port: 443,
        proto: "https",
      },
    ],
    ports: [
      {
        name: "http",
        port: 80,
        proto: "http",
      },
    ],
    resources: {
      memory: "256M",
      cpu: 0.1,
    },
    healthCheck: {
      proto: "http",
      portName: "http",
    },
    dependencies: [],
    databases: [],
    version: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  };

  const buildArtifacts: BuildArtifact[] = [
    {
      image: {
        repository: "registry.quick-azure.up.onmetal.dev",
        name: "raf-test",
        tag: "7ojbutasSzKYqzhupZdhPT",
      },
    },
  ];

  const clusterName = "quick-azure";

  const options: ArgoDeploymentOptions = {
    app,
    appConfig,
    buildArtifacts,
    clusterName,
  };

  it("should synthesize the correct Kubernetes manifest", () => {
    const testingApp = Testing.app();
    const chart = new ArgoRollout(testingApp, app.name, options);
    const manifest = Testing.synth(chart);

    expect(manifest).toMatchSnapshot();
  });

  it("should throw an error if external port is not found in appConfig.ports", () => {
    const invalidAppConfig: ApplicationConfig = {
      ...appConfig,
      external: [
        { name: "https", portName: "invalid-port", port: 443, proto: "https" },
      ],
    };

    const invalidOptions: ArgoDeploymentOptions = {
      ...options,
      appConfig: invalidAppConfig,
    };

    expect(() => {
      new ArgoRollout(Testing.app(), app.name, invalidOptions);
    }).toThrowError(
      "External port https not found in application config ports"
    );
  });

  it("should throw an error if no artifact is found for the repository", () => {
    const invalidBuildArtifacts: BuildArtifact[] = [];

    const invalidOptions = {
      ...options,
      buildArtifacts: invalidBuildArtifacts,
    };

    expect(() => {
      new ArgoRollout(Testing.app(), app.name, invalidOptions);
    }).toThrowError(
      `No artifact found for registry.${clusterName}.up.onmetal.dev`
    );
  });

  it("should construct httpRoutes array correctly", () => {
    const chart = new ArgoRollout(Testing.app(), app.name, options);
    const manifest = Testing.synth(chart);

    const rolloutSpec = manifest.find((item) => item.kind === "Rollout").spec
      .strategy.canary.trafficRouting.plugins["argoproj-labs/gatewayAPI"]
      .httpRoutes;

    expect(rolloutSpec).toEqual([
      { name: "raf-test-https-443-to-http-80-c89c9570" },
    ]);
  });

  it("should construct cpu and memory requests correctly", () => {
    const chart = new ArgoRollout(Testing.app(), app.name, options);
    const manifest = Testing.synth(chart);

    (manifest.find(
      (item) => item.kind === "Rollout"
    ) as Rollout)!.spec!.template!.spec!.containers.forEach((container) => {
      const requests: any = container!.resources!.requests!;
      expect(requests.cpu).toEqual(appConfig.resources.cpu);
      expect(requests.memory).toEqual(appConfig.resources.memory);
    });
  });
});
