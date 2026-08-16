use morphai::{Client, Config, Message};

#[tokio::main]
async fn main() {
    let cfg = Config::from_env();
    if !cfg.configured() {
        eprintln!("MORPH_AI_API_KEY not set");
        std::process::exit(1);
    }
    let client = Client::new(cfg);
    let content = std::env::var("MORPH_AI_PING_SIZE")
        .ok()
        .and_then(|s| s.parse().ok())
        .map(|n| format!("Reply with exactly: ok. {}", "z".repeat(n)))
        .unwrap_or_else(|| "Reply with exactly: ok".to_string());
    let out = client
        .chat_completion(&[Message {
            role: "user".to_string(),
            content,
        }])
        .await;
    match out {
        Ok(s) => println!("OK: {}", s.trim()),
        Err(e) => {
            eprintln!("ERR: {}", e);
            std::process::exit(1);
        }
    }
}
