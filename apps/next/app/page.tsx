// import Image from "next/image";
import { Header } from "@/components/Header";
import { Hero } from "@/components/Hero";
import { Footer } from "@/components/Footer";
// import { PrimaryFeatures } from "@/components/PrimaryFeatures";
// import { SecondaryFeatures } from "@/components/SecondaryFeatures";
// import { CallToAction } from "@/components/CallToAction";
// import { Testimonials } from "@/components/Testimonials";
// import { Pricing } from "@/components/Pricing";
// import { Faqs } from "@/components/Faqs";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        {/* <PrimaryFeatures />
        <SecondaryFeatures />
        <CallToAction />
        <Testimonials />
        <Pricing />
        <Faqs /> */}
        {/* <Footer /> */}
      </main>
    </>
  );
}
