use anyhow::{Context, Result};
use opencv::{
    core,
    imgproc,
    prelude::*,
};

/// Returns bounding boxes (in original image coordinates) of all detected
/// watermark regions. Detection is limited to the bottom-right corner.
pub fn detect_watermark(img: &core::Mat) -> Result<Vec<core::Rect>> {
    let rows = img.rows();
    let cols = img.cols();

    // Focus on bottom-right corner: right 25% × bottom 8%.
    let zone = core::Rect::new(
        (cols as f64 * 0.75) as i32,
        (rows as f64 * 0.92) as i32,
        (cols as f64 * 0.25) as i32,
        (rows as f64 * 0.08) as i32,
    );
    let full_img_rect = core::Rect::new(0, 0, cols, rows);
    let zone = zone & full_img_rect;
    if zone.empty() {
        return Ok(vec![]);
    }

    let crop = core::Mat::roi(img, zone).context("roi crop failed")?;

    // Grayscale (img is in native BGR format).
    let mut gray = core::Mat::default();
    imgproc::cvt_color(&crop, &mut gray, imgproc::COLOR_BGR2GRAY, 0)
        .context("cvt_color BGR→GRAY failed")?;

    // Otsu threshold.
    let mut thresh = core::Mat::default();
    imgproc::threshold(
        &gray,
        &mut thresh,
        0.0,
        255.0,
        imgproc::THRESH_BINARY | imgproc::THRESH_OTSU,
    )
    .context("threshold failed")?;

    // Connected components with stats.
    let mut labels  = core::Mat::default();
    let mut stats   = core::Mat::default();
    let mut centroids = core::Mat::default();
    let num_labels = imgproc::connected_components_with_stats(
        &thresh,
        &mut labels,
        &mut stats,
        &mut centroids,
        8,
        core::CV_32S,
    )
    .context("connected_components_with_stats failed")?;

    let mut results = Vec::new();

    for i in 1..num_labels {
        let area   = *stats.at_2d::<i32>(i, imgproc::CC_STAT_AREA as i32).unwrap_or(&0);
        let left   = *stats.at_2d::<i32>(i, imgproc::CC_STAT_LEFT as i32).unwrap_or(&0);
        let top    = *stats.at_2d::<i32>(i, imgproc::CC_STAT_TOP as i32).unwrap_or(&0);
        let width  = *stats.at_2d::<i32>(i, imgproc::CC_STAT_WIDTH as i32).unwrap_or(&0);
        let height = *stats.at_2d::<i32>(i, imgproc::CC_STAT_HEIGHT as i32).unwrap_or(&0);

        // Ultra-sensitive filter matching Go code.
        if area > 10 && area < 15000 {
            let x1 = (zone.x + left - 3).max(0);
            let y1 = (zone.y + top  - 3).max(0);
            let x2 = (zone.x + left + width  + 3).min(cols);
            let y2 = (zone.y + top  + height + 3).min(rows);
            results.push(core::Rect::new(x1, y1, x2 - x1, y2 - y1));
        }
    }

    Ok(results)
}
