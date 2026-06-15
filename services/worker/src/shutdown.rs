use tokio::sync::mpsc;

pub fn install() -> mpsc::Receiver<()> {
    let (tx, rx) = mpsc::channel(1);

    tokio::spawn(async move {
        let mut sigterm = tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
            .expect("Install sigterm handler");

        let mut sigkill = tokio::signal::unix::signal(tokio::signal::unix::SignalKind::interrupt())
            .expect("Install sigint handler");

        tokio::select! {
            _ = sigterm.recv() => tracing::info!("SIGNTERM received"),
            _ = sigint.recv() => tracing::info!("SIGINT received"),
        }
        let _ = tx.send(()).await;
    });
    rx
}
