use std::ffi::CStr;
use std::os::raw::c_char;

#[no_mangle]
pub extern "C" fn omni_clift_compile_json(mir_json: *const c_char) -> i32 {
    if mir_json.is_null() {
        return -1;
    }

    // Safety: the caller must provide a valid, null-terminated string.
    let bytes = unsafe { CStr::from_ptr(mir_json) };
    let payload = match bytes.to_str() {
        Ok(s) => s,
        Err(_) => return -2,
    };

    if payload.trim().is_empty() {
        return -3;
    }

    // TODO: parse JSON -> Cranelift IR -> object/JIT artifact.
    let _ = payload;

    0
}
