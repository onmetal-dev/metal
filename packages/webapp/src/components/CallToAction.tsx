import Image from "next/image";

import { Button } from "@/components/ui/button";
import { Container } from "@/components/Container";
import backgroundImage from "@/images/background-call-to-action.jpg";
import Link from "next/link";

export function CallToAction() {
  return (
    <section
      id="get-started-today"
      className="relative overflow-hidden bg-primary py-32"
    >
      <Image
        className="absolute left-1/2 top-1/2 max-w-none -translate-x-1/2 -translate-y-1/2"
        src={backgroundImage}
        alt=""
        width={2347}
        height={1244}
        unoptimized
      />
      <Container className="relative">
        <div className="mx-auto max-w-lg text-center">
          <h2 className="font-display text-3xl tracking-tight text-white sm:text-4xl">
            Get started today
          </h2>
          <p className="mt-4 text-lg tracking-tight text-primary-foreground">
            It’s time to take control of your infrastructure.
          </p>
          <Button
            asChild
            variant="secondary"
            className="font-semibold rounded-3xl mt-10"
          >
            <Link href="/register">Get 6 months free</Link>
          </Button>
        </div>
      </Container>
    </section>
  );
}
