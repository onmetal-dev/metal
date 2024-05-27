/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'img.clerk.com',
        port: '',
      },
    ],
  },
  typescript: {
    // speed up builds--will happen in gh action separate from build
    ignoreBuildErrors: true,
  },
  eslint: {
    // speed up builds--`next lint` will happen in gh actions
    ignoreDuringBuilds: true,
  },
  experimental: {
    // for opentelemetry tracing
    instrumentationHook: true,
  }
};

export default nextConfig;
