import fs from "fs/promises";
import path from "path";

const nullByte = Buffer.from([0]);
const emptyImageBuffer = Buffer.from([
  0x52, 0x49, 0x46, 0x46, 0x24, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56,
  0x50, 0x38, 0x20, 0x18, 0x00, 0x00, 0x00, 0x30, 0x01, 0x00, 0x9d, 0x01, 0x2a,
  0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x34, 0x25, 0xa4, 0x00, 0x03, 0x70, 0x00,
  0xfe, 0xfb, 0xfd, 0x50, 0x00,
]);

const intfTypes = { NONE: 0, FILE: 1, BUFFER: 2 };

const constants = {
  TYPE_LOSSY: 0,
  TYPE_LOSSLESS: 1,
  TYPE_EXTENDED: 2,
};

const encodeResults = {
  LIB_NOT_READY: -1,
  LIB_INVALID_CONFIG: -2,
  SUCCESS: 0,
  VP8_ENC_ERROR_OUT_OF_MEMORY: 1,
  VP8_ENC_ERROR_BITSTREAM_OUT_OF_MEMORY: 2,
  VP8_ENC_ERROR_NULL_PARAMETER: 3,
  VP8_ENC_ERROR_INVALID_CONFIGURATION: 4,
  VP8_ENC_ERROR_BAD_DIMENSION: 5,
  VP8_ENC_ERROR_PARTITION0_OVERFLOW: 6,
  VP8_ENC_ERROR_PARTITION_OVERFLOW: 7,
  VP8_ENC_ERROR_BAD_WRITE: 8,
  VP8_ENC_ERROR_FILE_TOO_BIG: 9,
  VP8_ENC_ERROR_USER_ABORT: 10,
  VP8_ENC_ERROR_LAST: 11,
};

const imageHints = { DEFAULT: 0, PICTURE: 1, PHOTO: 2, GRAPH: 3 };
const imagePresets = {
  DEFAULT: 0,
  PICTURE: 1,
  PHOTO: 2,
  DRAWING: 3,
  ICON: 4,
  TEXT: 5,
};

function VP8Width(data) {
  return ((data[7] << 8) | data[6]) & 0b0011111111111111;
}
function VP8Height(data) {
  return ((data[9] << 8) | data[8]) & 0b0011111111111111;
}
function VP8LWidth(data) {
  return (((data[2] << 8) | data[1]) & 0b0011111111111111) + 1;
}
function VP8LHeight(data) {
  return (
    ((((data[4] << 16) | (data[3] << 8) | data[2]) >> 6) & 0b0011111111111111) +
    1
  );
}
function doesVP8LHaveAlpha(data) {
  return !!(data[4] & 0b00010000);
}

function createBasicChunk(name, data) {
  const header = Buffer.alloc(8);
  header.write(name, 0);
  header.writeUInt32LE(data.length, 4);
  if (data.length & 1)
    return { size: data.length + 9, chunks: [header, data, nullByte] };
  return { size: data.length + 8, chunks: [header, data] };
}

/**
 * riff/webp reader
 */
class WebPReader {
  constructor() {
    this.type = intfTypes.NONE;
  }

  async readFile(p) {
    this.type = intfTypes.FILE;
    this.fileBuf = await fs.readFile(p);
    this.path = p;
    this.cursor = 0;
  }

  readBuffer(buf) {
    this.type = intfTypes.BUFFER;
    this.buf = Buffer.isBuffer(buf) ? buf : Buffer.from(buf);
    this.cursor = 0;
  }

  async readBytes(n, mod) {
    if (this.type === intfTypes.FILE) {
      const end = Math.min(this.cursor + n, this.fileBuf.length);
      const b = this.fileBuf.subarray(this.cursor, end);
      this.cursor = end;
      return mod ? b : b.length === n ? b : undefined;
    }
    if (this.type === intfTypes.BUFFER) {
      const b = this.buf.subarray(this.cursor, this.cursor + n);
      this.cursor += n;
      return b;
    }
    throw new Error("reader not initialized");
  }

  async readFileHeader() {
    const buf = await this.readBytes(12);
    if (!buf) throw new Error("reached end while reading header");
    if (buf.toString("utf8", 0, 4) !== "RIFF")
      throw new Error("bad header (not RIFF)");
    if (buf.toString("utf8", 8, 12) !== "WEBP")
      throw new Error("bad header (not WEBP)");
    return { fileSize: buf.readUInt32LE(4) };
  }

