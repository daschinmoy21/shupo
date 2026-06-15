use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Job {
    pub job_id: String,
    pub owner_id: String,
    pub input_key: String,
    pub size: i64,
    pub content_type: String,
    pub kind: String,
    pub enqueued_at: i64,
    pub attempts: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DlqEntry {
    pub job: Job,
    pub error: String,
    pub error_at: String,
    pub failed_at: i64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum JobState {
    Queued,
    Running,
    Completed,
    Failed,
    Cancelled,
}

impl JobState {
    pub fn is_terminal(self) -> bool {
        matches!(
            self,
            JobState::Completed | JobState::Failed | JobState::Cancelled
        )
    }
}

#[derive(thiserror::Error, Debug)]
pub enum ProcessError {
    #[error("transient: {0}")]
    Transient(String),

    #[error("permanent: {0}")]
    Permanent(String),

    #[error("cancelled")]
    Cancelled,
}
