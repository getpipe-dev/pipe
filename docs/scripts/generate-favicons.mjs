import sharp from "sharp";
import { fileURLToPath } from "url";
import { dirname, join } from "path";
import { writeFileSync } from "fs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const publicDir = join(__dirname, "..", "public");

// Teal-branded favicon SVG matching hub site
const FAVICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36" fill="none">
  <rect width="36" height="36" rx="8" fill="#5eead4"/>
  <path d="M10 11h6v14h-3V14h-3v-3zm10 0h6v3h-3v11h-3V11z" fill="#0a0f1a"/>
</svg>`;

const sizes = [
  { name: "favicon-16x16.png", size: 16 },
  { name: "favicon-32x32.png", size: 32 },
  { name: "apple-touch-icon.png", size: 180 },
  { name: "android-chrome-192x192.png", size: 192 },
  { name: "android-chrome-512x512.png", size: 512 },
];

async function generateFavicons() {
  // Write the new teal SVG favicon
  writeFileSync(join(publicDir, "favicon.svg"), FAVICON_SVG.trim() + "\n");
  console.log("Generated: favicon.svg");

  // Generate PNG variants from SVG
  const svgBuffer = Buffer.from(FAVICON_SVG);

  for (const { name, size } of sizes) {
    await sharp(svgBuffer, { density: Math.max(72, Math.round((size / 36) * 72)) })
      .resize(size, size)
      .png()
      .toFile(join(publicDir, name));
    console.log(`Generated: ${name} (${size}x${size})`);
  }

  // Generate favicon.ico (32x32 PNG wrapped as ICO)
  const png32 = await sharp(svgBuffer, { density: 72 })
    .resize(32, 32)
    .png()
    .toBuffer();

  // ICO format: header + directory entry + PNG data
  const icoHeader = Buffer.alloc(6);
  icoHeader.writeUInt16LE(0, 0); // reserved
  icoHeader.writeUInt16LE(1, 2); // ICO type
  icoHeader.writeUInt16LE(1, 4); // 1 image

  const dirEntry = Buffer.alloc(16);
  dirEntry.writeUInt8(32, 0);  // width
  dirEntry.writeUInt8(32, 1);  // height
  dirEntry.writeUInt8(0, 2);   // color palette
  dirEntry.writeUInt8(0, 3);   // reserved
  dirEntry.writeUInt16LE(1, 4);  // color planes
  dirEntry.writeUInt16LE(32, 6); // bits per pixel
  dirEntry.writeUInt32LE(png32.length, 8);  // image size
  dirEntry.writeUInt32LE(22, 12); // offset (6 header + 16 dir entry)

  const ico = Buffer.concat([icoHeader, dirEntry, png32]);
  writeFileSync(join(publicDir, "favicon.ico"), ico);
  console.log("Generated: favicon.ico");
}

generateFavicons().catch((err) => {
  console.error(err);
  process.exit(1);
});
