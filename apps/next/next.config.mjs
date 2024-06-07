import TerserPlugin from 'terser-webpack-plugin';

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
  },
  webpack: (
    config,
    { buildId, dev, isServer, defaultLoaders, nextRuntime, webpack },
  ) => {
      if (isServer) {
          config.devtool = "eval-source-map"
          // https://github.com/open-telemetry/opentelemetry-js/issues/4173#issuecomment-1822938936
          config.ignoreWarnings = [
            { module: /opentelemetry/ },
          ]
      }
      config.optimization = {
        minimize: true,
      minimizer: [
        new TerserPlugin({
          terserOptions: {
            keep_fnames: true, // don't strip function names in production https://docs.temporal.io/dev-guide/typescript/debugging#works-in-dev-but-not-in-prod
          },
        }),
      ],
    }
    return config
  },

};

export default nextConfig;
