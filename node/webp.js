#!/usr/bin/env node

const fs = require("fs");
const { tmpdir } = require("os");
const Crypto = require("crypto");
const webp = require("node-webpmux");
const path = require("path");
const { exec } = require("child_process");
const { promisify } = require("util");

const execAsync = promisify(exec);

function fileToBuffer(filePath) {
  if (!fs.existsSync(filePath)) {
    throw new Error(`File not found: ${filePath}`);
  }
  return fs.readFileSync(filePath);
}

async function imageToWebp(buffer) {
  const tmpInput = path.join(tmpdir(), `input_${Date.now()}.tmp`);
  const tmpOutput = path.join(tmpdir(), `output_${Date.now()}.webp`);

  try {
    fs.writeFileSync(tmpInput, buffer);
    await execAsync(`ffmpeg -i "${tmpInput}" -vcodec libwebp -filter:v fps=fps=15 -lossless 1 -loop 0 -preset default -an -vsync 0 "${tmpOutput}" -y`);
    const result = fs.readFileSync(tmpOutput);

    fs.unlinkSync(tmpInput);
    fs.unlinkSync(tmpOutput);

    return result;
  } catch (error) {
    if (fs.existsSync(tmpInput)) fs.unlinkSync(tmpInput);
    if (fs.existsSync(tmpOutput)) fs.unlinkSync(tmpOutput);
    throw new Error(`Image conversion failed: ${error.message}`);
  }
}

async function videoToWebp(buffer) {
  const tmpInput = path.join(tmpdir(), `input_${Date.now()}.tmp`);
  const tmpOutput = path.join(tmpdir(), `output_${Date.now()}.webp`);

  try {
    fs.writeFileSync(tmpInput, buffer);
    await execAsync(`ffmpeg -i "${tmpInput}" -vcodec libwebp -filter:v fps=fps=15,scale=320:320:flags=lanczos:force_original_aspect_ratio=decrease -loop 0 -preset default -an -vsync 0 -s 512:512 "${tmpOutput}" -y`);
    const result = fs.readFileSync(tmpOutput);

    fs.unlinkSync(tmpInput);
    fs.unlinkSync(tmpOutput);

    return result;
  } catch (error) {
    if (fs.existsSync(tmpInput)) fs.unlinkSync(tmpInput);
    if (fs.existsSync(tmpOutput)) fs.unlinkSync(tmpOutput);
    throw new Error(`Video conversion failed: ${error.message}`);
  }
}

async function processToWebp(mediaBuffer, metadata, mediaType) {
  let processedMedia;

  switch (mediaType.toLowerCase()) {
    case 'image':
      processedMedia = await imageToWebp(mediaBuffer);
      break;
    case 'video':
      processedMedia = await videoToWebp(mediaBuffer);
      break;
    case 'webp':
    default:
      processedMedia = mediaBuffer;
      break;
  }

  if (metadata.author || metadata.packname) {
    const tmpFileIn = path.join(tmpdir(), `${Crypto.randomBytes(6).readUIntLE(0, 6).toString(36)}.webp`);

    fs.writeFileSync(tmpFileIn, processedMedia);

    const img = new webp.Image();
    const json = {
      "sticker-pack-id": "https://github.com/AstroX11/user-bot",
      "sticker-pack-name": metadata.packname || "",
      "sticker-pack-publisher": metadata.author || "",
      emojis: metadata.categories || [""],
    };

    const exifAttr = Buffer.from([
      0x49, 0x49, 0x2a, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x41, 0x57,
      0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00,
    ]);
    const jsonBuff = Buffer.from(JSON.stringify(json), "utf-8");
    const exif = Buffer.concat([exifAttr, jsonBuff]);
    exif.writeUIntLE(jsonBuff.length, 14, 4);

    await img.load(tmpFileIn);
    img.exif = exif;

    const tmpFileOut = path.join(tmpdir(), `${Crypto.randomBytes(6).readUIntLE(0, 6).toString(36)}.webp`);
    await img.save(tmpFileOut);

    fs.unlinkSync(tmpFileIn);
    return fs.readFileSync(tmpFileOut);
  }

  return processedMedia;
}

function parseArgs() {
  const args = process.argv.slice(2);

  if (args.length < 2) {
    console.error('Usage: node webp.js <mediaType> <inputFile> [author <authorName>] [packname <packName>] [categories <cat1,cat2>]');
    process.exit(1);
  }

  const mediaType = args[0];
  const inputFile = args[1];

  let author = '';
  let packname = '';
  let categories = [];

  for (let i = 2; i < args.length; i += 2) {
    const key = args[i];
    const value = args[i + 1];

    switch (key.toLowerCase()) {
      case 'author':
        author = value;
        break;
      case 'packname':
        packname = value;
        break;
      case 'categories':
        categories = value.split(',').map(cat => cat.trim());
        break;
    }
  }

  return {
    mediaType,
    inputFile,
    metadata: {
      author,
      packname,
      categories: categories.length > 0 ? categories : undefined
    }
  };
}

async function main() {
  try {
    const { mediaType, inputFile, metadata } = parseArgs();

    // Read input file
    const mediaBuffer = fileToBuffer(inputFile);

    // Process to WebP with metadata
    const webpBuffer = await processToWebp(mediaBuffer, metadata, mediaType);

    // Write to output file
    const baseName = path.basename(inputFile, path.extname(inputFile));
    const outputPath = `${baseName}_sticker.webp`;

    fs.writeFileSync(outputPath, webpBuffer);

    console.log(`WebP created: ${outputPath}`);

  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}