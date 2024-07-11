import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import "./globals.css";
import { ThemeProvider } from "@/providers/ThemeProvider";
import { Toaster } from "@/components/ui/toaster";
import { ClerkProvider } from "@clerk/nextjs";
import { dark } from "@clerk/themes";
import { CommandStoreProvider } from "@/providers/CommandStoreProvider";

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
      <html lang="en" suppressHydrationWarning>
        <body className={GeistSans.className}>
          <CommandStoreProvider>
            <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
              {children}
            </ThemeProvider>
          </CommandStoreProvider>
          <Toaster />
        </body>
      </html>
    </ClerkProvider>
  );
}
