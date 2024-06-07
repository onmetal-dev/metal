import {
  Application,
  ApplicationConfig,
  BuildArtifact,
  ExternalPort,
  Port,
} from "@/app/server/db/schema";
import { ApiObject, Chart, JsonPatch } from "cdk8s";
import * as kplus from "cdk8s-plus-29";
import { Construct } from "constructs";
import { components } from "./gen/argoSchemas";
import * as k8sArgo from "./gen/argoproj.io";
import * as k8sGateway from "./gen/gateway.networking.k8s.io";

type schemas = components["schemas"];
export type Rollout = schemas["Rollout"];
type RolloutSpecTemplate = schemas["Rollout"]["spec"]["template"];

export interface ArgoDeploymentOptions {
  app: Application;
  appConfig: ApplicationConfig;
  buildArtifacts: BuildArtifact[];
  clusterName: string;
}

export class ArgoRollout extends Chart {
  constructor(
    scope: Construct,
    id: string,
    { app, appConfig, buildArtifacts, clusterName }: ArgoDeploymentOptions
  ) {
    const labels = { app: app.name };
    const namespace = app.name;
    super(scope, id, { labels, namespace });

    // all external ports must have a corresponding referent in appConfig.ports
    for (const ep of appConfig.external) {
      if (!appConfig.ports.find((p: Port) => p.name === ep.portName)) {
        throw new Error(
          `External port ${ep.name} not found in application config ports`
        );
      }
    }

    // identify the build artifact for this cluster
    const repository = `registry.${clusterName}.up.onmetal.dev`;
    const artifact: BuildArtifact | undefined = buildArtifacts.find(
      (a: BuildArtifact) => a.image?.repository === repository
    );
    if (!artifact || !artifact.image) {
      throw new Error(`No artifact found for ${repository}`);
    }

    // create stable and canary services
    const services = new Map<string, kplus.Service>();
    for (const x of ["stable", "canary"]) {
      let service = new kplus.Service(this, x, {
        type: kplus.ServiceType.CLUSTER_IP,
        ports: appConfig.ports.map((p: Port): kplus.ServicePort => {
          if (p.proto !== "http") {
            throw new Error("only http proto supported right now");
          }
          return {
            name: p.name,
            port: p.port,
            protocol: kplus.Protocol.TCP,
          };
        }),
      });
      // cdk8s Service requires an IPodSelector to fill out the selector field, but we don't have this since the argo rollout will create the pods
      // so we will break the glass a bit
      ApiObject.of(service).addJsonPatch(
        JsonPatch.add("/spec/selector", {
          app: labels.app,
        })
      );
      services.set(x, service);
    }
    const canaryService = services.get("canary")!;
    const stableService = services.get("stable")!;

    // create httproutes that will route external traffic to the services
    const httpRoutes = appConfig.external.map((external: ExternalPort) => {
      const appPort = appConfig.ports.find(
        (p: Port) => p.name === external.portName
      );
      if (!appPort) {
        throw new Error(
          `External port ${external.name} references port name ${external.portName} which does not exist in application config`
        );
      }
      return new k8sGateway.HttpRoute(
        this,
        `${external.name}-${external.port}-to-${external.portName}-${appPort.port}`,
        {
          spec: {
            parentRefs: [
              {
                kind: "Gateway",
                name: "cilium",
                namespace: "gateway",
                port: external.proto === "https" ? 443 : 80,
              },
            ],
            hostnames: [`${app.name}.${clusterName}.up.onmetal.dev`],
            rules: [
              {
                matches: [
                  {
                    path: {
                      type: k8sGateway.HttpRouteSpecRulesMatchesPathType
                        .PATH_PREFIX,
                      value: "/",
                    },
                  },
                ],
                backendRefs: [
                  {
                    name: stableService.name,
                    kind: stableService.kind,
                    port: stableService.port,
                  },
                  {
                    name: canaryService.name,
                    kind: canaryService.kind,
                    port: canaryService.port,
                  },
                ],
              },
            ],
          },
        }
      );
    });

    new k8sArgo.Rollout(this, "rollout", {
      spec: {
        replicas: 2,
        revisionHistoryLimit: 5,
        selector: {
          matchLabels: {
            app: app.name,
          },
        },
        template: {
          metadata: {
            labels,
          },
          spec: {
            imagePullSecrets: [
              {
                name: `regcred-${clusterName}`,
              },
            ],
            containers: [
              {
                name: app.name,
                image: `${artifact.image.repository}/${artifact.image.name}:${
                  artifact.image.tag ?? artifact.image.digest ?? "latest"
                }`,
                ports: appConfig.ports.map((p: Port) => ({
                  name: p.name,
                  containerPort: p.port,
                })),
                resources: {
                  requests: {
                    memory: appConfig.resources.memory,
                    cpu: appConfig.resources.cpu,
                  },
                },
                env: [
                  { name: "PORT", value: "80" },
                  {
                    name: "NODE_IP",
                    valueFrom: { fieldRef: { fieldPath: "status.hostIP" } },
                  },
                  {
                    name: "OTEL_EXPORTER_OTLP_ENDPOINT",
                    value: "http://$(NODE_IP):4317",
                  },
                ],
              },
            ],
          },
        },
        strategy: {
          canary: {
            canaryService: canaryService.name,
            stableService: stableService.name,
            trafficRouting: {
              plugins: {
                "argoproj-labs/gatewayAPI": {
                  namespace: this.namespace,
                  httpRoutes: httpRoutes.map((r) => ({
                    name: r.name,
                  })),
                },
              },
            },
            // TODO: let appconfig configure this
            steps: [{ setWeight: 100 }],
          },
        },
      },
    });
  }
}

export function parseRollout(rolloutJson: string): Rollout {
  return JSON.parse(rolloutJson) as Rollout;
}
