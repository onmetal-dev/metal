import type { Metadata } from "next";
import { Inter, Lexend } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/ThemeProvider";
import { Toaster } from "@/components/ui/toaster";
import { ClerkProvider } from "@clerk/nextjs";
import clsx from "clsx";
import { dark } from "@clerk/themes";
import { CommandItemsProvider, CommandMenu } from "@/components/CommandMenu";
import { ClerkRefresher } from "@/components/ClerkRefresher";

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
});
const lexend = Lexend({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-lexend",
});

export const metadata: Metadata = {
  title: "Metal",
  description: "Run your app on bare metal servers",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <ClerkProvider
      telemetry={false}
      signInFallbackRedirectUrl="/dashboard"
      signUpFallbackRedirectUrl="/dashboard"
      appearance={{
        baseTheme: dark,
        variables: {
          // keep these in sync with globals.css
          colorBackground: "hsl(222.2 84% 4.9%)",
          colorInputBackground: "hsl(222.2 84% 4.9%)",
          colorInputText: "hsl(210 40% 98%)",
          colorPrimary: "hsl(221.2 83.2% 53.3%)",
          colorTextOnPrimaryBackground: "hsl(222.2 47.4% 11.2%)",
        },
      }}
    >
      <html
        lang="en"
        className={clsx(
          "h-screen scroll-smooth bg-background antialiased",
          inter.variable,
          lexend.variable
        )}
        suppressHydrationWarning
      >
        <body className="h-full flex flex-col">
          <CommandItemsProvider>
            <CommandMenu />
            <ThemeProvider
              attribute="class"
              forcedTheme="dark"
              defaultTheme="dark"
              disableTransitionOnChange
            >
              <ClerkRefresher />
              {children}
            </ThemeProvider>
          </CommandItemsProvider>
          <Toaster />
        </body>
      </html>
    </ClerkProvider>
  );
}
