import { FC } from "react";
import styles from "./settings-page.module.css";

type OrganizationProfileLayoutProps = {
  children: React.ReactNode;
};

const OrganizationProfileLayout: FC<OrganizationProfileLayoutProps> = ({
  children,
}) => <div className={styles.foo}>{children}</div>;

export default OrganizationProfileLayout;
