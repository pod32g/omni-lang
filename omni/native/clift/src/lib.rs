use std::ffi::CStr;
use std::os::raw::{c_char, c_int};
use cranelift_codegen::ir::{Function, InstBuilder, types::*, AbiParam, Signature};
use cranelift_codegen::settings::{self, Configurable};
use cranelift_codegen::isa::CallConv;
use cranelift_frontend::{FunctionBuilder, FunctionBuilderContext};
use cranelift_module::Module;
use cranelift_object::{ObjectModule, ObjectBuilder};
use target_lexicon::Triple;
use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum CompileError {
    #[error("Invalid MIR JSON: {0}")]
    InvalidJson(String),
    #[error("MIR parsing error: {0}")]
    MirParse(String),
    #[error("Cranelift compilation error: {0}")]
    CraneliftError(String),
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}

#[derive(Debug, Deserialize, Serialize)]
struct MirModule {
    functions: Vec<MirFunction>,
}

#[derive(Debug, Deserialize, Serialize)]
struct MirFunction {
    name: String,
    return_type: String,
    params: Vec<MirParam>,
    blocks: Vec<MirBlock>,
}

#[derive(Debug, Deserialize, Serialize)]
struct MirParam {
    name: String,
    param_type: String,
    id: u32,
}

#[derive(Debug, Deserialize, Serialize)]
struct MirBlock {
    name: String,
    instructions: Vec<MirInstruction>,
    terminator: MirTerminator,
}

#[derive(Debug, Deserialize, Serialize)]
struct MirInstruction {
    id: u32,
    op: String,
    inst_type: String,
    operands: Vec<MirOperand>,
}

#[derive(Debug, Deserialize, Serialize)]
struct MirTerminator {
    op: String,
    operands: Vec<MirOperand>,
}

