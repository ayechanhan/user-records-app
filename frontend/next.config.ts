import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // @tanstack/react-table ships ESM-only; transpiling it here is what lets
  // next/jest transform it too (Jest otherwise treats all of node_modules
  // as CommonJS and chokes on the bare `import`).
  transpilePackages: ["@tanstack/react-table", "@tanstack/table-core"],
};

export default nextConfig;
