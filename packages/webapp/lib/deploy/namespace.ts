import { Chart, ChartProps } from "cdk8s";
import * as k8s from "cdk8s-plus-29/lib/imports/k8s";
import { Construct } from "constructs";

export class Namespace extends Chart {
  constructor(scope: Construct, id: string, props: ChartProps) {
    super(scope, id, props);
    new k8s.KubeNamespace(this, "namespace", {
      metadata: {
        name: props.namespace,
      },
    });
  }
}
