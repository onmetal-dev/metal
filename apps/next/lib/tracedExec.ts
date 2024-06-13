import { serviceName } from "@lib/constants";
import { SpanStatusCode, trace } from "@opentelemetry/api";
import { ExecException, exec as execCB } from "child_process";
import util from "util";
const exec = util.promisify(execCB);

// tracedExec is a function that runs a command and traces it
// it takes in the span name, attributes to set, and the command to run
// it returns true on success, false if the command fails
// it adds a lot of attributes in order to make debugging easier
export async function tracedExec({
  spanName,
  spanAttributes,
  command,
  directory,
}: {
  spanName: string;
  spanAttributes?: Record<string, string>;
  command: string;
  directory?: string;
}): Promise<{ stdout: string; stderr: string }> {
  return await trace
    .getTracer(serviceName)
    .startActiveSpan(spanName, async (span) => {
      if (spanAttributes) {
        span.setAttributes(spanAttributes);
      }
      try {
        console.log("DEBUG: exec", command);
        const ret = await exec(command, { cwd: directory });
        span.setAttributes(ret);
        span.end();
        return ret;
      } catch (e: any) {
        const error = e as ExecException;
        span.setAttributes({
          code: error.code,
          stdout: error.stdout,
          stderr: error.stderr,
          message: error.message,
        });
        span.setStatus({ code: SpanStatusCode.ERROR });
        span.end();
        throw new Error(`${spanName}: ${error.stdout} ${error.stderr}`);
      }
    });
}
