/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    ignoreBuildErrors: false, // surface TS errors in build log
  },
  eslint: {
    ignoreDuringBuilds: true, // don't let ESLint block the build
  },
  images: {
    unoptimized: true,
  },
}

export default nextConfig