  async readChunkHeader() {
    const buf = await this.readBytes(8, true);
    if (buf.length === 0) return { fourCC: "\x00\x00\x00\x00", size: 0 };
    if (buf.length < 8)
      throw new Error("reached end while reading chunk header");
    return { fourCC: buf.toString("utf8", 0, 4), size: buf.readUInt32LE(4) };
  }

  async readChunkContents(size) {
    const buf = await this.readBytes(size);
    if (size & 1) await this.readBytes(1);
    return buf;
  }

  async readChunk_raw(n, size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error(`reached end while reading ${n} chunk`);
    return { raw: buf };
  }

  async readChunk_VP8(size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error("reached end while reading VP8 chunk");
    return { raw: buf, width: VP8Width(buf), height: VP8Height(buf) };
  }

  async readChunk_VP8L(size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error("reached end while reading VP8L chunk");
    return {
      raw: buf,
      alpha: doesVP8LHaveAlpha(buf),
      width: VP8LWidth(buf),
      height: VP8LHeight(buf),
    };
  }

  async readChunk_VP8X(size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error("reached end while reading VP8X chunk");
    return {
      raw: buf,
      hasICCP: !!(buf[0] & 0b00100000),
      hasAlpha: !!(buf[0] & 0b00010000),
      hasEXIF: !!(buf[0] & 0b00001000),
      hasXMP: !!(buf[0] & 0b00000100),
      hasAnim: !!(buf[0] & 0b00000010),
      width: buf.readUIntLE(4, 3) + 1,
      height: buf.readUIntLE(7, 3) + 1,
    };
  }

  async readChunk_ANIM(size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error("reached end while reading ANIM chunk");
    return {
      raw: buf,
      bgColor: buf.subarray(0, 4),
      loops: buf.readUInt16LE(4),
    };
  }

  async readChunk_ANMF(size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error("reached end while reading ANMF chunk");
    const out = {
      raw: buf,
      x: buf.readUIntLE(0, 3),
      y: buf.readUIntLE(3, 3),
      width: buf.readUIntLE(6, 3) + 1,
      height: buf.readUIntLE(9, 3) + 1,
      delay: buf.readUIntLE(12, 3),
      blend: !(buf[15] & 0b00000010),
      dispose: !!(buf[15] & 0b00000001),
    };
    let keep = true;
    const anmfReader = new WebPReader();
    anmfReader.readBuffer(buf);
    anmfReader.cursor = 16;
    while (keep) {
      const header = await anmfReader.readChunkHeader();
      switch (header.fourCC) {
        case "VP8 ":
          if (!out.vp8) {
            out.type = constants.TYPE_LOSSY;
            out.vp8 = await anmfReader.readChunk_VP8(header.size);
            if (out.alph) out.vp8.alpha = true;
          }
          break;
        case "VP8L":
          if (!out.vp8l) {
            out.type = constants.TYPE_LOSSLESS;
            out.vp8l = await anmfReader.readChunk_VP8L(header.size);
          }
          break;
        case "ALPH":
          if (!out.alph) {
            out.alph = await anmfReader.readChunk_ALPH(header.size);
            if (out.vp8) out.vp8.alpha = true;
          }
          break;
        case "\x00\x00\x00\x00":
        default:
          keep = false;
          break;
      }
      if (anmfReader.cursor >= buf.length) break;
    }
    return out;
  }

  async readChunk_ALPH(size) {
    return this.readChunk_raw("ALPH", size);
  }
  async readChunk_ICCP(size) {
    return this.readChunk_raw("ICCP", size);
  }
  async readChunk_EXIF(size) {
    return this.readChunk_raw("EXIF", size);
  }
  async readChunk_XMP(size) {
    return this.readChunk_raw("XMP ", size);
  }

  async readChunk_skip(size) {
    const buf = await this.readChunkContents(size);
    if (!buf) throw new Error("reached end while skipping chunk");
  }

