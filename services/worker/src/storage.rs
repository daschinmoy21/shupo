use async_trait::async_trait;
use std::path::PathBuf;

#[async_trait]
pub trait Storage {
    async fn download(&self, key: &str) -> anyhow::Result<PathBuf>;
    async fn upload(&self, key: &str, path: &PathBuf, content_type: &str) -> anyhow::Result<()>;
}

pub struct S3Storage;

impl S3Storage {
    pub async fn new(
        _endpoint: &str,
        _bucket: &str,
        _access_key: &str,
        _secret_key: &str,
    ) -> anyhow::Result<Self> {
        Ok(Self)
    }
}

#[async_trait]
impl Storage for S3Storage {
    async fn download(&self, _key: &str) -> anyhow::Result<PathBuf> {
        Ok(PathBuf::new())
    }
    async fn upload(&self, _key: &str, _path: &PathBuf, _content_type: &str) -> anyhow::Result<()> {
        Ok(())
    }
}
