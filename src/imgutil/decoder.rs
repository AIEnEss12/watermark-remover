use anyhow::{bail, Context, Result};
use opencv::{core::Vector, imgcodecs, imgproc, prelude::*};

/// Decode any image format from raw bytes into a BGR Mat.
///
/// ENCAR JPEGs are stored with RGB channel order; OpenCV decodes them as BGR,
/// which visually swaps R and B. We apply a one-time BGR→RGB swap so the
/// rest of the pipeline works with visually-correct pixels.
pub fn decode_image(data: &[u8]) -> Result<Mat> {
    let buf: Vector<u8> = Vector::from_slice(data);
    let raw = imgcodecs::imdecode(&buf, imgcodecs::IMREAD_COLOR)
        .context("imdecode failed")?;

    if raw.empty() {
        bail!("imdecode returned empty mat");
    }

    Ok(raw) // Natively BGR
}
