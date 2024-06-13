import { FC } from "react";
import "./settings-page.module.css";

type OrganizationProfileLayoutProps = {
  children: React.ReactNode;
};

const OrganizationProfileLayout: FC<OrganizationProfileLayoutProps> = ({
  children,
}) => <div>{children}</div>;

export default OrganizationProfileLayout;
