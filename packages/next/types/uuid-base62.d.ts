declare module "uuid-base62" {
  export function encode(uuid: string): string;
  export function decode(str: string): string;
  export function v4(): string;
}
