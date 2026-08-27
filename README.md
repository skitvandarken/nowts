# nowts
The deploy-first JS/TS framework powered by Vite and Go. Instant type-checks, automatic public/ asset compression, and zero-config SSH/cPanel shipping.


# ⚡ nowTS

> **The deploy-first JavaScript/TypeScript framework.**
> Ship production-ready web applications directly to any Linux VPS, SSH server, or cPanel hosting in seconds — powered by Vite, Go-based TypeScript compilation, and native asset optimization.

---

## 💡 Why nowTS?

Modern web development often traps developers between expensive SaaS hosting lock-ins and complex CI/CD pipeline setups. **nowTS** eliminates that friction. It combines Angular-inspired CLI ergonomics with Vite bundling, lightning-fast Go type-checking, and zero-touch remote provisioning.

Whether deploying to a $2/month cPanel shared host or a minimal 512MB RAM Linux VPS, `nowTS` turns deployment into a single, instantaneous step.

---

## ✨ Key Features

* 🚀 **Deploy-First Workflow:** Configure SSH/SFTP or cPanel hosting on initialization (`now init`). Ship updates instantly with zero-downtime atomic symlink swaps.
* ⚡ **Go-Powered Speed:** Native integration with the Go-rewritten TypeScript compiler (`tsgo`) for near-instant type checks without Node V8 boot overhead.
* 🖼️ **Native Asset Optimization:** Automatically compresses and converts `public/` assets (PNG, JPEG) into WebP/AVIF during development (`now serve`) and production (`now run`).
* 📦 **Angular-Inspired CLI Ergonomics:** Generate structured components, services, and guards effortlessly using clean commands like `now g c <name>`.
* 🌐 **Infrastructure Freedom:** Zero vendor lock-in. Native support for Caddy, Nginx, and Apache/LiteSpeed (`.htaccess` auto-generation).
