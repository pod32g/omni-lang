use std::ffi::CStr;
use std::os::raw::{c_char, c_int};

/// Compiles MIR JSON to validate the input format.
///
/// # Safety
/// The `mir_json` pointer must be a valid, null-terminated C string.
/// The caller is responsible for ensuring the pointer is valid for the duration of the call.
#[no_mangle]
pub unsafe extern "C" fn omni_clift_compile_json(mir_json: *const c_char) -> c_int {
    if mir_json.is_null() {
        return -1;
    }

    // Safety: the caller must provide a valid, null-terminated string.
    let bytes = CStr::from_ptr(mir_json);
    let payload = match bytes.to_str() {
        Ok(s) => s,
        Err(_) => return -2,
    };

    if payload.trim().is_empty() {
        return -3;
    }

    // For now, just validate the JSON and return success
    // TODO: Implement actual MIR to Cranelift IR translation
    match serde_json::from_str::<serde_json::Value>(payload) {
        Ok(_) => 0,
        Err(_) => -4,
    }
}

/// Compiles MIR JSON to a native object file.
///
/// # Safety
/// Both `mir_json` and `output_path` pointers must be valid, null-terminated C strings.
/// The caller is responsible for ensuring the pointers are valid for the duration of the call.
#[no_mangle]
pub unsafe extern "C" fn omni_clift_compile_to_object(
    mir_json: *const c_char,
    output_path: *const c_char,
) -> c_int {
    if mir_json.is_null() || output_path.is_null() {
        return -1;
    }

    let mir_str = CStr::from_ptr(mir_json);
    let output_str = CStr::from_ptr(output_path);

    let _mir_payload = match mir_str.to_str() {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let output_path_str = match output_str.to_str() {
        Ok(s) => s,
        Err(_) => return -3,
    };

    // For now, just create a minimal object file
    // TODO: Implement actual compilation
    match std::fs::write(output_path_str, b"# Minimal object file placeholder\n") {
        Ok(_) => 0,
        Err(_) => -4,
    }
}
