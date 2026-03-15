use serde::{Deserialize, Serialize};
use tauri::AppHandle;
use tauri_plugin_store::StoreExt;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub backend_url: String,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            backend_url: String::new(),
        }
    }
}

#[tauri::command]
pub async fn get_config(app: AppHandle) -> Result<AppConfig, String> {
    let store = app.store("config.json").map_err(|e| e.to_string())?;
    
    let config = store
        .get("config")
        .and_then(|v| serde_json::from_value(v.clone()).ok())
        .unwrap_or_default();
    
    Ok(config)
}

#[tauri::command]
pub async fn set_config(app: AppHandle, config: AppConfig) -> Result<(), String> {
    let store = app.store("config.json").map_err(|e| e.to_string())?;
    
    store.set("config", serde_json::to_value(&config).map_err(|e| e.to_string())?);
    store.save().map_err(|e| e.to_string())?;
    
    Ok(())
}

#[tauri::command]
pub async fn test_connection(url: String) -> Result<bool, String> {
    let client = reqwest::Client::builder()
        .danger_accept_invalid_certs(true)
        .build()
        .map_err(|e| e.to_string())?;
    
    let health_url = format!("{}/health", url.trim_end_matches('/'));
    
    match client.get(&health_url).send().await {
        Ok(resp) if resp.status().is_success() => Ok(true),
        Ok(resp) => Err(format!("Server returned status {}", resp.status())),
        Err(e) => Err(format!("Connection failed: {}", e)),
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiRequest {
    pub url: String,
    pub method: String,
    pub headers: std::collections::HashMap<String, String>,
    #[serde(default)]
    pub body: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiResponse {
    pub status: u16,
    pub headers: std::collections::HashMap<String, String>,
    pub body: String,
}

#[tauri::command]
pub async fn api_request(request: ApiRequest) -> Result<ApiResponse, String> {
    let client = reqwest::Client::builder()
        .danger_accept_invalid_certs(true)
        .build()
        .map_err(|e| e.to_string())?;
    
    let mut req = match request.method.to_uppercase().as_str() {
        "GET" => client.get(&request.url),
        "POST" => client.post(&request.url),
        "PUT" => client.put(&request.url),
        "DELETE" => client.delete(&request.url),
        "PATCH" => client.patch(&request.url),
        _ => return Err(format!("Unsupported method: {}", request.method)),
    };
    
    for (key, value) in &request.headers {
        req = req.header(key, value);
    }
    
    if let Some(body) = &request.body {
        req = req.body(body.clone());
    }
    
    let response = req.send().await.map_err(|e| e.to_string())?;
    
    let status = response.status().as_u16();
    let headers: std::collections::HashMap<String, String> = response
        .headers()
        .iter()
        .filter_map(|(k, v)| Some((k.to_string(), v.to_str().ok()?.to_string())))
        .collect();
    
    let body = response.text().await.unwrap_or_default();
    
    Ok(ApiResponse { status, headers, body })
}
