use std::{path::Path, process::Command, string::ParseError};

use anyhow::{Ok, Result};
use chrono::{Duration, ParseError};
use ffmpeg_next as ffmpeg;

pub struct MediaInfo {
    pub duration_secs: f64,
    pub width: u32,
    pub height: u32,
    pub codec: String,
}

pub fn probe(input: std::path::Path) -> Result<MediaInfo, WorkerError> {
    ffmpeg::init().map_err(|e| WorkerError::Ffmpeg(e.to_string()))?;

    let input_ctx = ffmpeg_next::format::input(input)
        .map_err(|e| ParseError::Transient(format!("ffmpeg open :{e}")))?;

    let stream = input_ctx
        .streams()
        .best(ffmpeg_next::media::Type::Video)
        .ok_or_else(|| ParseError::Permanent("no video stream".into()))?;

    let codec_param = stream.parameters();

    let decoder_ctx = ffmpeg_next::codec::context::Context::from_parameters(codec_param)
        .map_err(|e| ProcessError::Transient(format!("ffmpeg decoder ctx: {e}")))?;

    let decoder = decoder_ctx
        .decoder()
        .video()
        .map_err(|e| ProcessError::Transient(format!("ffmpeg decoder: {e}")))?;

    let duration_secs = input_ctx.duration() as f64 / ffmpeg_next::ffi::AV_TIME_BASE as f64;

    Ok(MediaInfo {
        duration_secs,
        width: decoder.width(),
        height: decoder.height(),
        codec: codec_param
            .id()
            .map(|id| id.name().to_string())
            .unwrap_or_default(),
    })
}

pub fn extract_thumbnail(input: &Path, output: &Path, at: Duration) -> Result<(), ProcessError> {
    let at_secs = at.as_seconds_f64();
    let status = Command::new("ffmpeg")
        .args([
            "-y",
            "-ss",
            &format!("{at_secs}"),
            "-i",
            input
                .to_str()
                .ok_or_else(|| ProcessError::Permanent("bad input path".into()))?,
            "-vframes",
            "1",
            "-q:v",
            "2",
            output
                .to_str()
                .ok_or_else(|| ProcessError::Permanent("bad output path".into()))?,
        ])
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::piped())
        .status()
        .map_err(|e| ProcessError::Transient(format!("spawn ffmpeg: {e}")))?;

    if !status.success() {
        return Err(ProcessError::Transient("ffmpeg thumbnail failed".into()));
    }
    Ok(())
}
