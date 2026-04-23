use opencv::{core::*, imgproc, imgcodecs, prelude::*};

fn main() -> Result<(), opencv::Error> {
    let mut bg = Mat::zeros(100, 100, CV_8UC3)?.to_mat()?;
    let fg = Mat::new_rows_cols_with_default(20, 20, CV_8UC3, Scalar::all(255.0))?;
    
    // Create a ROI
    let roi = Mat::roi(&bg, Rect::new(10, 10, 20, 20))?;
    
    // Attempt to copy fg into roi using the code from blend_with_alpha
    unsafe {
        let dst_ptr = &roi as *const Mat as *mut Mat;
        fg.copy_to(&mut *dst_ptr)?;
    }

    // Save background and see if the white square appears!
    imgcodecs::imwrite("output/blend_bg.png", &bg, &Vector::new())?;

    Ok(())
}
