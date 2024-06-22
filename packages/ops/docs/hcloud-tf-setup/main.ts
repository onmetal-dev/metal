import { Construct } from "constructs";
import { App, TerraformOutput, TerraformStack } from "cdktf";
import { DataHcloudImage } from "./.gen/providers/hcloud/data-hcloud-image";
import { HcloudProvider } from "./.gen/providers/hcloud/provider";
import {
  provider as packerProvider,
  dataPackerFiles,
  image as packerImage,
} from "./.gen/providers/packer";
import {
  provider as tlsProvider,
  privateKey as tlsPrivateKey,
} from "@cdktf/provider-tls";
import * as path from "path";
import { TerraformHcloudKubeHetzner } from "./.gen/modules/terraform-hcloud-kube-hetzner";
import createClient from "openapi-fetch";
import type { paths } from "./hetzner/hetzner";

const hcloudToken = process.env.HCLOUD_TOKEN!; // todo: pull from db with user config

type ImagesSuccessResponse =
  paths["/images"]["get"]["responses"][200]["content"]["application/json"]["images"];

async function getImageData(): Promise<ImagesSuccessResponse> {
  const client = createClient<paths>({
    baseUrl: "https://api.hetzner.cloud/v1",
    headers: {
      Authorization: `Bearer ${hcloudToken}`,
    },
  });
  const { data, error } = await client.GET("/images", {
    params: {},
  });
  if (error) {
    throw new Error("Failed to fetch images data");
  }
  return data.images;
}

async function main() {
  const app = new App();
  const imageData = await getImageData();
  new PackerImageStack(app, "packer-images", imageData);
  new KubeHetznerSatck(app, "hcloud-kube-hetzner");
  app.synth();
}

main().catch((error) => console.error(error));

class PackerImageStack extends TerraformStack {
  constructor(scope: Construct, id: string, images: ImagesSuccessResponse) {
    super(scope, id);
    const microosArmSnapshotId: number | undefined = images.find(
      (image) => image.name === "OpenSUSE MicroOS ARM by Kube-Hetzner"
    )?.id;
    const microosX86SnapshotId: number | undefined = images.find(
      (image) => image.name === "OpenSUSE MicroOS x86 by Kube-Hetzner"
    )?.id;
    if (!microosArmSnapshotId || !microosX86SnapshotId) {
      new packerProvider.PackerProvider(this, "packer", {});
      const packerFileX86 = new dataPackerFiles.DataPackerFiles(
        this,
        "packerfileX86",
        {
          file: path.join(process.cwd(), "hcloud-microos-snapshot-x86.pkr.hcl"),
        }
      );
      const packerFileArm = new dataPackerFiles.DataPackerFiles(
        this,
        "packerfileArm",
        {
          file: path.join(process.cwd(), "hcloud-microos-snapshot-arm.pkr.hcl"),
        }
      );
      new packerImage.Image(this, "microos-x86", {
        name: "opensuse/microos",
        force: true, // overwrite existing image if it exists
        environment: {
          HCLOUD_TOKEN: hcloudToken,
        },
        directory: process.cwd(),
        file: "hcloud-microos-snapshot-x86.pkr.hcl",
        triggers: {
          fileHash: packerFileX86.filesHash,
        },
      });
      new packerImage.Image(this, "microos-arm", {
        name: "opensuse/microos",
        force: true, // overwrite existing image if it exists
        environment: {
          HCLOUD_TOKEN: hcloudToken,
        },
        directory: process.cwd(),
        file: "hcloud-microos-snapshot-arm.pkr.hcl",
        triggers: {
          fileHash: packerFileArm.filesHash,
        },
      });
    }
  }
}

class KubeHetznerSatck extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    // create a keypair to use for ssh into the servers
    // TODO: save these for the user to download?
    new tlsProvider.TlsProvider(this, "tls", {});
    const pk = new tlsPrivateKey.PrivateKey(this, "ssh-key", {
      algorithm: "ED25519",
    });
    new TerraformOutput(this, "ssh_private_key", {
      value: pk.privateKeyOpenssh,
      sensitive: true,
    });
    new TerraformOutput(this, "ssh_public_key", {
      value: pk.publicKeyOpenssh,
    });

    const hcloud = new HcloudProvider(this, "hcloud", {
      token: hcloudToken,
    });
    const x86Image = new DataHcloudImage(this, "x86-image", {
      withSelector: "creator=kube-hetzner",
      withArchitecture: "x86",
    });
    const armImage = new DataHcloudImage(this, "arm-image", {
      withSelector: "creator=kube-hetzner",
      withArchitecture: "arm",
    });
    const kube = new TerraformHcloudKubeHetzner(this, "hcloudkube", {
      hcloudToken,
      sshPrivateKey: pk.privateKeyOpenssh,
      sshPublicKey: pk.publicKeyOpenssh,
      providers: [hcloud],
      microosArmSnapshotId: armImage.id.toString(),
      microosX86SnapshotId: x86Image.id.toString(),
      loadBalancerLocation: "ash",
      networkRegion: "us-east",
      agentNodepools: [
        //https://docs.hetzner.com/cloud/servers/overview/
        {
          name: "agent-small",
          server_type: "ccx13",
          location: "ash",
          labels: [],
          taints: [],
          count: 1,
          placement_group: "default",
        },
        {
          name: "egress",
          server_type: "ccx13",
          location: "ash",
          labels: ["node.kubernetes.io/role=egress"],
          taints: ["node.kubernetes.io/role=egress:NoSchedule"],
          floating_ip: true,
          count: 1,
          placement_group: "default",
        },
      ],
      controlPlaneNodepools: [
        {
          name: "control-plane-ash",
          server_type: "ccx13",
          location: "ash",
          labels: [],
          taints: [],
          count: 1,
          placement_group: "default",
        },
      ],
    });
    new TerraformOutput(this, "agents_public_ipv4", {
      value: kube.agentsPublicIpv4Output,
    });
    new TerraformOutput(this, "cluster_name", {
      value: kube.clusterNameOutput,
    });
    new TerraformOutput(this, "control_planes_public_ipv4", {
      value: kube.controlPlanesPublicIpv4Output,
    });
    new TerraformOutput(this, "ingress_public_ipv4", {
      value: kube.ingressPublicIpv4Output,
    });
    new TerraformOutput(this, "ingress_public_ipv6", {
      value: kube.ingressPublicIpv6Output,
    });
    new TerraformOutput(this, "k3s_endpoint", {
      value: kube.k3SEndpointOutput,
    });
    new TerraformOutput(this, "k3s_token", {
      value: kube.k3STokenOutput,
      sensitive: true,
    });
    new TerraformOutput(this, "kubeconfig", {
      value: kube.kubeconfigOutput,
      sensitive: true,
    });
    new TerraformOutput(this, "kubeconfig_data", {
      value: kube.kubeconfigDataOutput,
      sensitive: true,
    });
    new TerraformOutput(this, "kubeconfig_file", {
      value: kube.kubeconfigFileOutput,
      sensitive: true,
    });
    new TerraformOutput(this, "network_id", { value: kube.networkIdOutput });
    new TerraformOutput(this, "ssh_key_id", { value: kube.sshKeyIdOutput });
  }
}
