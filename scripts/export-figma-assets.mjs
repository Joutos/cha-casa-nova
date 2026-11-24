#!/usr/bin/env node
// Baixa do Figma os assets visuais originais para frontend/public/images:
//   hero.jpg   — foto do casal (preenchimento de imagem do nó 2:18, sem a película)
//   flower.svg — ilustração da flor (nó 3:42)
//   branch.svg — ilustração do ramo (nó 3:25, já rotacionado como no Figma)
//
// Uso:  FIGMA_TOKEN=figd_xxx node scripts/export-figma-assets.mjs
// O token pessoal é criado em Figma → Settings → Security → Personal access tokens
// (escopo "File content: read-only"). Ele não é salvo em lugar nenhum.

import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const FILE_KEY = process.env.FIGMA_FILE_KEY ?? "w93Fely3wvFygnUAvkAZxc";
const TOKEN = process.env.FIGMA_TOKEN;
const OUT = join(dirname(fileURLToPath(import.meta.url)), "..", "frontend", "public", "images");

if (!TOKEN) {
  console.error("Defina FIGMA_TOKEN com um token pessoal do Figma (somente leitura).");
  process.exit(1);
}

async function figma(path) {
  const res = await fetch(`https://api.figma.com/v1${path}`, { headers: { "X-Figma-Token": TOKEN } });
  if (!res.ok) throw new Error(`Figma ${path}: HTTP ${res.status} ${await res.text()}`);
  return res.json();
}

async function download(url, file) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`download ${file}: HTTP ${res.status}`);
  const buf = Buffer.from(await res.arrayBuffer());
  await writeFile(join(OUT, file), buf);
  console.log(`✓ ${file} (${Math.round(buf.length / 1024)} KB)`);
}

await mkdir(OUT, { recursive: true });

// 1) Foto: pega o imageRef do preenchimento do nó 2:18 e baixa a imagem original.
const nodes = await figma(`/files/${FILE_KEY}/nodes?ids=2:18`);
const fills = nodes.nodes["2:18"]?.document?.fills ?? [];
const imageRef = fills.find((f) => f.type === "IMAGE")?.imageRef;
if (!imageRef) throw new Error("Não encontrei a imagem no nó 2:18");
const images = await figma(`/files/${FILE_KEY}/images`);
const photoUrl = images.meta?.images?.[imageRef];
if (!photoUrl) throw new Error("URL da foto não retornada pelo Figma");
await download(photoUrl, "hero.jpg");

// 2) Ilustrações em SVG.
const svg = await figma(`/images/${FILE_KEY}?ids=3:42,3:25&format=svg&svg_include_id=false&svg_simplify_stroke=true`);
await download(svg.images["3:42"], "flower.svg");
await download(svg.images["3:25"], "branch.svg");

console.log("\nPronto. Rode o build do frontend (ou docker compose build) para incluir os arquivos.");
console.log("Dica: se a foto for grande, otimize-a (ex.: 1920px de largura, qualidade 80).");
