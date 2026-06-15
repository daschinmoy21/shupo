use anyhow::Result;
use tracing_subscriber::EnvFilter;

use crate::storage::S3Storage;

mod config;
mod ffmpeg;
mod reclaim;
mod shutdown;
mod state;
mod status;
mod storage;
mod worker;

#[tokio::main]
async fn main() -> Result<()> {
    let filter =
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info,worker=debug"));
    tracing_subscriber::fmt()
        .with_env_filter(filter)
        .json()
        .init();

    let cfg = Config::from_env().context("loading config")?;
    info!(?cfg, "worker starting");

    let redis_client = redis::Client::open(cfg.redis_url.clone())?;
    let redis = redis::aio::ConnectionManager::new(redis_client).await?;

    let storage = S3Storage::new(
        &cfg.s3_endpoint,
        &cfg.s3_bucket,
        &cfg.s3_access_key,
        &cfg.s3_secret_key,
    )
    .await?;

    let status = RedisStatus::new(redis.clone());

    let consumer = format!("worker-{}", uuid::Uuid::new_v4());
    let worker = Arc::new(Worker::new(redis, storage, status, consumer));

    let (reclaim_shutdown_tx, reclaim_shutdown_rx) = tokio::sync::watch::channel(false);
    let reclaim_handle = tokio::spawn({
        let worker = Arc::clone(&worker);
        async move {
            reclaim::reclaim_loop(worker, reclaim_shutdown_rx).await;
        }
    });

    let shutdown_rx = shutdown::install();

    let res = worker.run(shutdown_rx).await;

    let _ = reclaim_shutdown_tx.send(true);
    let _ = reclaim_handle.await;

    info!("worker exited cleanly");
    res
}
