use std::collections::HashSet;

/// Admin role — grants Users Panel access and all app permissions.
pub const ROLE_ADMIN: &str = "Admin";

pub const PERM_MORPH_UTIL: &str = "morph_util";
pub const PERM_MORPH_BOOKI: &str = "morph_booki";
pub const PERM_MORPH_ENGI: &str = "morph_engi";
pub const PERM_INBOX_MESSAGE: &str = "inbox_message";

pub const ALL_APP_PERMISSIONS: [&str; 3] = [PERM_MORPH_UTIL, PERM_MORPH_BOOKI, PERM_MORPH_ENGI];

pub fn is_admin(roles: &[String]) -> bool {
    roles.iter().any(|r| r == ROLE_ADMIN)
}

pub fn default_app_permissions() -> Vec<String> {
    ALL_APP_PERMISSIONS.iter().map(|p| (*p).to_string()).collect()
}

pub fn normalize_app_permissions(perms: &[String]) -> Vec<String> {
    let allowed: HashSet<&str> = ALL_APP_PERMISSIONS.iter().copied().collect();
    let mut out: Vec<String> = perms
        .iter()
        .filter(|p| allowed.contains(p.as_str()))
        .cloned()
        .collect();
    out.sort();
    out.dedup();
    out
}

pub fn effective_permissions(roles: &[String], user_permissions: &[String]) -> Vec<String> {
    let mut set: HashSet<String> = HashSet::new();
    if is_admin(roles) {
        for p in ALL_APP_PERMISSIONS {
            set.insert(p.to_string());
        }
    } else {
        for p in normalize_app_permissions(user_permissions) {
            set.insert(p);
        }
    }
    if is_admin(roles) || !set.is_empty() {
        set.insert(PERM_INBOX_MESSAGE.to_string());
    }
    let mut v: Vec<String> = set.into_iter().collect();
    v.sort();
    v
}

pub fn require_role(roles: &[String], required: &str) -> bool {
    roles.iter().any(|r| r == required)
}

pub fn require_any_permission(
    roles: &[String],
    user_permissions: &[String],
    required: &[&str],
) -> bool {
    let perms = effective_permissions(roles, user_permissions);
    required.iter().any(|p| perms.iter().any(|up| up == *p))
}

pub fn roles_for_is_admin(is_admin_flag: bool) -> Vec<String> {
    if is_admin_flag {
        vec![ROLE_ADMIN.to_string()]
    } else {
        vec![]
    }
}
