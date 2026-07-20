import type { NextConfig } from "next";

// Static export when building for the Go embed (STATIC_EXPORT=1).
// Dev keeps a rewrite so /api hits the Go server on :8080.
const staticExport = process.env.STATIC_EXPORT === "1";

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
      destination: "http://127.0.0.1:8080/api/:path*",
    },
  ];
}

export default nextConfig;
