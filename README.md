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

---

## 🛠️ Quick Start

### 1. Installation

Install the native standalone `now` CLI binary:

```bash
curl -sSL https://nowts.dev/install.sh | sh

```

*(Alternatively, initialize via npm wrapper: `npx create-nowts-app my-app`)*

### 2. Initialize Project

Create a new application and configure your deployment target (SSH or cPanel):

```bash
now init my-app
cd my-app

```

### 3. Scaffold & Develop

```bash
# Start local development server with on-the-fly asset optimization
now serve

# Generate a new component
now g c navbar

```

### 4. Ship to Production

Compile TypeScript via Go, build Vite chunks, optimize images, and deploy:

```bash
now run

```

---

## 📁 Project Structure

```text
my-app/
├── .now/
│   └── deploy.config.json    # Saved target credentials (SSH/cPanel)
├── public/                    # Raw assets (auto-optimized on build)
│   └── logo.png
├── src/
│   ├── app/
│   │   ├── core/              # Services, Guards, and Interceptors
│   │   ├── features/          # Feature components
│   │   └── app.component.ts
│   └── main.ts
├── now.config.ts              # Framework & Vite settings
├── tsconfig.json
└── package.json

```

---

## 📋 CLI Command Reference

| Command | Alias | Description |
| --- | --- | --- |
| `now init [name]` | — | Interactively creates a project and configures deployment targets. |
| `now serve` | `now dev` | Launches the Vite dev server with image optimization middleware. |
| `now run` | `now build` | Runs Go type-checking, builds chunks, and deploys to target. |
| `now g c <name>` | `now generate component` | Scaffolds a component (`.ts`, `.html`, `.css`). |
| `now g s <name>` | `now generate service` | Scaffolds a `@Injectable` service. |

---

## 🌐 Supported Deployment Targets

### 1. Linux VPS (Caddy / Nginx)

* Automatically provisions directory structures (`releases/` and `current/`).
* Executes zero-downtime atomic symlink switching.
* Native Caddy HTTPS integration.

### 2. cPanel / Shared Hosting

* Direct deployment via SFTP or WebDAV to `public_html/`.
* Auto-generates optimized `.htaccess` fallback rules for SPA routing.
* Leverages existing cPanel AutoSSL without root privileges.

---

## 🤝 Contributing

Contributions are what make the open-source community an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Check out our [CONTRIBUTING.md](https://www.google.com/search?q=CONTRIBUTING.md) for local development setup and guidelines.

---

## 📄 License

Distributed under the MIT License. See [`LICENSE`](https://www.google.com/search?q=LICENSE) for more information.
