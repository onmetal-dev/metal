export interface LogoProps extends React.ComponentPropsWithoutRef<"svg"> {
  width?: string;
  height?: string;
}

export function Logo(props: React.ComponentPropsWithoutRef<"svg">) {
  return (
    <svg
      width="512"
      height="512"
      viewBox="0 0 512 512"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <rect width="512" height="512" rx="76.8" fill="#2563EB" />
      <path
        d="M163.345 164.218H190.255L253.527 318.764H255.709L318.982 164.218H345.891V350.4H324.8V208.945H322.982L264.8 350.4H244.436L186.255 208.945H184.436V350.4H163.345V164.218Z"
        fill="#0F172A"
      />
      <path
        d="M61.4682 53.8546V100.4H53.0364V62.0591H52.7636L41.8773 69.0137V61.2864L53.4454 53.8546H61.4682Z"
        fill="#0F172A"
      />
    </svg>
  );
}
