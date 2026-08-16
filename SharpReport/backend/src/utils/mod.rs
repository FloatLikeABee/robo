pub mod crypto;

use std::error::Error;
use std::fmt;

#[derive(Debug)]
pub struct AppError {
    pub message: String,
    pub cause: Option<Box<dyn Error>>,
}

impl AppError {
    pub fn new(message: &str) -> Self {
        Self {
            message: message.to_string(),
            cause: None,
        }
    }

    pub fn with_cause(message: &str, cause: Box<dyn Error>) -> Self {
        Self {
            message: message.to_string(),
            cause: Some(cause),
        }
    }
}

impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.message)?;
        if let Some(cause) = &self.cause {
            write!(f, " (caused by: {})", cause)?;
        }
        Ok(())
    }
}

impl Error for AppError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        self.cause.as_ref().map(|e| e.as_ref())
    }
}

impl From<sqlx::Error> for AppError {
    fn from(err: sqlx::Error) -> Self {
        AppError::with_cause("Database error", Box::new(err))
    }
}

impl From<reqwest::Error> for AppError {
    fn from(err: reqwest::Error) -> Self {
        AppError::with_cause("HTTP error", Box::new(err))
    }
}

impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self {
        AppError::with_cause("IO error", Box::new(err))
    }
}