  async read() {
    let keep = true,
      first = true;
    const out = {};
    while (keep) {
      const { fourCC, size } = await this.readChunkHeader();
      switch (fourCC) {
        case "VP8 ":
          if (!out.vp8) {
            out.vp8 = await this.readChunk_VP8(size);
            if (out.alph) out.vp8.alpha = true;
            if (first) {
              out.type = constants.TYPE_LOSSY;
              keep = false;
            }
          } else await this.readChunk_skip(size);
          break;
        case "VP8L":
          if (!out.vp8l) {
            out.vp8l = await this.readChunk_VP8L(size);
            if (first) {
              out.type = constants.TYPE_LOSSLESS;
              keep = false;
            }
          } else await this.readChunk_skip(size);
          break;
        case "VP8X":
          if (!out.extended) {
            out.type = constants.TYPE_EXTENDED;
            out.extended = await this.readChunk_VP8X(size);
          } else await this.readChunk_skip(size);
          break;
        case "ANIM":
          if (!out.anim) {
            const { raw, bgColor, loops } = await this.readChunk_ANIM(size);
            out.anim = {
              bgColor: [bgColor[2], bgColor[1], bgColor[0], bgColor[3]],
              loops,
              frames: [],
              raw,
            };
          } else await this.readChunk_skip(size);
          break;
        case "ANMF":
          out.anim.frames.push(await this.readChunk_ANMF(size));
          break;
        case "ALPH":
          if (!out.alph) {
            out.alph = await this.readChunk_ALPH(size);
            if (out.vp8) out.vp8.alpha = true;
          } else await this.readChunk_skip(size);
          break;
        case "ICCP":
          if (!out.iccp) out.iccp = await this.readChunk_ICCP(size);
          else await this.readChunk_skip(size);
          break;
        case "EXIF":
          if (!out.exif) out.exif = await this.readChunk_EXIF(size);
          else await this.readChunk_skip(size);
          break;
        case "XMP ":
          if (!out.xmp) out.xmp = await this.readChunk_XMP(size);
          else await this.readChunk_skip(size);
          break;
        case "\x00\x00\x00\x00":
          keep = false;
          break;
        default:
          await this.readChunk_skip(size);
          break;
      }
      first = false;
    }
    return out;
  }
}

/**
 * riff/webp writer
 */
class WebPWriter {
  constructor() {
    this.type = intfTypes.NONE;
    this.chunks = [];
    this.width = 0;
    this.height = 0;
    this.vp8x = null;
  }

  reset() {
    this.chunks.length = 0;
    this.width = 0;
    this.height = 0;
  }
  writeFile(p) {
    this.type = intfTypes.FILE;
    this.path = p;
  }
  writeBuffer() {
    this.type = intfTypes.BUFFER;
  }

  async commit() {
    if (this.type === intfTypes.NONE) throw new Error("writer not initialized");
    if (this.chunks.length === 0) throw new Error("nothing to write");
    let size = 4;
    for (let i = 1; i < this.chunks.length; i++) size += this.chunks[i].length;
    this.chunks[0].writeUInt32LE(size, 4);
    const out = Buffer.concat(this.chunks);
    if (this.type === intfTypes.FILE) {
      await fs.writeFile(this.path, out);
      return;
    }
    return out;
  }

  writeBytes(...chunks) {
    if (this.type === intfTypes.NONE) throw new Error("writer not initialized");
    this.chunks.push(...chunks);
  }

  writeFileHeader() {
    const buf = Buffer.alloc(12);
    buf.write("RIFF", 0);
    buf.write("WEBP", 8);
    this.writeBytes(buf);
  }

  writeChunk_VP8(vp8) {
    this.writeBytes(...createBasicChunk("VP8 ", vp8.raw).chunks);
  }
  writeChunk_VP8L(vp8l) {
    this.writeBytes(...createBasicChunk("VP8L", vp8l.raw).chunks);
  }

  writeChunk_VP8X(vp8x) {
    const buf = Buffer.alloc(18);
    buf.write("VP8X", 0);
    buf.writeUInt32LE(10, 4);
    buf.writeUIntLE(vp8x.width - 1, 12, 3);
    buf.writeUIntLE(vp8x.height - 1, 15, 3);
    if (vp8x.hasICCP) buf[8] |= 0b00100000;
    if (vp8x.hasAlpha) buf[8] |= 0b00010000;
    if (vp8x.hasEXIF) buf[8] |= 0b00001000;
    if (vp8x.hasXMP) buf[8] |= 0b00000100;
    if (vp8x.hasAnim) buf[8] |= 0b00000010;
    this.vp8x = buf;
    this.writeBytes(buf);
  }

