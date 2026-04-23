use anyhow::{Context, Result};
use opencv::prelude::*;
use std::fs;
use std::path::Path;
use watermark_remover::imgutil;

fn main() -> Result<()> {
    let args: Vec<String> = std::env::args().collect();
    
    let mut url = String::new();
    let mut out_dir = "output".to_string();
    let mut format = "webp".to_string();

    let mut i = 1;
    while i < args.len() {
        match args[i].as_str() {
            "--url" | "-u" => {
                if i + 1 < args.len() {
                    url = args[i+1].clone();
                    i += 1;
                }
            }
            "--out" | "-o" => {
                if i + 1 < args.len() {
                    out_dir = args[i+1].clone();
                    i += 1;
                }
            }
            "--format" | "-f" => {
                if i + 1 < args.len() {
                    format = args[i+1].clone();
                    i += 1;
                }
            }
            _ => {
                // If it doesn't start with - or --, treat it as the first missing positional
                if url.is_empty() && !args[i].starts_with('-') {
                    url = args[i].clone();
                }
            }
        }
        i += 1;
    }

    if url.is_empty() {
        anyhow::bail!("Usage: tester --url <URL> [--out <DIR>] [--format <webp|jpeg>]");
    }

    // Run tokio runtime manually
    let rt = tokio::runtime::Runtime::new()?;
    rt.block_on(async {
        // Create output dir
        fs::create_dir_all(&out_dir).context("failed to create output directory")?;

        println!("Downloading image from {}", url);
        let client = reqwest::Client::new();
        let resp = client.get(&url).send().await?;
        if !resp.status().is_success() {
            anyhow::bail!("failed to fetch image: status code {}", resp.status());
        }

        let data = resp.bytes().await?.to_vec();

        // Decode
        let img = imgutil::decode_image(&data).context("failed to decode image")?;

        // Save a copy of the original (converted to webp for comparison)
        let orig_bytes = imgutil::encode_webp(&img, 100).context("failed to encode original to webp")?;
        let copy_path = Path::new(&out_dir).join("copy.webp");
        fs::write(&copy_path, orig_bytes).context("failed to save original copy")?;
        println!("Saved original copy to {}", copy_path.display());

        // Detect
        let bboxes = imgutil::detect_watermark(&img).context("failed to detect watermarks")?;
        println!("Detected {} watermark zones: {:?}", bboxes.len(), bboxes);

        // Remove
        let result = imgutil::remove_watermark(&img, &bboxes, "logo.png").context("failed to process image")?;

        // Encode
        let ext = if format == "webp" { "webp" } else { "jpg" };
        let res_bytes = if format == "webp" {
            imgutil::encode_webp(&result, 80).context("failed to encode result to webp")?
        } else {
            imgutil::encode_jpeg(&result, 90).context("failed to encode result to jpeg")?
        };

        // Save
        let output_path = Path::new(&out_dir).join(format!("test_result.{}", ext));
        fs::write(&output_path, res_bytes).context("failed to save result")?;

        println!("Success! Result saved to {}", output_path.display());

        Ok(())
    })
}
