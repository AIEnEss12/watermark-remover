use moka::sync::Cache as MokaCache;
use std::sync::Arc;

#[derive(Clone)]
pub struct CachedImage {
    pub data: Arc<Vec<u8>>,
    pub content_type: &'static str,
}

pub struct ImageCache {
    inner: MokaCache<String, CachedImage>,
}

impl ImageCache {
    pub fn new(capacity: u64) -> Self {
        Self {
            inner: MokaCache::builder()
                .max_capacity(capacity)
                .build(),
        }
    }

    pub fn get(&self, key: &str) -> Option<CachedImage> {
        self.inner.get(key)
    }

    pub fn insert(&self, key: String, data: Vec<u8>, content_type: &'static str) {
        self.inner.insert(
            key,
            CachedImage {
                data: Arc::new(data),
                content_type,
            },
        );
    }
}