  updateChunk_VP8X_size(width, height) {
    this.vp8x.writeUIntLE(width - 1, 12, 3);
    this.vp8x.writeUIntLE(height - 1, 15, 3);
  }

  writeChunk_ANIM(anim) {
    const buf = Buffer.alloc(14);
    buf.write("ANIM", 0);
    buf.writeUInt32LE(6, 4);
    buf.writeUInt8(anim.bgColor[2], 8);
    buf.writeUInt8(anim.bgColor[1], 9);
    buf.writeUInt8(anim.bgColor[0], 10);
    buf.writeUInt8(anim.bgColor[3], 11);
    buf.writeUInt16LE(anim.loops, 12);
    this.writeBytes(buf);
  }

  writeChunk_ANMF(anmf) {
    const buf = Buffer.alloc(24);
    const { img } = anmf;
    let size = 16;
    let alpha = false;
    buf.write("ANMF", 0);
    buf.writeUIntLE(anmf.x, 8, 3);
    buf.writeUIntLE(anmf.y, 11, 3);
    buf.writeUIntLE(anmf.delay, 20, 3);
    if (!anmf.blend) buf[23] |= 0b00000010;
    if (anmf.dispose) buf[23] |= 0b00000001;

    switch (img.type) {
      case constants.TYPE_LOSSY: {
        let b;
        this.width = Math.max(this.width, img.vp8.width + anmf.x);
        this.height = Math.max(this.height, img.vp8.height + anmf.y);
        buf.writeUIntLE(img.vp8.width - 1, 14, 3);
        buf.writeUIntLE(img.vp8.height - 1, 17, 3);
        this.writeBytes(buf);
        if (img.vp8.alpha) {
          b = createBasicChunk("ALPH", img.alph.raw);
          this.writeBytes(...b.chunks);
          size += b.size;
        }
        b = createBasicChunk("VP8 ", img.vp8.raw);
        this.writeBytes(...b.chunks);
        size += b.size;
        break;
      }
      case constants.TYPE_LOSSLESS: {
        const b = createBasicChunk("VP8L", img.vp8l.raw);
        this.width = Math.max(this.width, img.vp8l.width + anmf.x);
        this.height = Math.max(this.height, img.vp8l.height + anmf.y);
        buf.writeUIntLE(img.vp8l.width - 1, 14, 3);
        buf.writeUIntLE(img.vp8l.height - 1, 17, 3);
        if (img.vp8l.alpha) alpha = true;
        this.writeBytes(buf, ...b.chunks);
        size += b.size;
        break;
      }
      case constants.TYPE_EXTENDED: {
        if (img.extended.hasAnim) {
          const fr = img.anim.frames;
          if (img.extended.hasAlpha) alpha = true;
          for (let i = 0; i < fr.length; i++) {
            const b = Buffer.alloc(8);
            const c = fr[i].raw;
            this.width = Math.max(this.width, fr[i].width + anmf.x);
            this.height = Math.max(this.height, fr[i].height + anmf.y);
            b.write("ANMF", 0);
            b.writeUInt32LE(c.length, 4);
            c.writeUIntLE(anmf.x, 0, 3);
            c.writeUIntLE(anmf.y, 3, 3);
            c.writeUIntLE(anmf.delay, 12, 3);
            if (!anmf.blend) c[15] |= 0b00000010;
            else c[15] &= 0b11111101;
            if (anmf.dispose) c[15] |= 0b00000001;
            else c[15] &= 0b11111110;
            this.writeBytes(b, c);
            if (c.length & 1) this.writeBytes(nullByte);
          }
        } else {
          let b;
          this.width = Math.max(this.width, img.extended.width + anmf.x);
          this.height = Math.max(this.height, img.extended.height + anmf.y);
          if (img.vp8) {
            buf.writeUIntLE(img.vp8.width - 1, 14, 3);
            buf.writeUIntLE(img.vp8.height - 1, 17, 3);
            this.writeBytes(buf);
            if (img.alph) {
              b = createBasicChunk("ALPH", img.alph.raw);
              alpha = true;
              this.writeBytes(...b.chunks);
              size += b.size;
            }
            b = createBasicChunk("VP8 ", img.vp8.raw);
            this.writeBytes(...b.chunks);
            size += b.size;
          } else if (img.vp8l) {
            buf.writeUIntLE(img.vp8l.width - 1, 14, 3);
            buf.writeUIntLE(img.vp8l.height - 1, 17, 3);
            if (img.vp8l.alpha) alpha = true;
            b = createBasicChunk("VP8L", img.vp8l.raw);
            this.writeBytes(buf, ...b.chunks);
            size += b.size;
          }
        }
        break;
      }
      default:
        throw new Error("unknown image type");
    }
    buf.writeUInt32LE(size, 4);
    if (alpha && this.vp8x) this.vp8x[8] |= 0b00010000;
  }

