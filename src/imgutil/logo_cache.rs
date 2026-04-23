use anyhow::{bail, Context, Result};
use once_cell::sync::OnceCell;
use opencv::{
    core,
    imgcodecs, imgproc,
    prelude::*,
};
use std::collections::HashMap;
use std::sync::Mutex;

/// Wrapper so cached Mat can live in a static (Mat is not Sync by default,
/// but read-only access via Clone is safe here).
struct MatSync(core::Mat);
unsafe impl Send for MatSync {}
unsafe impl Sync for MatSync {}

static FULL_LOGO: OnceCell<MatSync> = OnceCell::new();
// Map: target_width → (bgr_mat, alpha_mat)
static LOGO_CACHE: OnceCell<Mutex<HashMap<i32, (core::Mat, core::Mat)>>> = OnceCell::new();

fn logo_cache() -> &'static Mutex<HashMap<i32, (core::Mat, core::Mat)>> {
    LOGO_CACHE.get_or_init(|| Mutex::new(HashMap::new()))
}

/// Returns clones of a cached (BGR_logo, alpha_mask) pair scaled to `target_w`.
pub fn get_scaled_logo(logo_path: &str, target_w: i32) -> Result<Option<(core::Mat, core::Mat)>> {
    // Load full-res logo once.
    let full_wrapper = FULL_LOGO.get_or_try_init(|| {
        let logo = imgcodecs::imread(logo_path, imgcodecs::IMREAD_UNCHANGED)
            .context("failed to load logo")?;
        if logo.empty() {
            bail!("logo file is empty or not found: {}", logo_path);
        }
        Ok::<_, anyhow::Error>(MatSync(logo))
    })?;

    let full = &full_wrapper.0;

    // Look up or compute scaled version.
    {
        let cache = logo_cache().lock().unwrap();
        if let Some((bgr, alpha)) = cache.get(&target_w) {
            return Ok(Some((bgr.try_clone()?, alpha.try_clone()?)));
        }
    }

    // Scale with INTER_LANCZOS4 (matches Go Lanczos quality).
    let orig_w = full.cols();
    let orig_h = full.rows();
    if orig_w == 0 { return Ok(None); }

    let scale   = target_w as f64 / orig_w as f64;
    let target_h = (orig_h as f64 * scale) as i32;

    let mut resized = core::Mat::default();
    imgproc::resize(
        full,
        &mut resized,
        core::Size::new(target_w, target_h),
        0.0, 0.0,
        imgproc::INTER_LANCZOS4,
    )
    .context("logo resize failed")?;

    // Split into BGR + alpha.
    let channels = resized.channels();
    let (bgr, alpha) = if channels == 4 {
        let mut chans: core::Vector<core::Mat> = core::Vector::new();
        core::split(&resized, &mut chans).context("logo split failed")?;
        let b = chans.get(0)?;
        let g = chans.get(1)?;
        let r = chans.get(2)?;
        let a = chans.get(3)?;
        let mut bgr_chans: core::Vector<core::Mat> = core::Vector::new();
        bgr_chans.push(b);
        bgr_chans.push(g);
        bgr_chans.push(r);
        let mut bgr_mat = core::Mat::default();
        core::merge(&bgr_chans, &mut bgr_mat).context("logo merge BGR failed")?;
        (bgr_mat, a)
    } else {
        // No alpha — use all-255 mask.
        let mut bgr_mat = core::Mat::default();
        resized.copy_to(&mut bgr_mat)?;
        let white = core::Mat::new_rows_cols_with_default(
                target_h, target_w,
                core::CV_8UC1,
                core::Scalar::all(255.0),
            )?;
        (bgr_mat, white)
    };

    let ret = (bgr.try_clone()?, alpha.try_clone()?);

    {
        let mut cache = logo_cache().lock().unwrap();
        cache.entry(target_w).or_insert((bgr, alpha));
    }

    Ok(Some(ret))
}
