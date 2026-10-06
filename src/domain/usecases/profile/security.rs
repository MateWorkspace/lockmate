use crate::domain::{
    models::{AccessClaims, AppContext},
    usecases::UsecaseFuture,
};

pub trait Security: Send + Sync {
    fn change_password<'a>(
        &'a self,
        context: &'a AppContext,
        caller: &'a AccessClaims,
        input: ChangePasswordRequest,
    ) -> UsecaseFuture<'a, ()>;
}

pub struct ChangePasswordRequest {
    pub current_password: String,
    pub new_password: String,
}