  writeChunk_ALPH(alph) {
    this.writeBytes(...createBasicChunk("ALPH", alph.raw).chunks);
  }
  writeChunk_ICCP(iccp) {
    this.writeBytes(...createBasicChunk("ICCP", iccp.raw).chunks);
  }
  writeChunk_EXIF(exif) {
    this.writeBytes(...createBasicChunk("EXIF", exif.raw).chunks);
  }
  writeChunk_XMP(xmp) {
    this.writeBytes(...createBasicChunk("XMP ", xmp.raw).chunks);
  }
}

/**
 * high-level image api
 */
class Image {
  constructor() {
    this.data = null;
    this.loaded = false;
    this.path = "";
  }

  clear() {
    this.data = null;
    this.path = "";
    this.loaded = false;
  }

  get width() {
    const d = this.data;
    return !this.loaded
      ? undefined
      : d.extended
        ? d.extended.width
        : d.vp8l
          ? d.vp8l.width
          : d.vp8
            ? d.vp8.width
            : undefined;
  }
  get height() {
    const d = this.data;
    return !this.loaded
      ? undefined
      : d.extended
        ? d.extended.height
        : d.vp8l
          ? d.vp8l.height
          : d.vp8
            ? d.vp8.height
            : undefined;
  }
  get type() {
    return this.loaded ? this.data.type : undefined;
  }
  get hasAnim() {
    return this.loaded
      ? this.data.extended
        ? this.data.extended.hasAnim
        : false
      : false;
  }
  get hasAlpha() {
    return this.loaded
      ? this.data.extended
        ? this.data.extended.hasAlpha
        : this.data.vp8
          ? this.data.vp8.alpha
          : this.data.vp8l
            ? this.data.vp8l.alpha
            : false
      : false;
  }
  get anim() {
    return this.hasAnim ? this.data.anim : undefined;
  }
  get frames() {
    return this.anim ? this.anim.frames : undefined;
  }

  get iccp() {
    return this.data.extended
      ? this.data.extended.hasICCP
        ? this.data.iccp.raw
        : undefined
      : undefined;
  }
  set iccp(raw) {
    if (!this.data.extended) this._convertToExtended();
    if (raw === undefined) {
      this.data.extended.hasICCP = false;
      delete this.data.iccp;
    } else {
      this.data.iccp = { raw };
      this.data.extended.hasICCP = true;
    }
  }

  get exif() {
    return this.data.extended
      ? this.data.extended.hasEXIF
        ? this.data.exif.raw
        : undefined
      : undefined;
  }
  set exif(raw) {
    if (!this.data.extended) this._convertToExtended();
    if (raw === undefined) {
      this.data.extended.hasEXIF = false;
      delete this.data.exif;
    } else {
      this.data.exif = { raw };
      this.data.extended.hasEXIF = true;
    }
  }

  get xmp() {
    return this.data.extended
      ? this.data.extended.hasXMP
        ? this.data.xmp.raw
        : undefined
      : undefined;
  }
  set xmp(raw) {
    if (!this.data.extended) this._convertToExtended();
    if (raw === undefined) {
      this.data.extended.hasXMP = false;
      delete this.data.xmp;
    } else {
      this.data.xmp = { raw };
      this.data.extended.hasXMP = true;
    }
  }

