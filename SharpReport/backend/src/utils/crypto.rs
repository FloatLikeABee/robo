use aes_gcm::aead::generic_array::GenericArray;
use aes_gcm::{
    Aes256Gcm, Nonce,
    aead::{Aead, AeadCore, KeyInit, OsRng},
};
use base64::{Engine as _, engine::general_purpose};

pub struct CryptoService {
    key: Vec<u8>,
}

impl CryptoService {
    pub fn new(secret: &str) -> Self {
        // Derive a 32-byte key from the secret
        let mut key = vec![0u8; 32];
        let secret_bytes = secret.as_bytes();
        for i in 0..32 {
            key[i] = secret_bytes[i % secret_bytes.len()];
        }
        Self { key }
    }

    pub fn encrypt(&self, data: &str) -> Result<String, CryptoError> {
        let cipher = Aes256Gcm::new(GenericArray::from_slice(&self.key));
        let nonce = Aes256Gcm::generate_nonce(&mut OsRng);

        let encrypted_data = cipher
            .encrypt(&nonce, data.as_bytes())
            .map_err(|e| CryptoError::Encryption(e.to_string()))?;

        let mut result = Vec::with_capacity(nonce.len() + encrypted_data.len());
        result.extend_from_slice(&nonce);
        result.extend_from_slice(&encrypted_data);

        Ok(general_purpose::STANDARD.encode(result))
    }

    pub fn decrypt(&self, encrypted_data: &str) -> Result<String, CryptoError> {
        let cipher = Aes256Gcm::new(GenericArray::from_slice(&self.key));

        let decoded_data = general_purpose::STANDARD
            .decode(encrypted_data)
            .map_err(|e| CryptoError::Decoding(e.to_string()))?;

        if decoded_data.len() < 12 {
            return Err(CryptoError::InvalidData);
        }

        let (nonce_bytes, ciphertext) = decoded_data.split_at(12);
        let nonce = Nonce::from_slice(nonce_bytes);

        let decrypted_data = cipher
            .decrypt(nonce, ciphertext)
            .map_err(|e| CryptoError::Decryption(e.to_string()))?;

        String::from_utf8(decrypted_data).map_err(|e| CryptoError::Utf8(e.to_string()))
    }
}

#[derive(Debug, thiserror::Error)]
pub enum CryptoError {
    #[error("Encryption error: {0}")]
    Encryption(String),

    #[error("Decryption error: {0}")]
    Decryption(String),

    #[error("Base64 decoding error: {0}")]
    Decoding(String),

    #[error("UTF-8 conversion error: {0}")]
    Utf8(String),

    #[error("Invalid encrypted data")]
    InvalidData,
}
