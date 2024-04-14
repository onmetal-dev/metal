"use client";
import { useTheme } from "next-themes";
import { Skeleton } from "./ui/skeleton";
import { useEffect, useState } from "react";

export function LogoWide(props: React.ComponentPropsWithoutRef<"svg">) {
  const [mounted, setMounted] = useState(false);
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return <Skeleton className="h-[40px] w-[109px] bg-background" />;
  }

  if (resolvedTheme === "dark") {
    return <LogoWideDark {...props} />;
  } else {
    return <LogoWideLight {...props} />;
  }
}

function LogoWideLight(props: React.ComponentPropsWithoutRef<"svg">) {
  return (
    <svg
      width="109"
      height="40"
      viewBox="0 0 109 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <rect width="40" height="40" rx="6" fill="#2563EB" />
      <path
        d="M12.7614 12.4545H14.8636L19.8068 24.5284H19.9773L24.9205 12.4545H27.0227V27H25.375V15.9489H25.233L20.6875 27H19.0966L14.5511 15.9489H14.4091V27H12.7614V12.4545Z"
        fill="#0F172A"
      />
      <path
        d="M4.8022 4.36364V8H4.14347V5.00462H4.12216L3.27166 5.54794V4.94425L4.17543 4.36364H4.8022Z"
        fill="#0F172A"
      />
      <path
        d="M52.7614 13.4545H54.8636L59.8068 25.5284H59.9773L64.9205 13.4545H67.0227V28H65.375V16.9489H65.233L60.6875 28H59.0966L54.5511 16.9489H54.4091V28H52.7614V13.4545ZM74.9009 28.2273C73.8498 28.2273 72.9431 27.9953 72.1808 27.5312C71.4232 27.0625 70.8384 26.4091 70.4265 25.571C70.0193 24.7282 69.8157 23.7481 69.8157 22.6307C69.8157 21.5133 70.0193 20.5284 70.4265 19.6761C70.8384 18.8191 71.4113 18.1515 72.1452 17.6733C72.8839 17.1903 73.7456 16.9489 74.7305 16.9489C75.2987 16.9489 75.8597 17.0436 76.4137 17.233C76.9677 17.4223 77.4719 17.7301 77.9265 18.1562C78.381 18.5777 78.7433 19.1364 79.0131 19.8324C79.283 20.5284 79.418 21.3854 79.418 22.4034V23.1136H71.0089V21.6648H77.7134C77.7134 21.0492 77.5903 20.5 77.3441 20.017C77.1026 19.5341 76.757 19.1529 76.3072 18.8736C75.8621 18.5942 75.3365 18.4545 74.7305 18.4545C74.0629 18.4545 73.4852 18.6203 72.9975 18.9517C72.5146 19.2784 72.1429 19.7045 71.8825 20.2301C71.622 20.7557 71.4918 21.3191 71.4918 21.9205V22.8864C71.4918 23.7102 71.6339 24.4086 71.918 24.9815C72.2068 25.5497 72.6069 25.983 73.1183 26.2812C73.6296 26.5748 74.2238 26.7216 74.9009 26.7216C75.3413 26.7216 75.739 26.66 76.0941 26.5369C76.454 26.4091 76.7641 26.2197 77.0245 25.9688C77.2849 25.7131 77.4862 25.3958 77.6282 25.017L79.2475 25.4716C79.0771 26.0208 78.7906 26.5038 78.3881 26.9205C77.9857 27.3324 77.4885 27.6544 76.8967 27.8864C76.3048 28.1136 75.6396 28.2273 74.9009 28.2273ZM86.712 17.0909V18.5114H81.0586V17.0909H86.712ZM82.7063 14.4773H84.3825V24.875C84.3825 25.3485 84.4511 25.7036 84.5884 25.9403C84.7305 26.1723 84.9104 26.3286 85.1282 26.4091C85.3507 26.4848 85.5851 26.5227 85.8313 26.5227C86.016 26.5227 86.1675 26.5133 86.2859 26.4943C86.4042 26.4706 86.4989 26.4517 86.57 26.4375L86.9109 27.9432C86.7972 27.9858 86.6386 28.0284 86.435 28.071C86.2314 28.1184 85.9734 28.142 85.6609 28.142C85.1874 28.142 84.7234 28.0402 84.2688 27.8366C83.819 27.633 83.445 27.3229 83.1467 26.9062C82.8531 26.4896 82.7063 25.964 82.7063 25.3295V14.4773ZM92.4435 28.2557C91.7522 28.2557 91.1249 28.1255 90.5614 27.8651C89.998 27.5999 89.5505 27.2188 89.2191 26.7216C88.8877 26.2197 88.7219 25.6136 88.7219 24.9034C88.7219 24.2784 88.8451 23.7718 89.0913 23.3835C89.3375 22.9905 89.6665 22.6828 90.0785 22.4602C90.4904 22.2377 90.945 22.072 91.4421 21.9631C91.944 21.8494 92.4483 21.7595 92.9549 21.6932C93.6178 21.608 94.1552 21.544 94.5671 21.5014C94.9838 21.4541 95.2868 21.3759 95.4762 21.267C95.6703 21.1581 95.7674 20.9687 95.7674 20.6989V20.642C95.7674 19.9413 95.5756 19.3968 95.1921 19.0085C94.8133 18.6203 94.238 18.4261 93.4663 18.4261C92.6661 18.4261 92.0387 18.6013 91.5842 18.9517C91.1296 19.3021 90.81 19.6761 90.6254 20.0739L89.0344 19.5057C89.3185 18.8428 89.6973 18.3267 90.1708 17.9574C90.649 17.5833 91.1699 17.3229 91.7333 17.1761C92.3015 17.0246 92.8602 16.9489 93.4094 16.9489C93.7598 16.9489 94.1623 16.9915 94.6168 17.0767C95.0761 17.1572 95.5188 17.3253 95.945 17.581C96.3758 17.8366 96.7333 18.2225 97.0174 18.7386C97.3015 19.2547 97.4435 19.946 97.4435 20.8125V28H95.7674V26.5227H95.6822C95.5685 26.7595 95.3791 27.0128 95.114 27.2827C94.8488 27.5526 94.4961 27.7822 94.0558 27.9716C93.6154 28.161 93.078 28.2557 92.4435 28.2557ZM92.6992 26.75C93.3621 26.75 93.9208 26.6198 94.3754 26.3594C94.8346 26.099 95.1803 25.7628 95.4123 25.3509C95.649 24.9389 95.7674 24.5057 95.7674 24.0511V22.517C95.6964 22.6023 95.5401 22.6804 95.2987 22.7514C95.0619 22.8177 94.7873 22.8769 94.4748 22.929C94.167 22.9763 93.8664 23.0189 93.5728 23.0568C93.284 23.09 93.0496 23.1184 92.8697 23.142C92.4341 23.1989 92.0269 23.2912 91.6481 23.419C91.274 23.5421 90.971 23.7292 90.739 23.9801C90.5117 24.2263 90.3981 24.5625 90.3981 24.9886C90.3981 25.571 90.6135 26.0114 91.0444 26.3097C91.48 26.6032 92.0316 26.75 92.6992 26.75ZM102.179 13.4545V28H100.503V13.4545H102.179Z"
        fill="#0F172A"
      />
    </svg>
  );
}