  _convertToExtended() {
    if (!this.loaded) throw new Error("no image loaded");
    this.data.type = constants.TYPE_EXTENDED;
    this.data.extended = {
      hasICCP: false,
      hasAlpha: false,
      hasEXIF: false,
      hasXMP: false,
      hasAnim: this.data.extended?.hasAnim || false,
      width: this.data.vp8
        ? this.data.vp8.width
        : this.data.vp8l
          ? this.data.vp8l.width
          : 1,
      height: this.data.vp8
        ? this.data.vp8.height
        : this.data.vp8l
          ? this.data.vp8l.height
          : 1,
    };
  }

  async _demuxFrame(d, frame) {
    const { hasICCP, hasEXIF, hasXMP } = this.data.extended
      ? this.data.extended
      : { hasICCP: false, hasEXIF: false, hasXMP: false };
    const hasAlpha = frame.vp8 && frame.vp8.alpha;
    const writer = new WebPWriter();
    if (typeof d === "string") writer.writeFile(d);
    else writer.writeBuffer();
    writer.writeFileHeader();
    if (hasICCP || hasEXIF || hasXMP || hasAlpha) {
      writer.writeChunk_VP8X({
        hasICCP,
        hasEXIF,
        hasXMP,
        hasAlpha: (frame.vp8l && frame.vp8l.alpha) || hasAlpha,
        width: frame.width,
        height: frame.height,
      });
    }
    if (frame.vp8l) writer.writeChunk_VP8L(frame.vp8l);
    else if (frame.vp8) {
      if (frame.vp8.alpha) writer.writeChunk_ALPH(frame.alph);
      writer.writeChunk_VP8(frame.vp8);
    } else throw new Error("frame has no VP8/VP8L");
    if (hasICCP || hasEXIF || hasXMP || hasAlpha) {
      if (this.data.extended.hasICCP) writer.writeChunk_ICCP(this.data.iccp);
      if (this.data.extended.hasEXIF) writer.writeChunk_EXIF(this.data.exif);
      if (this.data.extended.hasXMP) writer.writeChunk_XMP(this.data.xmp);
    }
    return writer.commit();
  }