#[derive(Debug, Deserialize, Serialize)]
struct MirOperand {
    kind: String, // "value" or "literal"
    value: Option<u32>,
    literal: Option<String>,
    operand_type: String,
}

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

    // Parse and validate MIR JSON
    match serde_json::from_str::<MirModule>(payload) {
        Ok(_) => 0,
        Err(e) => {
            eprintln!("MIR JSON parse error: {}", e);
            -4
        }
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

    let mir_payload = match mir_str.to_str() {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let output_path_str = match output_str.to_str() {
        Ok(s) => s,
        Err(_) => return -3,
    };

    match compile_mir_to_object(mir_payload, output_path_str) {
        Ok(_) => 0,
        Err(e) => {
            eprintln!("Compilation error: {}", e);
            -4
        }
    }
}

fn compile_mir_to_object(mir_json: &str, output_path: &str) -> Result<(), CompileError> {
    // Parse MIR JSON
    let mir_module: MirModule = serde_json::from_str(mir_json)
        .map_err(|e| CompileError::InvalidJson(e.to_string()))?;

    // Set up Cranelift
    let mut flag_builder = settings::builder();
    flag_builder.set("opt_level", "speed").unwrap();
    let flags = settings::Flags::new(flag_builder);
    let isa = cranelift_codegen::isa::lookup(Triple::host())
        .map_err(|e| CompileError::CraneliftError(e.to_string()))?
        .finish(flags)
        .map_err(|e| CompileError::CraneliftError(e.to_string()))?;

    // Create object module
    let object_builder = ObjectBuilder::new(isa, "omni", cranelift_module::default_libcall_names())
        .map_err(|e| CompileError::CraneliftError(e.to_string()))?;
    let mut object_module = ObjectModule::new(object_builder);

    // Compile each function
    for mir_func in &mir_module.functions {
        compile_function(&mut object_module, mir_func)?;
    }

    // Generate object file
    let object_product = object_module.finish();
    let object_data = object_product.emit()
        .map_err(|e| CompileError::CraneliftError(e.to_string()))?;
    std::fs::write(output_path, object_data)
        .map_err(|e| CompileError::IoError(e))?;

    Ok(())
}

fn compile_function(
    object_module: &mut ObjectModule,
    mir_func: &MirFunction,
) -> Result<(), CompileError> {
    // Create Cranelift function
    let mut sig = Signature::new(CallConv::SystemV);
    
    // Add parameters
    for param in &mir_func.params {
        let param_type = omni_type_to_cranelift(&param.param_type)?;
        sig.params.push(AbiParam::new(param_type));
    }
    
    // Add return type
    let return_type = omni_type_to_cranelift(&mir_func.return_type)?;
    sig.returns.push(AbiParam::new(return_type));

    let mut func = Function::with_name_signature(
        cranelift_codegen::ir::UserFuncName::user(0, 0), // Use index 0 for now
        sig,
    );

    // Build function body
    let mut builder_ctx = FunctionBuilderContext::new();
    let mut builder = FunctionBuilder::new(&mut func, &mut builder_ctx);

    // Create entry block
    let entry_block = builder.create_block();
    builder.append_block_params_for_function_params(entry_block);
    builder.switch_to_block(entry_block);

    // Compile basic blocks
    let mut block_map = std::collections::HashMap::new();
    for (i, mir_block) in mir_func.blocks.iter().enumerate() {
        let block = if i == 0 {
            entry_block
        } else {
            builder.create_block()
        };
        block_map.insert(mir_block.name.clone(), block);
    }

    // Compile instructions for each block
    for mir_block in &mir_func.blocks {
        let block = block_map[&mir_block.name];
        builder.switch_to_block(block);

        // Compile instructions
        for mir_inst in &mir_block.instructions {
            compile_instruction(&mut builder, mir_inst)?;
        }

        // Compile terminator
        compile_terminator(&mut builder, &mir_block.terminator, &block_map)?;
    }

    // Finalize function
    builder.finalize();

    // Add function to module
    let func_id = object_module.declare_function(
        &mir_func.name,
        cranelift_module::Linkage::Export,
        &func.signature,
    ).map_err(|e| CompileError::CraneliftError(e.to_string()))?;

    // Create a context for the function
    let mut ctx = cranelift_codegen::Context::new();
    ctx.func = func;
    object_module.define_function(func_id, &mut ctx)
        .map_err(|e| CompileError::CraneliftError(e.to_string()))?;

    Ok(())
}

fn omni_type_to_cranelift(omni_type: &str) -> Result<cranelift_codegen::ir::Type, CompileError> {
    match omni_type {
        "int" => Ok(I32),
        "float" => Ok(F32),
        "bool" => Ok(I8), // Use I8 for bool for now
        "void" => Ok(I32), // Use I32 for void for now
        _ => Err(CompileError::MirParse(format!("Unsupported type: {}", omni_type))),
    }
}

fn compile_instruction(
    builder: &mut FunctionBuilder,
    mir_inst: &MirInstruction,
) -> Result<(), CompileError> {
    match mir_inst.op.as_str() {
        "const.int" => {
            let value = mir_inst.operands[0].literal.as_ref()
                .ok_or_else(|| CompileError::MirParse("Expected literal operand".to_string()))?
                .parse::<i32>()
                .map_err(|_| CompileError::MirParse("Invalid integer literal".to_string()))?;
            
            let val = builder.ins().iconst(I32, value as i64);
            // Store the result for later use
            // TODO: Implement proper value mapping
        }
        _ => {
            return Err(CompileError::MirParse(format!("Unsupported instruction: {}", mir_inst.op)));
        }
    }
    Ok(())
}

fn compile_terminator(
    builder: &mut FunctionBuilder,
    terminator: &MirTerminator,
    _block_map: &std::collections::HashMap<String, cranelift_codegen::ir::Block>,
) -> Result<(), CompileError> {
    match terminator.op.as_str() {
        "ret" => {
            if terminator.operands.is_empty() {
                builder.ins().return_(&[]);
            } else {
                // TODO: Handle return values
                builder.ins().return_(&[]);
            }
        }
        _ => {
            return Err(CompileError::MirParse(format!("Unsupported terminator: {}", terminator.op)));
        }
    }
    Ok(())
}
