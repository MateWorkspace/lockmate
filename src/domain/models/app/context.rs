use uuid::Uuid;

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct AppContext {
    pub actor: Option<String>,
    pub trace_id: Option<Uuid>,
}
