use opencv::{core::*, imgproc, photo, imgcodecs, prelude::*};

fn main() -> Result<(), opencv::Error> {
    let mut img = Mat::new_rows_cols_with_default(100, 100, CV_8UC3, Scalar::all(128.0))?;
    
    // Draw a red square (in RGB space, Red is ch 0. Wait, RGB space implies ch 0 is Red).
    // Let's create an RGB image where Red is ch 0.
    let red = Scalar::new(255.0, 0.0, 0.0, 0.0);
    imgproc::rectangle(&mut img, Rect::new(40, 40, 20, 20), red, -1, imgproc::LINE_8, 0)?;

    let mut mask = Mat::zeros(100, 100, CV_8UC1)?.to_mat()?;
    imgproc::rectangle(&mut mask, Rect::new(40, 40, 20, 20), Scalar::all(255.0), -1, imgproc::LINE_8, 0)?;

    let mut inpainted = Mat::default();
    photo::inpaint(&img, &mask, &mut inpainted, 30.0, photo::INPAINT_TELEA)?;

    // Save outputs
    imgcodecs::imwrite("output/orig.jpg", &img, &Vector::new())?;
    imgcodecs::imwrite("output/inpainted.jpg", &inpainted, &Vector::new())?;

    Ok(())
}
