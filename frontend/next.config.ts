import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  ...(isDev
    ? {
        async rewrites() {
          return [
            {
              source: "/api/:path*",
              destination: "http://127.0.0.1:8085/api/:path*",
            },
            {
              source: "/v1/:path*",
              destination: "http://127.0.0.1:8085/v1/:path*",
            },
          ];
        },
      }
    : {
        output: "export",
      }),
};

export default nextConfig;
