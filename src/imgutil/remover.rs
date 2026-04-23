use anyhow::{Context, Result};
use opencv::{
    core,
    imgproc,
    photo,
    prelude::*,
};

use super::logo_cache::get_scaled_logo;

/// Remove watermarks from `img` (BGR) and overlay our logo.
pub fn remove_watermark(img: &Mat, bboxes: &[core::Rect], logo_path: &str) -> Result<Mat> {
    let cols = img.cols();
    let rows = img.rows();

    let result = if bboxes.is_empty() {
        img.try_clone().context("clone failed")?
    } else {
        let mut mask = core::Mat::zeros(rows, cols, core::CV_8UC1).context("zeros")?.to_mat()?;
        for &bbox in bboxes {
            imgproc::rectangle(&mut mask, bbox, core::Scalar::all(255.0), -1, imgproc::LINE_8, 0)?;
        }
        let kernel = imgproc::get_structuring_element(imgproc::MORPH_RECT, core::Size::new(7, 7), core::Point::new(-1, -1))?;
        imgproc::dilate(&mask.try_clone()?, &mut mask, &kernel, core::Point::new(-1, -1), 1, core::BORDER_CONSTANT, core::Scalar::all(0.0))?;
        let mut inpainted = core::Mat::default();
        photo::inpaint(img, &mask, &mut inpainted, 30.0, photo::INPAINT_TELEA)?;
        inpainted
    };

    let mut result_bgr = result;
    let logo_path = if logo_path.is_empty() { "logo.png" } else { logo_path };
    let target_w = (cols as f64 * 0.18) as i32;

    if let Ok(Some((logo_bgr, alpha))) = get_scaled_logo(logo_path, target_w) {
        let lw = logo_bgr.cols();
        let lh = logo_bgr.rows();
        let padding = 10;
        let off_x = (cols - lw - padding).max(0);
        let off_y = (rows - lh - padding).max(0);
        let actual_w = lw.min(cols - off_x);
        let actual_h = lh.min(rows - off_y);

        if actual_w > 0 && actual_h > 0 {
            if !bboxes.is_empty() {
                let bx1 = (off_x - 3).max(0);
                let by1 = (off_y - 3).max(0);
                let bx2 = (off_x + actual_w + 3).min(cols);
                let by2 = (off_y + actual_h + 3).min(rows);
                let blur_rect = core::Rect::new(bx1, by1, bx2 - bx1, by2 - by1);
                {
                    let blur_roi = result_bgr.roi(blur_rect)?;
                    let mut blurred = core::Mat::default();
                    imgproc::gaussian_blur(&blur_roi, &mut blurred, core::Size::new(5, 5), 0.0, 0.0, core::BORDER_DEFAULT)?;
                    
                    let bh = by2 - by1;
                    let bw = bx2 - bx1;
                    for y in 0..bh {
                        for x in 0..bw {
                            let src_px = blurred.at_2d::<core::Vec3b>(y, x)?;
                            let dst_px = result_bgr.at_2d_mut::<core::Vec3b>(by1 + y, bx1 + x)?;
                            *dst_px = *src_px;
                        }
                    }
                }
            }

            let logo_crop = logo_bgr.roi(core::Rect::new(0, 0, actual_w, actual_h))?.try_clone()?;
            let alpha_crop = alpha.roi(core::Rect::new(0, 0, actual_w, actual_h))?.try_clone()?;
            blend_with_alpha(&logo_crop, &alpha_crop, &mut result_bgr, off_x, off_y, actual_w, actual_h)?;
        }
    }

    Ok(result_bgr)
}

fn blend_with_alpha(logo_bgr: &core::Mat, alpha: &core::Mat, dst_parent: &mut core::Mat, off_x: i32, off_y: i32, w: i32, h: i32) -> Result<()> {
    let dst = dst_parent.roi(core::Rect::new(off_x, off_y, w, h))?.try_clone()?;

    let mut logo_f  = core::Mat::default();
    let mut dst_f   = core::Mat::default();
    let mut alpha_f = core::Mat::default();

    logo_bgr.convert_to(&mut logo_f,  core::CV_32FC3, 1.0 / 255.0, 0.0)?;
    dst.convert_to(&mut dst_f,        core::CV_32FC3, 1.0 / 255.0, 0.0)?;
    alpha.convert_to(&mut alpha_f,    core::CV_32F,   1.0 / 255.0, 0.0)?;

    // Expand alpha to 3 channels.
    let mut alpha3 = core::Mat::default();
    let mut alpha_chans = core::Vector::<core::Mat>::new();
    alpha_chans.push(alpha_f.try_clone()?);
    alpha_chans.push(alpha_f.try_clone()?);
    alpha_chans.push(alpha_f.try_clone()?);
    core::merge(&alpha_chans, &mut alpha3)?;

    // blended = logo * alpha3 + dst * (1 - alpha3)
    let mut logo_premul = core::Mat::default();
    core::multiply(&logo_f, &alpha3, &mut logo_premul, 1.0, -1)?;
    let mut inv_alpha = core::Mat::default();
    alpha3.convert_to(&mut inv_alpha, -1, -1.0, 1.0)?;
    let mut dst_premul = core::Mat::default();
    core::multiply(&dst_f, &inv_alpha, &mut dst_premul, 1.0, -1)?;

    let mut blended_f = core::Mat::default();
    core::add(&logo_premul, &dst_premul, &mut blended_f, &core::no_array(), -1)?;

    let mut blended_u8 = core::Mat::default();
    blended_f.convert_to(&mut blended_u8, core::CV_8UC3, 255.0, 0.0)?;

    for y in 0..h {
        for x in 0..w {
            let src_px = blended_u8.at_2d::<core::Vec3b>(y, x)?;
            let dst_px = dst_parent.at_2d_mut::<core::Vec3b>(off_y + y, off_x + x)?;
            *dst_px = *src_px;
        }
    }

    Ok(())
}
