/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: true,
    experimental: { appDir: false },
    async rewrites() {
        const api = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
        return [
            { source: '/api/:path*', destination: `${api}/api/:path*` }
        ]
    }
}

module.exports = nextConfig