  async _save(
    writer,
    {
      width = undefined,
      height = undefined,
      frames = undefined,
      bgColor = [255, 255, 255, 255],
      loops = 0,
      delay = 100,
      x = 0,
      y = 0,
      blend = true,
      dispose = false,
      exif = false,
      iccp = false,
      xmp = false,
    } = {},
  ) {
    const _width = width !== undefined ? width : this.width - 1;
    const _height = height !== undefined ? height : this.height - 1;
    const isAnim = this.hasAnim || frames !== undefined;

    if (_width < 0 || _width > 1 << 24) throw new Error("width out of range");
    if (_height < 0 || _height > 1 << 24)
      throw new Error("height out of range");
    if (_height * _width > Math.pow(2, 32) - 1)
      throw new Error(`width * height too large (${_width}, ${_height})`);

    if (isAnim) {
      if (loops < 0 || loops >= 1 << 24) throw new Error("loops out of range");
      if (delay < 0 || delay >= 1 << 24) throw new Error("delay out of range");
      if (x < 0 || x >= 1 << 24) throw new Error("x out of range");
      if (y < 0 || y >= 1 << 24) throw new Error("y out of range");
    } else {
      if (_width === 0 || _height === 0)
        throw new Error("width/height cannot be 0");
    }

    writer.writeFileHeader();

    switch (this.type) {
      case constants.TYPE_LOSSY:
        writer.writeChunk_VP8(this.data.vp8);
        break;
      case constants.TYPE_LOSSLESS:
        writer.writeChunk_VP8L(this.data.vp8l);
        break;
      case constants.TYPE_EXTENDED: {
        const hasICCP = iccp === true ? !!this.iccp : iccp;
        const hasEXIF = exif === true ? !!this.exif : exif;
        const hasXMP = xmp === true ? !!this.xmp : xmp;
        writer.writeChunk_VP8X({
          hasICCP,
          hasEXIF,
          hasXMP,
          hasAlpha: this.data.alph || (this.data.vp8l && this.data.vp8l.alpha),
          hasAnim: isAnim,
          width: _width,
          height: _height,
        });
        if (hasICCP)
          writer.writeChunk_ICCP(iccp !== true ? iccp : this.data.iccp);
        if (isAnim) {
          const _frames = frames || this.frames;
          writer.writeChunk_ANIM({ bgColor, loops });
          for (let i = 0; i < _frames.length; i++) {
            const fr = _frames[i];
            const _delay = fr.delay == undefined ? delay : fr.delay;
            const _x = fr.x == undefined ? x : fr.x;
            const _y = fr.y == undefined ? y : fr.y;
            const _blend = fr.blend == undefined ? blend : fr.blend;
            const _dispose = fr.dispose == undefined ? dispose : fr.dispose;
            let img;
            if (_delay < 0 || _delay >= 1 << 24)
              throw new Error(`delay out of range on frame ${i}`);
            if (_x < 0 || _x >= 1 << 24)
              throw new Error(`x out of range on frame ${i}`);
            if (_y < 0 || _y >= 1 << 24)
              throw new Error(`y out of range on frame ${i}`);
            if (fr.path) {
              img = new Image();
              await img.load(fr.path);
              img = img.data;
            } else if (fr.buffer) {
              img = new Image();
              await img.load(fr.buffer);
              img = img.data;
            } else if (fr.img) {
              img = fr.img.data;
            } else {
              img = fr;
            }
            writer.writeChunk_ANMF({
              x: _x,
              y: _y,
              delay: _delay,
              blend: _blend,
              dispose: _dispose,
              img,
            });
          }
          if (_width === 0 || _height === 0)
            writer.updateChunk_VP8X_size(
              _width === 0 ? writer.width : _width,
              _height === 0 ? writer.height : _height,
            );
        } else {
          if (this.data.vp8) {
            if (this.data.alph) writer.writeChunk_ALPH(this.data.alph);
            writer.writeChunk_VP8(this.data.vp8);
          } else if (this.data.vp8l) writer.writeChunk_VP8L(this.data.vp8l);
        }
        if (hasEXIF)
          writer.writeChunk_EXIF(exif !== true ? exif : this.data.exif);
        if (hasXMP) writer.writeChunk_XMP(xmp !== true ? xmp : this.data.xmp);
        break;
      }
      default:
        throw new Error("unknown image type");
    }

    return writer.commit();
  }

  async load(d) {
    const reader = new WebPReader();
    if (typeof d === "string") {
      await reader.readFile(d);
      this.path = d;
    } else reader.readBuffer(d);
    this.data = await reader.read();
    this.loaded = true;
  }

  convertToAnim() {
    if (!this.data.extended) this._convertToExtended();
    if (this.hasAnim) return;
    if (this.data.vp8) delete this.data.vp8;
    if (this.data.vp8l) delete this.data.vp8l;
    if (this.data.alph) delete this.data.alph;
    this.data.extended.hasAnim = true;
    this.data.anim = { bgColor: [255, 255, 255, 255], loops: 0, frames: [] };
  }

