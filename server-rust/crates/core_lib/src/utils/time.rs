use chrono::{DateTime, TimeZone, Utc};

/// 获取当前 UTC 时间
pub fn now() -> DateTime<Utc> {
    Utc::now()
}

/// 从时间戳创建 DateTime
pub fn from_timestamp(ts: i64) -> DateTime<Utc> {
    Utc.timestamp_opt(ts, 0).unwrap()
}

/// 格式化时间为 RFC3339
pub fn format_rfc3339(dt: &DateTime<Utc>) -> String {
    dt.to_rfc3339()
}
