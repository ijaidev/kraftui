import type { NextConfig } from "next";

// Static export when building for the Go embed (STATIC_EXPORT=1).
// Dev rewrites /api to the Go server on KRAFTUI_PORT (default 5200).
const staticExport = process.env.STATIC_EXPORT === "1";
const apiPort = process.env.KRAFTUI_PORT || "5200";

const nextConfig: NextConfig = {
  images: {
    unoptimized: true,
  },
  reactCompiler: true,
};

if (staticExport) {
  nextConfig.output = "export";
} else {
  nextConfig.rewrites = async () => [
    {
      source: "/api/:path*",
      destination: `http://127.0.0.1:${apiPort}/api/:path*`,
    },
  ];
}

export default nextConfig;
