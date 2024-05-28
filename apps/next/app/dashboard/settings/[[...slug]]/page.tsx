"use client";

import { OrganizationProfile } from "@clerk/nextjs";

const DotIcon = () => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 512 512"
      fill="currentColor"
    >
      <path d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512z" />
    </svg>
  );
};

const CustomPage = () => {
  return (
    <div>
      <h1>Custom Organization Profile Page</h1>
      <p>This is the custom organization profile page</p>
    </div>
  );
};

const OrganizationProfilePage = () => (
  <OrganizationProfile path="/dashboard/settings" routing="path">
    {/* You can pass the content as a component */}
    <OrganizationProfile.Page
      label="Custom Page"
      labelIcon={<DotIcon />}
      url="custom-page"
    >
      <CustomPage />
    </OrganizationProfile.Page>

    {/* You can also pass the content as direct children */}
    <OrganizationProfile.Page label="Terms" labelIcon={<DotIcon />} url="terms">
      <div>
        <h1>Custom Terms Page</h1>
        <p>This is the custom terms page</p>
      </div>
    </OrganizationProfile.Page>
  </OrganizationProfile>
);

export default OrganizationProfilePage;
