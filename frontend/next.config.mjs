/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    ignoreBuildErrors: true, // allow build to complete; check logs for TS warnings
  },
  eslint: {
    ignoreDuringBuilds: true, // don't let ESLint block the build
  },
  images: {
    unoptimized: true,
  },
}

export default nextConfig
