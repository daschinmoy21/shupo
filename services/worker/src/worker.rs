use std::sync::{Arc, mpsc};

use anyhow::Ok;
use tracing::info;

use crate::state::{Job, JobState, ProcessError};

pub const MAX_ATTEMPTS: i8 = 3;
const POLL_TIMEOUT: i32 = 5_000;

pub struct Worker<S, U> {
    pub redis: redis::aio::ConnectionManager,
    pub storage: S,
    pub status: U,
    pub consumer: String,
}

pub struct ProcessOutcome {
    pub thumb: String,
    pub duration: u64,
    pub width: u64,
    pub height: u64,
}

impl ProcessOutcome {
    pub fn to_fileds(&self) -> Vec<(&'static str, String)> {
        vec![
            ("thumb", self.thumb.clone()),
            ("duration", format!("{:.3}", self.duration)),
            ("width", self.width.to_string()),
            ("height", self.height.to_string()),
        ]
    }
}

impl<S, U> Worker<S, U>
where
    S: Storage + Sync + Send + 'static,
    U: StatusUpdater + Send + Sync + 'static,
{
    pub fn new(
        redis: redis::aio::ConnectionManager,
        storage: S,
        status: U,
        consumer: String,
    ) -> Self {
        Self {
            redis,
            storage,
            status,
            consumer,
        };

        pub async fn run(self: Arc<Self>, mut shutdown: mpsc::Receiver<()>) -> anyhow::Result<()> {
            self.ensure_group().await?;
            info!(consumer = %self.consumer,"worker ready");

            loop {
                tokio::select! {
                    _ = shutdown.recv() => {
                        info!("worker shutting down");
                        return Ok(());
                    }
                    job_result = self.read_one_job() => {
                        match job_result {
                            Ok(Some((entry_id, job))) => {
                                self.handle(entry_id, job).await;
                            }
                            Ok(None) => {}
                            Err(e) => {
                                error!(error = %e, "xreadgroup failed; backing off");
                                sleep(Duration::from_secs(1)).await;
                            }
                        }
                    }
                }
            }
        }
    }
    todo!("Complete worker");
}
