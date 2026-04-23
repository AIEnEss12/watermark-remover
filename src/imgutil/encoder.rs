use anyhow::{Context, Result};
use opencv::{core::Vector, imgcodecs, prelude::*};

pub const PROCESSING_TAG: &[u8] = b"\x00PROCESSED_BY_WATERMARK_REMOVER\x00";

/// Returns true if the byte slice ends with our processing tag.
#[inline]
pub fn has_processing_tag(data: &[u8]) -> bool {
    data.ends_with(PROCESSING_TAG)
}

fn encode(img: &(impl MatTraitConst + opencv::core::ToInputArray), ext: &str, params: &[i32]) -> Result<Vec<u8>> {
    let mut buf: Vector<u8> = Vector::new();
    let p: Vector<i32> = Vector::from_slice(params);
    imgcodecs::imencode(ext, img, &mut buf, &p).context("imencode failed")?;
    let mut out = buf.to_vec();
    out.extend_from_slice(PROCESSING_TAG);
    Ok(out)
}

/// Encode a BGR Mat to WebP bytes (with processing tag appended).
pub fn encode_webp(img: &(impl MatTraitConst + opencv::core::ToInputArray), quality: i32) -> Result<Vec<u8>> {
    encode(img, ".webp", &[imgcodecs::IMWRITE_WEBP_QUALITY, quality])
}

/// Encode a BGR Mat to JPEG bytes (with processing tag appended).
pub fn encode_jpeg(img: &(impl MatTraitConst + opencv::core::ToInputArray), quality: i32) -> Result<Vec<u8>> {
    encode(img, ".jpg", &[imgcodecs::IMWRITE_JPEG_QUALITY, quality])
}
