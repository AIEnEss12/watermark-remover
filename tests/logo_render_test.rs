use opencv::{core::*, imgproc, imgcodecs, prelude::*};

fn main() -> Result<(), opencv::Error> {
    let mut bg = Mat::new_rows_cols_with_default(100, 100, CV_8UC3, Scalar::all(0.0))?; // Black
    let fg = Mat::new_rows_cols_with_default(20, 20, CV_8UC3, Scalar::new(255.0, 0.0, 0.0, 0.0))?; // Blue
    let alpha = Mat::new_rows_cols_with_default(20, 20, CV_8UC1, Scalar::all(255.0))?; // Opaque

    // Let's paste fg over bg.
    let w = 20; let h = 20;
    
    let dst = Mat::roi(&bg, Rect::new(10, 10, w, h))?;
    let mut logo_f  = Mat::default();
    let mut dst_f   = Mat::default();
    let mut alpha_f = Mat::default();

    fg.convert_to(&mut logo_f,  CV_32FC3, 1.0 / 255.0, 0.0)?;
    dst.convert_to(&mut dst_f,        CV_32FC3, 1.0 / 255.0, 0.0)?;
    alpha.convert_to(&mut alpha_f,    CV_32F,   1.0 / 255.0, 0.0)?;

    let mut alpha3 = Mat::default();
    let mut alpha_chans: Vector<Mat> = Vector::new();
    alpha_chans.push(alpha_f.try_clone()?);
    alpha_chans.push(alpha_f.try_clone()?);
    alpha_chans.push(alpha_f.try_clone()?);
    merge(&alpha_chans, &mut alpha3)?;

    let mut logo_premul = Mat::default();
    multiply(&logo_f, &alpha3, &mut logo_premul, 1.0, -1)?;

    let mut inv_alpha = Mat::default();
    subtract(&Scalar::all(1.0), &alpha3, &mut inv_alpha, &no_array(), -1)?;

    let mut dst_premul = Mat::default();
    multiply(&dst_f, &inv_alpha, &mut dst_premul, 1.0, -1)?;

    let mut blended_f = Mat::default();
    add(&logo_premul, &dst_premul, &mut blended_f, &no_array(), -1)?;

    let mut blended_u8 = Mat::default();
    blended_f.convert_to(&mut blended_u8, CV_8UC3, 255.0, 0.0)?;

    for y in 0..h {
        for x in 0..w {
            let src_px = blended_u8.at_2d::<Vec3b>(y, x).unwrap();
            let dst_px = bg.at_2d_mut::<Vec3b>(10 + y, 10 + x).unwrap();
            *dst_px = *src_px;
        }
    }

    imgcodecs::imwrite("output/debug_overlay.png", &bg, &Vector::new())?;
    Ok(())
}
