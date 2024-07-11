import Link from "next/link";

export function Footer() {
  return (
    <div className="z-20 w-full bg-background/95 shadow backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="mx-4 md:mx-8 flex h-14 items-center">
        <p className="text-xs md:text-sm leading-loose text-muted-foreground text-left">
          Made with <code className="mx-0.5">❤️</code> in Cincinnati, OH. The
          source code is available on{" "}
          <Link
            href="https://github.com/onmetal-dev/metal"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium underline underline-offset-4"
          >
            GitHub
          </Link>
          . Join us on{" "}
          <Link
            href="https://discord.gg/onmetal"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium underline underline-offset-4"
          >
            Discord
          </Link>
          .
        </p>
      </div>
    </div>
  );
}
