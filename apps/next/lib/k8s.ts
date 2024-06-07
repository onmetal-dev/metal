import { ExecException, exec as execCB } from "child_process";
import util from "util";
const exec = util.promisify(execCB);

export async function findOrCreateNamespace(
  kubeconfigFilename: string,
  namespace: string
): Promise<void> {
  try {
    await exec(
      `kubectl --kubeconfig=${kubeconfigFilename} get namespace ${namespace}`
    );
  } catch (e: any) {
    const error = e as ExecException;
    if (error.code === 1) {
      // Namespace not found, create it
      await exec(
        `kubectl --kubeconfig=${kubeconfigFilename} create namespace ${namespace}`
      );
    } else {
      throw e;
    }
  }
}
