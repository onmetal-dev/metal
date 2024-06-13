"use client";

import { OrganizationProfile } from "@clerk/nextjs";
import { DollarSign, IceCreamCone } from "lucide-react";
import BillingPage from "./_profile-pages/billing-page";

const OrganizationProfilePage = () => (
  <OrganizationProfile
    path="/dashboard/settings"
    routing="path"
    appearance={{
      // The styles are used to make the settings box full-page-width. Comment them out and
      // see how the settings box only fills about 70% of the screen.
      elements: {
        rootBox: "w-full border border-slate-300 dark:border-slate-700",
        cardBox: "w-full grid-cols-8",
        navbar: "col-span-1",
        scrollBox: "col-span-7",
        headerTitle: "hidden",
        profileSection: "border-t-0 border-b",
      },
    }}
  >
    <OrganizationProfile.Page
      label="Billing"
      labelIcon={<DollarSign className="h-3.5 w-3.5" />}
      url="billing"
    >
      <BillingPage />
    </OrganizationProfile.Page>

    {/* You can also pass the content as direct children */}
    <OrganizationProfile.Page
      label="Custom"
      labelIcon={<IceCreamCone className="h-3.5 w-3.5" />}
      url="custom"
    >
      <div>
        <h1>Custom Organization Profile Page</h1>
        <p>This is a custom organization profile page</p>
      </div>
    </OrganizationProfile.Page>
  </OrganizationProfile>
);

export default OrganizationProfilePage;