  async demux({
    path: outPath = undefined,
    buffers = false,
    frame = -1,
    prefix = "#FNAME#",
    start = 0,
    end = 0,
  } = {}) {
    if (!this.hasAnim) throw new Error("this image isn't an animation");
    let _end = end === 0 ? this.frames.length : end;
    const bufs = [];
    if (start < 0) start = 0;
    if (_end >= this.frames.length) _end = this.frames.length - 1;
    if (start > _end) {
      const n = start;
      start = _end;
      _end = n;
    }
    if (frame !== -1) {
      start = _end = frame;
    }
    for (let i = start; i <= _end; i++) {
      const fname = this.path ? path.basename(this.path, ".webp") : "frame";
      const t = await this._demuxFrame(
        outPath
          ? `${outPath}/${prefix}_${i}.webp`.replace(/#FNAME#/g, fname)
          : undefined,
        this.anim.frames[i],
      );
      if (buffers) bufs.push(t);
    }
    if (buffers) return bufs;
  }

  async replaceFrame(frameIndex, d) {
    if (!this.hasAnim) throw new Error("webp isn't animated");
    if (typeof frameIndex !== "number")
      throw new Error("frame index expects a number");
    if (frameIndex < 0 || frameIndex >= this.frames.length)
      throw new Error(
        `frame index out of bounds (0 <= index < ${this.frames.length})`,
      );
    const r = new Image();
    const fr = this.frames[frameIndex];
    await r.load(d);
    switch (r.type) {
      case constants.TYPE_LOSSY:
      case constants.TYPE_LOSSLESS:
        break;
      case constants.TYPE_EXTENDED:
        if (r.hasAnim)
          throw new Error("merging animations not currently supported");
        break;
      default:
        throw new Error("unknown webp type");
    }
    switch (fr.type) {
      case constants.TYPE_LOSSY:
        if (fr.vp8?.alpha) delete fr.alph;
        delete fr.vp8;
        break;
      case constants.TYPE_LOSSLESS:
        delete fr.vp8l;
        break;
      default:
        throw new Error("unknown frame type");
    }
    switch (r.type) {
      case constants.TYPE_LOSSY:
        fr.vp8 = r.data.vp8;
        fr.type = constants.TYPE_LOSSY;
        break;
      case constants.TYPE_LOSSLESS:
        fr.vp8l = r.data.vp8l;
        fr.type = constants.TYPE_LOSSLESS;
        break;
      case constants.TYPE_EXTENDED:
        if (r.data.vp8) {
          fr.vp8 = r.data.vp8;
          if (r.data.vp8.alpha) fr.alph = r.data.alph;
          fr.type = constants.TYPE_LOSSY;
        } else if (r.data.vp8l) {
          fr.vp8l = r.data.vp8l;
          fr.type = constants.TYPE_LOSSLESS;
        }
        break;
    }
    fr.width = r.width;
    fr.height = r.height;
  }

  async save(
    p = this.path,
    {
      width = this.width,
      height = this.height,
      frames = this.frames,
      bgColor = this.hasAnim ? this.anim.bgColor : [255, 255, 255, 255],
      loops = this.hasAnim ? this.anim.loops : 0,
      delay = 100,
      x = 0,
      y = 0,
      blend = true,
      dispose = false,
      exif = !!this.exif,
      iccp = !!this.iccp,
      xmp = !!this.xmp,
    } = {},
  ) {
    const writer = new WebPWriter();
    if (p !== null) writer.writeFile(p);
    else writer.writeBuffer();
    return this._save(writer, {
      width,
      height,
      frames,
      bgColor,
      loops,
      delay,
      x,
      y,
      blend,
      dispose,
      exif,
      iccp,
      xmp,
    });
  }

  async getImageData() {
    if (!Image.libwebp)
      throw new Error("must call Image.initLib() before using getImageData");
    if (this.hasAnim)
      throw new Error("calling getImageData on animations is not supported");
    const buf = await this.save(null);
    return Image.libwebp.decodeImage(buf, this.width, this.height);
  }

  static async save(d, opts) {
    if (opts.frames && (opts.width === undefined || opts.height === undefined))
      throw new Error("must provide both width and height when passing frames");
    return (await Image.getEmptyImage(!!opts.frames)).save(d, opts);
  }

  static async getEmptyImage(ext) {
    const img = new Image();
    await img.load(emptyImageBuffer);
    if (ext) img.exif = undefined;
    return img;
  }

  static async generateFrame({
    path: p = undefined,
    buffer = undefined,
    img = undefined,
    x = undefined,
    y = undefined,
    delay = undefined,
    blend = undefined,
    dispose = undefined,
  } = {}) {
    let _img = img;
    if ((!p && !buffer && !img) || (p && buffer && img))
      throw new Error("must provide either `path`, `buffer`, or `img`");
    if (!img) {
      _img = new Image();
      if (p) await _img.load(p);
      else await _img.load(buffer);
    }
    if (_img.hasAnim)
      throw new Error("merging animations is not currently supported");
    return { img: _img, x, y, delay, blend, dispose };
  }

  static from(webp) {
    const img = new Image();
    img.data = webp.data;
    img.loaded = webp.loaded;
    img.path = webp.path;
    return img;
  }
}

export {
  constants,
  encodeResults,
  imageHints as hints,
  imagePresets as presets,
  WebPReader,
  WebPWriter,
  Image,
};

export const TYPE_LOSSY = constants.TYPE_LOSSY;
export const TYPE_LOSSLESS = constants.TYPE_LOSSLESS;
export const TYPE_EXTENDED = constants.TYPE_EXTENDED;
