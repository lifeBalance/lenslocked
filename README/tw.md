# Set Up Tailwind

The goal is to run: 

- `npm run watch:css` alongside `air` while **developing**
- `npm run build:css` before **deploying**.

## Install Tailwind via npm

From the repo root (same level as go.mod), initialise npm if you haven’t already:

```sh
npm init -y
```

Install Tailwind and the [Tailwind CLI](https://tailwindcss.com/docs/installation/tailwind-cli):

```sh
npm install -D tailwindcss @tailwindcss/cli
```

Add `node_modules` to `.gitignore` and also to `exclude_dir` in the `.air.toml` file.

## Create a tiny source stylesheet

Add something like `tailwind/styles.css` containing the three layer imports and any custom tweaks you need:

```css
@import "tailwindcss";

/* Your overrides can live here */
```

Tailwind v4 will auto‑discover your HTML/templ files; no `tailwind.config.js` required!

## Generate production-ready CSS

Choose an output location Go can serve, e.g. `public/assets/tailwind.css`. Add scripts to `package.json`:

```json
{
  "scripts": {
    "build:css": "tailwindcss -i ./tailwind/styles.css -o ./assets/styles.css --minify",
    "watch:css": "tailwindcss -i ./tailwind/styles.css -o ./assets/styles.css --watch"
  }
}
```

Running `npm run watch:css` during **development** mirrors what the CDN did, but with your own bundle.

## Wire it into your Go templates

In templates/tailwind.gohtml (or wherever you inject styles), replace the CDN `<script>` with:

```html
<link rel="stylesheet" href="/assets/styles.css">
```

## Serve Assets

In `cmd/server/server.go` we need to add an endpoint to serve static files such as our stylesheet, favicon, etc.

```go
assetsHandler := http.FileServer(http.Dir("assets"))
r.Get("/assets/*", http.StripPrefix("/assets", assetsHandler).ServeHTTP)
```

Ensure `/assets` is reachable — either via your existing **static handler** or by copying into whatever directory you embed/serve.

## Keep the output out of version control if desired

Add `public/assets/tailwind.css` to `.gitignore` (or only commit the built file if you’d rather avoid a build step in production).

> That’s it — no config file needed, and you get full control over **Tailwind v4** locally. 
