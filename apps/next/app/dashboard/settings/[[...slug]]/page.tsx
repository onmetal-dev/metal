"use client";

import { OrganizationProfile } from "@clerk/nextjs";
import { DollarSign, IceCreamCone } from "lucide-react";
import BillingPage from "./_profile-pages/billing-page";

const OrganizationProfilePage = () => (
  <OrganizationProfile path="/dashboard/settings" routing="path">
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
