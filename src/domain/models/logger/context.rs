use uuid::Uuid;

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct LoggerContext {
    pub tag: String,
    pub actor: Option<String>,
    pub trace_id: Option<Uuid>,
}
