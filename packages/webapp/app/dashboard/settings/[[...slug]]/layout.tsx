import { FC } from "react";

type OrganizationProfileLayoutProps = {
  children: React.ReactNode;
};

const OrganizationProfileLayout: FC<OrganizationProfileLayoutProps> = ({
  children,
}) => <div>{children}</div>;

export default OrganizationProfileLayout;