function LogoWideDark(props: React.ComponentPropsWithoutRef<"svg">) {
  return (
    <svg
      width="109"
      height="40"
      viewBox="0 0 109 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <rect width="40" height="40" rx="6" fill="#2563EB" />
      <path
        d="M12.7614 12.4545H14.8636L19.8068 24.5284H19.9773L24.9205 12.4545H27.0227V27H25.375V15.9489H25.233L20.6875 27H19.0966L14.5511 15.9489H14.4091V27H12.7614V12.4545Z"
        fill="#0F172A"
      />
      <path
        d="M4.8022 4.36364V8H4.14347V5.00462H4.12216L3.27166 5.54794V4.94425L4.17543 4.36364H4.8022Z"
        fill="#0F172A"
      />
      <path
        d="M52.7614 12.4545H54.8636L59.8068 24.5284H59.9773L64.9205 12.4545H67.0227V27H65.375V15.9489H65.233L60.6875 27H59.0966L54.5511 15.9489H54.4091V27H52.7614V12.4545ZM74.9009 27.2273C73.8498 27.2273 72.9431 26.9953 72.1808 26.5312C71.4232 26.0625 70.8384 25.4091 70.4265 24.571C70.0193 23.7282 69.8157 22.7481 69.8157 21.6307C69.8157 20.5133 70.0193 19.5284 70.4265 18.6761C70.8384 17.8191 71.4113 17.1515 72.1452 16.6733C72.8839 16.1903 73.7456 15.9489 74.7305 15.9489C75.2987 15.9489 75.8597 16.0436 76.4137 16.233C76.9677 16.4223 77.4719 16.7301 77.9265 17.1562C78.381 17.5777 78.7433 18.1364 79.0131 18.8324C79.283 19.5284 79.418 20.3854 79.418 21.4034V22.1136H71.0089V20.6648H77.7134C77.7134 20.0492 77.5903 19.5 77.3441 19.017C77.1026 18.5341 76.757 18.1529 76.3072 17.8736C75.8621 17.5942 75.3365 17.4545 74.7305 17.4545C74.0629 17.4545 73.4852 17.6203 72.9975 17.9517C72.5146 18.2784 72.1429 18.7045 71.8825 19.2301C71.622 19.7557 71.4918 20.3191 71.4918 20.9205V21.8864C71.4918 22.7102 71.6339 23.4086 71.918 23.9815C72.2068 24.5497 72.6069 24.983 73.1183 25.2812C73.6296 25.5748 74.2238 25.7216 74.9009 25.7216C75.3413 25.7216 75.739 25.66 76.0941 25.5369C76.454 25.4091 76.7641 25.2197 77.0245 24.9688C77.2849 24.7131 77.4862 24.3958 77.6282 24.017L79.2475 24.4716C79.0771 25.0208 78.7906 25.5038 78.3881 25.9205C77.9857 26.3324 77.4885 26.6544 76.8967 26.8864C76.3048 27.1136 75.6396 27.2273 74.9009 27.2273ZM86.712 16.0909V17.5114H81.0586V16.0909H86.712ZM82.7063 13.4773H84.3825V23.875C84.3825 24.3485 84.4511 24.7036 84.5884 24.9403C84.7305 25.1723 84.9104 25.3286 85.1282 25.4091C85.3507 25.4848 85.5851 25.5227 85.8313 25.5227C86.016 25.5227 86.1675 25.5133 86.2859 25.4943C86.4042 25.4706 86.4989 25.4517 86.57 25.4375L86.9109 26.9432C86.7972 26.9858 86.6386 27.0284 86.435 27.071C86.2314 27.1184 85.9734 27.142 85.6609 27.142C85.1874 27.142 84.7234 27.0402 84.2688 26.8366C83.819 26.633 83.445 26.3229 83.1467 25.9062C82.8531 25.4896 82.7063 24.964 82.7063 24.3295V13.4773ZM92.4435 27.2557C91.7522 27.2557 91.1249 27.1255 90.5614 26.8651C89.998 26.5999 89.5505 26.2188 89.2191 25.7216C88.8877 25.2197 88.7219 24.6136 88.7219 23.9034C88.7219 23.2784 88.8451 22.7718 89.0913 22.3835C89.3375 21.9905 89.6665 21.6828 90.0785 21.4602C90.4904 21.2377 90.945 21.072 91.4421 20.9631C91.944 20.8494 92.4483 20.7595 92.9549 20.6932C93.6178 20.608 94.1552 20.544 94.5671 20.5014C94.9838 20.4541 95.2868 20.3759 95.4762 20.267C95.6703 20.1581 95.7674 19.9687 95.7674 19.6989V19.642C95.7674 18.9413 95.5756 18.3968 95.1921 18.0085C94.8133 17.6203 94.238 17.4261 93.4663 17.4261C92.6661 17.4261 92.0387 17.6013 91.5842 17.9517C91.1296 18.3021 90.81 18.6761 90.6254 19.0739L89.0344 18.5057C89.3185 17.8428 89.6973 17.3267 90.1708 16.9574C90.649 16.5833 91.1699 16.3229 91.7333 16.1761C92.3015 16.0246 92.8602 15.9489 93.4094 15.9489C93.7598 15.9489 94.1623 15.9915 94.6168 16.0767C95.0761 16.1572 95.5188 16.3253 95.945 16.581C96.3758 16.8366 96.7333 17.2225 97.0174 17.7386C97.3015 18.2547 97.4435 18.946 97.4435 19.8125V27H95.7674V25.5227H95.6822C95.5685 25.7595 95.3791 26.0128 95.114 26.2827C94.8488 26.5526 94.4961 26.7822 94.0558 26.9716C93.6154 27.161 93.078 27.2557 92.4435 27.2557ZM92.6992 25.75C93.3621 25.75 93.9208 25.6198 94.3754 25.3594C94.8346 25.099 95.1803 24.7628 95.4123 24.3509C95.649 23.9389 95.7674 23.5057 95.7674 23.0511V21.517C95.6964 21.6023 95.5401 21.6804 95.2987 21.7514C95.0619 21.8177 94.7873 21.8769 94.4748 21.929C94.167 21.9763 93.8664 22.0189 93.5728 22.0568C93.284 22.09 93.0496 22.1184 92.8697 22.142C92.4341 22.1989 92.0269 22.2912 91.6481 22.419C91.274 22.5421 90.971 22.7292 90.739 22.9801C90.5117 23.2263 90.3981 23.5625 90.3981 23.9886C90.3981 24.571 90.6135 25.0114 91.0444 25.3097C91.48 25.6032 92.0316 25.75 92.6992 25.75ZM102.179 12.4545V27H100.503V12.4545H102.179Z"
        fill="#DBEAFE"
      />
    </svg>
  );
}
