use opencv::{core::*, imgproc, imgcodecs, prelude::*};

fn main() -> Result<(), opencv::Error> {
    let logo = imgcodecs::imread("logo.png", imgcodecs::IMREAD_UNCHANGED)?;
    println!("Logo channels: {}", logo.channels());
    let mut chans: Vector<Mat> = Vector::new();
    opencv::core::split(&logo, &mut chans)?;
    if chans.len() == 4 {
        let a = chans.get(3)?;
        let non_zero = opencv::core::count_non_zero(&a)?;
        println!("Alpha non-zero pixels: {}", non_zero);
    }
    
    Ok(())
}
