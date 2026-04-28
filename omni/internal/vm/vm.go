package vm

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/omni-lang/omni/internal/logging"
	"github.com/omni-lang/omni/internal/mir"
)

// instructionHandler defines the signature for instruction execution functions
type instructionHandler func(map[string]*mir.Function, *frame, mir.Instruction) (Result, error)

// instructionHandlers maps instruction names to their execution functions
var instructionHandlers map[string]instructionHandler

var (
	fileHandleMu      sync.Mutex
	fileHandleCounter = 3
	fileHandleTable   = make(map[int]*os.File)

	testingSuitesMu     sync.Mutex
	testingSuiteCounter int
	testingSuites       = make(map[int]*testingSuite)

	cliArgsMu sync.RWMutex
	cliArgs   []string

	stdinReader = bufio.NewReader(os.Stdin)

	// Async/Promise support
	promiseMu      sync.RWMutex
	promiseCounter int
	promises       = make(map[int]*Promise)
	eventLoop      = make(chan func(), 100) // Event loop channel

	// Coverage tracking
	coverageMu      sync.RWMutex
	coverageEnabled bool
	coverageData    = make(map[string]*coverageEntry)
)

// coverageEntry tracks coverage for a function
type coverageEntry struct {
	FunctionName string `json:"function"`
	FilePath     string `json:"file"`
	LineNumber   int    `json:"line"`
	CallCount    int    `json:"count"`
}

// Promise represents an async operation that may complete in the future
type Promise struct {
	ID      int
	Value   Result
	Error   error
	Done    bool
	Waiters []chan Result // Channels waiting for this promise to resolve
	mu      sync.Mutex
}

type testingSuite struct {
	total  int
	failed int
}

func newTestingSuiteID() int {
	testingSuitesMu.Lock()
	defer testingSuitesMu.Unlock()
	testingSuiteCounter++
	id := testingSuiteCounter
	testingSuites[id] = &testingSuite{}
	return id
}

func ensureTestingSuiteLocked(id int) *testingSuite {
	suite, ok := testingSuites[id]
	if !ok {
		suite = &testingSuite{}
		testingSuites[id] = suite
	}
	return suite
}

func recordTestingResultLocked(suite *testingSuite, name string, passed bool, message string) {
	fmt.Printf("Running test: %s\n", name)
	if passed {
		fmt.Printf("✓ %s PASSED\n", name)
	} else {
		fmt.Printf("✗ %s FAILED\n", name)
		if message != "" {
			fmt.Printf("  %s\n", message)
		}
		suite.failed++
	}
	suite.total++
}

func suiteIDFromResult(res Result) (int, bool) {
	if res.Type == "std.testing.Suite" {
		if data, ok := res.Value.(map[string]interface{}); ok {
			switch v := data["id"].(type) {
			case int:
				return v, true
			case int64:
				return int(v), true
			case int32:
				return int(v), true
			case uint64:
				return int(v), true
			case uint32:
				return int(v), true
			case float64:
				return int(v), true
			}
		}
	}
	id, err := toInt(res)
	if err != nil {
		return 0, false
	}
	return id, true
}

func stringFromResult(res Result) string {
	if str, err := toString(res); err == nil {
		return str
	}
	return fmt.Sprint(res.Value)
}

func boolFromResult(res Result) bool {
	b, err := toBool(res)
	return err == nil && b
}

func SetCLIArgs(args []string) {
	cliArgsMu.Lock()
	defer cliArgsMu.Unlock()
	if len(args) == 0 {
		cliArgs = nil
		return
	}
	cliArgs = append([]string(nil), args...)
}

func cloneCLIArgs() []string {
	cliArgsMu.RLock()
	defer cliArgsMu.RUnlock()
	if len(cliArgs) == 0 {
		return nil
	}
	out := make([]string, len(cliArgs))
	copy(out, cliArgs)
	return out
}

func readLineFromStdin() (string, error) {
	line, err := stdinReader.ReadString('\n')
	if errors.Is(err, io.EOF) {
		if len(line) == 0 {
			return "", io.EOF
		}
	} else if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func readAllFromStdin() (string, error) {
	data, err := io.ReadAll(stdinReader)
	if err != nil {
		return string(data), err
	}
	return string(data), nil
}

// parseIntStrict mirrors omni_io_parse_int / omni_io_is_int: requires
// the entire string to be a decimal integer in int32 range.
func parseIntStrict(s string) (int32, bool) {
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(n), true
}

func parseFloatStrict(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// decodeStringLiteral strips wrapping double quotes and decodes the
// escape sequences accepted by the lexer: \n, \t, \r, \\, \", \', \0,
// \xHH, \uHHHH. Anything else inside a backslash is left as the
// backslash + char (defensive — the lexer should have rejected unknown
// escapes already).
func decodeStringLiteral(raw string) string {
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	if !strings.ContainsRune(raw, '\\') {
		return raw
	}
	var out strings.Builder
	out.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' || i+1 >= len(raw) {
			out.WriteByte(raw[i])
			continue
		}
		switch raw[i+1] {
		case 'n':
			out.WriteByte('\n')
			i++
		case 't':
			out.WriteByte('\t')
			i++
		case 'r':
			out.WriteByte('\r')
			i++
		case '\\':
			out.WriteByte('\\')
			i++
		case '"':
			out.WriteByte('"')
			i++
		case '\'':
			out.WriteByte('\'')
			i++
		case '0':
			out.WriteByte(0)
			i++
		case 'x':
			if i+3 < len(raw) {
				if v, err := strconv.ParseUint(raw[i+2:i+4], 16, 8); err == nil {
					out.WriteByte(byte(v))
					i += 3
					continue
				}
			}
			out.WriteByte(raw[i])
		case 'u':
			if i+5 < len(raw) {
				if v, err := strconv.ParseUint(raw[i+2:i+6], 16, 32); err == nil {
					out.WriteRune(rune(v))
					i += 5
					continue
				}
			}
			out.WriteByte(raw[i])
		default:
			out.WriteByte(raw[i])
		}
	}
	return out.String()
}

// vmStringSlice extracts an []string from a Result whose value is an
// array<string>. Falls back to a permissive []interface{} unbox.
func vmStringSlice(r Result) []string {
	if s, ok := r.Value.([]string); ok {
		return s
	}
	if as, ok := r.Value.([]interface{}); ok {
		out := make([]string, 0, len(as))
		for _, a := range as {
			if s, ok := a.(string); ok {
				out = append(out, s)
			} else {
				out = append(out, fmt.Sprint(a))
			}
		}
		return out
	}
	return nil
}

// ansiWrap returns s wrapped in an SGR escape sequence with the given code.
func ansiWrap(s, code string) string {
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func ansiIntrinsic(fr *frame, operands []mir.Operand, code string) (Result, bool) {
	if len(operands) == 1 {
		s, _ := operandValue(fr, operands[0]).Value.(string)
		return Result{Type: "string", Value: ansiWrap(s, code)}, true
	}
	return Result{}, false
}

// vmSprintf mirrors omni_io_sprintf in the C runtime: %s is replaced in
// order by entries in args; %% emits a literal '%'; other % directives
// are left intact.
func vmSprintf(format string, args []string) string {
	var out strings.Builder
	out.Grow(len(format))
	argIdx := 0
	for i := 0; i < len(format); {
		if i+1 < len(format) && format[i] == '%' {
			switch format[i+1] {
			case 's':
				if argIdx < len(args) {
					out.WriteString(args[argIdx])
				}
				argIdx++
				i += 2
				continue
			case '%':
				out.WriteByte('%')
				i += 2
				continue
			}
		}
		out.WriteByte(format[i])
		i++
	}
	return out.String()
}

// newPromise creates a new promise and returns its ID
func newPromise() int {
	promiseMu.Lock()
	defer promiseMu.Unlock()
	promiseCounter++
	id := promiseCounter
	promises[id] = &Promise{
		ID:      id,
		Waiters: make([]chan Result, 0),
	}
	return id
}

// resolvePromise resolves a promise with a value
func resolvePromise(id int, value Result) {
	promiseMu.Lock()
	promise, ok := promises[id]
	if !ok {
		promiseMu.Unlock()
		return
	}
	promiseMu.Unlock()

	promise.mu.Lock()
	promise.Value = value
	promise.Done = true
	waiters := promise.Waiters
	promise.Waiters = nil
	promise.mu.Unlock()

	// Notify all waiters
	for _, ch := range waiters {
		select {
		case ch <- value:
		default:
		}
	}
}

// rejectPromise rejects a promise with an error
func rejectPromise(id int, err error) {
	promiseMu.Lock()
	promise, ok := promises[id]
	if !ok {
		promiseMu.Unlock()
		return
	}
	promiseMu.Unlock()

	promise.mu.Lock()
	promise.Error = err
	promise.Done = true
	waiters := promise.Waiters
	promise.Waiters = nil
	promise.mu.Unlock()

	// Notify all waiters with error result
	errorResult := Result{Type: "error", Value: err.Error()}
	for _, ch := range waiters {
		select {
		case ch <- errorResult:
		default:
		}
	}
}

// awaitPromise waits for a promise to resolve
func awaitPromise(id int) (Result, error) {
	promiseMu.RLock()
	promise, ok := promises[id]
	promiseMu.RUnlock()

	if !ok {
		return Result{}, fmt.Errorf("promise %d not found", id)
	}

	promise.mu.Lock()
	if promise.Done {
		value := promise.Value
		err := promise.Error
		promise.mu.Unlock()
		if err != nil {
			return Result{}, err
		}
		return value, nil
	}

	// Create a channel to wait for resolution
	ch := make(chan Result, 1)
	promise.Waiters = append(promise.Waiters, ch)
	promise.mu.Unlock()

	// Wait for the promise to resolve
	result := <-ch
	if result.Type == "error" {
		return Result{}, fmt.Errorf("%v", result.Value)
	}
	return result, nil
}

// execAwait handles await instructions
func execAwait(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("await instruction requires operand")
	}

	operand := operandValue(fr, inst.Operands[0])

	// Check if operand is a Promise
	if operand.Type == "Promise" {
		// Extract promise ID from value
		if promiseID, ok := operand.Value.(int); ok {
			result, err := awaitPromise(promiseID)
			if err != nil {
				return Result{}, err
			}
			return result, nil
		}
		// If value is a map with promise data, extract ID
		if promiseMap, ok := operand.Value.(map[string]interface{}); ok {
			if idVal, ok := promiseMap["id"]; ok {
				if promiseID, ok := idVal.(int); ok {
					result, err := awaitPromise(promiseID)
					if err != nil {
						return Result{}, err
					}
					return result, nil
				}
			}
		}
		return Result{}, fmt.Errorf("await: invalid promise value")
	}

	// If not a Promise, return as-is (for now - this allows awaiting non-promises for compatibility)
	return operand, nil
}

func hasFlag(name string) bool {
	args := cloneCLIArgs()
	if len(args) == 0 {
		return false
	}
	bare := "--" + name
	withValue := bare + "="
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, withValue) || arg == bare {
			return true
		}
	}
	return false
}

func getFlagValue(name, defaultValue string) string {
	args := cloneCLIArgs()
	if len(args) == 0 {
		return defaultValue
	}
	bare := "--" + name
	withValue := bare + "="
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, withValue) {
			return arg[len(withValue):]
		}
		if arg == bare {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				return args[i+1]
			}
			return "true"
		}
	}
	return defaultValue
}

func positionalArgValue(index int, defaultValue string) string {
	if index < 0 {
		return defaultValue
	}
	args := cloneCLIArgs()
	if len(args) == 0 {
		return defaultValue
	}
	positionals := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				i++
			}
			continue
		}
		if positionals == index {
			return arg
		}
		positionals++
	}
	return defaultValue
}

func buildTestingSuiteStruct(id int) map[string]interface{} {
	return map[string]interface{}{
		"id": id,
	}
}

func execTestingEqualFloat(fr *frame, operands []mir.Operand, precision int) (Result, bool) {
	if len(operands) < 4 {
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	}
	if precision < 0 {
		precision = 6
	}
	suiteIDRes := operandValue(fr, operands[0])
	id, ok := suiteIDFromResult(suiteIDRes)
	if !ok {
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	}
	name := stringFromResult(operandValue(fr, operands[1]))
	expectedRes := operandValue(fr, operands[2])
	actualRes := operandValue(fr, operands[3])
	expectedVal, expErr := toFloat(expectedRes)
	actualVal, actErr := toFloat(actualRes)

	passed := expErr == nil && actErr == nil
	if passed {
		tolerance := math.Pow10(-precision)
		if tolerance == 0 {
			tolerance = math.SmallestNonzeroFloat64
		}
		passed = math.Abs(expectedVal-actualVal) <= tolerance
	}

	format := fmt.Sprintf("%%.%df", precision)
	var message string
	if expErr != nil || actErr != nil {
		message = fmt.Sprintf("expected %v, got %v", expectedRes.Value, actualRes.Value)
	} else {
		message = fmt.Sprintf("expected "+format+", got "+format, expectedVal, actualVal)
	}

	testingSuitesMu.Lock()
	suite := ensureTestingSuiteLocked(id)
	recordTestingResultLocked(suite, name, passed, message)
	testingSuitesMu.Unlock()

	return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
}

type exitSignal struct {
	code int
}

// ExitError represents a VM-triggered process exit (e.g. std.os.exit).
type ExitError struct {
	Code int
}

func (e ExitError) Error() string {
	return fmt.Sprintf("vm exit with code %d", e.Code)
}

func init() {
	instructionHandlers = map[string]instructionHandler{
		"const":            execConst,
		"add":              execArithmetic,
		"sub":              execArithmetic,
		"mul":              execArithmetic,
		"div":              execArithmetic,
		"mod":              execArithmetic,
		"bitand":           execBitwise,
		"bitor":            execBitwise,
		"bitxor":           execBitwise,
		"lshift":           execBitwise,
		"rshift":           execBitwise,
		"strcat":           execStringConcat,
		"neg":              execUnary,
		"not":              execUnary,
		"bitnot":           execUnary,
		"cast":             execCast,
		"cmp.eq":           execComparison,
		"cmp.neq":          execComparison,
		"cmp.lt":           execComparison,
		"cmp.lte":          execComparison,
		"cmp.gt":           execComparison,
		"cmp.gte":          execComparison,
		"and":              execLogical,
		"or":               execLogical,
		"call":             execCall,
		"call.int":         execCall,
		"call.void":        execCall,
		"call.string":      execCall,
		"call.bool":        execCall,
		"call.char":        execCall,
		"iface.call":       execIfaceCall,
		"defer.push":       execDeferPush,
		"defer.push.func":  execDeferPushFunc,
		"defer.push.iface": execDeferPushIface,
		"defer.run":        execDeferRun,
		"struct.init":      execStructInit,
		"array.init":       execArrayInit,
		"slice.append":     execSliceAppend,
		"slice.slice":      execSliceSlice,
		"spawn":            execSpawn,
		"chan.make":        execChanMake,
		"chan.send":        execChanSend,
		"chan.recv":        execChanRecv,
		"chan.recv.ok":     execChanRecvOk,
		"chan.close":       execChanClose,
		"tuple.new":        execTupleNew,
		"tuple.extract":    execTupleExtract,
		"select":           execSelect,
		"index":            execIndex,
		"assign":           execAssign,
		"map.init":         execMapInit,
		"member":           execMember,
		"phi":              execPhi,
		"malloc":           execMalloc,
		"free":             execFree,
		"realloc":          execRealloc,
		"file.open":        execFileOpen,
		"file.close":       execFileClose,
		"file.read":        execFileRead,
		"file.write":       execFileWrite,
		"file.seek":        execFileSeek,
		"file.tell":        execFileTell,
		"file.exists":      execFileExists,
		"file.size":        execFileSize,
		"test.start":       execTestStart,
		"test.end":         execTestEnd,
		"assert":           execAssert,
		"assert.eq":        execAssertEq,
		"assert.true":      execAssertTrue,
		"assert.false":     execAssertFalse,
		"func.ref":         execFuncRef,
		"func.assign":      execFuncAssign,
		"func.call":        execFuncCall,
		"closure.create":   execClosureCreate,
		"closure.capture":  execClosureCapture,
		"closure.bind":     execClosureBind,
		"await":            execAwait,
	}
}

const inferTypePlaceholder = "<infer>"

// Result captures the outcome of executing the entry function.
type Result struct {
	Type  string
	Value interface{}
}

// Execute interprets the MIR module starting from the named entry function.
func Execute(mod *mir.Module, entry string) (res Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case exitSignal:
				res = Result{Type: "void"}
				err = ExitError{Code: v.code}
			case error:
				err = v
			default:
				panic(r)
			}
		}
	}()

	if mod == nil {
		return Result{}, fmt.Errorf("vm: nil module")
	}
	funcs := map[string]*mir.Function{}
	for _, fn := range mod.Functions {
		funcs[fn.Name] = fn
	}
	fn, ok := funcs[entry]
	if !ok {
		return Result{}, fmt.Errorf("vm: entry function %q not found", entry)
	}
	value, execErr := execFunction(funcs, fn, nil)
	if execErr != nil {
		return Result{}, execErr
	}
	return value, nil
}

type frame struct {
	values     map[mir.ValueID]Result
	deferStack []deferredCall // LIFO stack of pending defers
}

// deferredCall snapshots one enqueued defer. Three variants mirror the three
// MIR ops that enqueue them:
//
//	deferKindNamed: named function or intrinsic. `callee` is the MIR-mangled
//	  name; `recv` is unused.
//	deferKindFunc:  function-valued callee. `recv` holds the captured callable
//	  value (same encoding execFuncCall expects); `callee` is unused.
//	deferKindIface: method on an interface-typed receiver. `callee` holds the
//	  method name; `recv` holds the receiver whose Type tag drives runtime
//	  dispatch.
type deferredCall struct {
	kind   deferKind
	callee string
	recv   Result
	args   []Result
}

type deferKind int

const (
	deferKindNamed deferKind = iota
	deferKindFunc
	deferKindIface
)

func execFunction(funcs map[string]*mir.Function, fn *mir.Function, args []Result) (Result, error) {
	// Outer loop drives tail-call rebinding: when the inner block walker
	// detects a call in tail position (i.e. the result feeds directly into
	// `ret %callId`), it fills nextFn / nextArgs and breaks out of the
	// inner loop. We then reset the frame to the new fn+args and re-enter
	// without growing the Go stack. Both self-recursion and mutual
	// recursion benefit — the loop doesn't care whether the callee is the
	// same function or a different one.
	for {
		if len(args) != 0 && len(args) != len(fn.Params) {
			return Result{}, fmt.Errorf("vm: function %s expects %d arguments, got %d", fn.Name, len(fn.Params), len(args))
		}
		fr := &frame{values: make(map[mir.ValueID]Result)}
		for i, param := range fn.Params {
			var val Result
			if args != nil {
				val = args[i]
			} else {
				val = Result{Type: param.Type, Value: nil}
			}
			if val.Type == "" {
				val.Type = param.Type
			}
			fr.values[param.ID] = val
		}
		if len(fn.Blocks) == 0 {
			return Result{Type: "void"}, nil
		}
		blockMap := make(map[string]*mir.BasicBlock, len(fn.Blocks))
		for _, b := range fn.Blocks {
			blockMap[b.Name] = b
		}

		current := fn.Blocks[0]
		var nextFn *mir.Function
		var nextArgs []Result
		var ret Result
		var done bool
		var execErr error

	blockLoop:
		for {
			n := len(current.Instructions)
			for i, inst := range current.Instructions {
				// Tail-call detection: if this is the LAST instruction in
				// the block, it's a static-callee `call`, and the block's
				// terminator is `ret %thisInstID`, rebind (fn, args) and
				// re-enter the outer loop instead of recursing.
				if i == n-1 {
					if tFn, tArgs, ok := tailCallTarget(funcs, fr, current.Terminator, inst); ok {
						nextFn = tFn
						nextArgs = tArgs
						break blockLoop
					}
				}
				res, err := execInstruction(funcs, fr, inst)
				if err != nil {
					execErr = fmt.Errorf("vm: %s: %w", fn.Name, err)
					done = true
					break blockLoop
				}
				if inst.ID != mir.InvalidValue {
					fr.values[inst.ID] = res
				}
			}

			term := current.Terminator
			switch term.Op {
			case "ret":
				if len(term.Operands) == 0 {
					ret = Result{Type: "void"}
				} else {
					op := term.Operands[0]
					if op.Kind != mir.OperandValue {
						r, err := literalResult(op)
						if err != nil {
							execErr = err
						} else {
							ret = r
						}
					} else {
						ret = fr.values[op.Value]
					}
				}
				done = true
				break blockLoop
			case "br", "jmp":
				target, err := blockByOperand(blockMap, term.Operands[0])
				if err != nil {
					execErr = fmt.Errorf("vm: %s: %w", fn.Name, err)
					done = true
					break blockLoop
				}
				current = target
			case "cbr":
				if len(term.Operands) < 3 {
					execErr = fmt.Errorf("vm: %s: conditional branch requires condition and two targets", fn.Name)
					done = true
					break blockLoop
				}
				cond := operandValue(fr, term.Operands[0])
				b, err := toBool(cond)
				if err != nil {
					execErr = fmt.Errorf("vm: %s: %w", fn.Name, err)
					done = true
					break blockLoop
				}
				var targetOp mir.Operand
				if b {
					targetOp = term.Operands[1]
				} else {
					targetOp = term.Operands[2]
				}
				target, err := blockByOperand(blockMap, targetOp)
				if err != nil {
					execErr = fmt.Errorf("vm: %s: %w", fn.Name, err)
					done = true
					break blockLoop
				}
				current = target
			default:
				execErr = fmt.Errorf("unsupported terminator %q", term.Op)
				done = true
				break blockLoop
			}
		}

		if done {
			if execErr != nil {
				return Result{}, execErr
			}
			return ret, nil
		}
		// Tail-call rebind. Loop back to the top of execFunction with the
		// new function and arguments — no stack growth.
		fn = nextFn
		args = nextArgs
	}
}

// tailCallTarget returns (callee, args, true) if `inst` is a static-callee
// `call` whose result feeds directly into the block's `ret` terminator. The
// caller treats this as a tail call and rebinds the dispatch loop instead of
// recursing.
//
// Deliberately conservative: only plain "call" / "call.<type>" with a
// literal callee qualify. iface.call (dynamic dispatch via type tag),
// func.call (function-value receiver), intrinsics, and async/Promise calls
// fall through and execute normally — TCO for those is its own design
// problem.
// isStubBody returns true for a function whose body is the canonical
// "return literal" placeholder used by std.omni declarations whose real
// implementations live in execIntrinsic / the C runtime. Shape: a
// single block with one const instruction terminated by `ret <const>`.
// isVMIntrinsicOverride reports whether the VM has its own
// execIntrinsic case for a std.* callee. Used by the tail-call
// detector to avoid rebinding into the OmniLang body when the
// runtime path is the source of truth.
func isVMIntrinsicOverride(callee string) bool {
	switch callee {
	case "std.algorithms.bubble_sort",
		"std.algorithms.selection_sort",
		"std.algorithms.insertion_sort",
		"std.algorithms.linear_search",
		"std.algorithms.binary_search",
		"std.algorithms.find_max",
		"std.algorithms.find_min",
		"std.algorithms.count_occurrences",
		"std.algorithms.reverse",
		"std.algorithms.rotate",
		"std.algorithms.shuffle",
		"std.algorithms.unique",
		"std.algorithms.euclidean_distance",
		"std.math.random_seed",
		"std.math.random_int",
		"std.string.split",
		"std.string.split_lines",
		"std.string.split_words",
		"std.string.join",
		"std.string.replace",
		"std.string.replace_all",
		"std.string.replace_first",
		"std.string.replace_last",
		"std.string.find_all",
		"std.collections.size",
		"std.collections.get",
		"std.collections.set",
		"std.collections.has",
		"std.collections.remove",
		"std.collections.clear",
		"std.collections.queue_create",
		"std.collections.queue_enqueue",
		"std.collections.queue_dequeue",
		"std.collections.queue_peek",
		"std.collections.queue_is_empty",
		"std.collections.queue_size",
		"std.collections.queue_clear",
		"std.collections.stack_create",
		"std.collections.stack_push",
		"std.collections.stack_pop",
		"std.collections.stack_peek",
		"std.collections.stack_is_empty",
		"std.collections.stack_size",
		"std.collections.stack_clear",
		"std.collections.set_create",
		"std.collections.set_add",
		"std.collections.set_remove",
		"std.collections.set_contains",
		"std.collections.set_size",
		"std.collections.set_clear",
		"std.collections.set_union",
		"std.collections.set_intersection",
		"std.collections.set_difference",
		"std.algorithms.manhattan_distance",
		"std.algorithms.levenshtein_distance",
		"std.array.contains",
		"std.array.index_of",
		"std.array.append",
		"std.array.prepend",
		"std.array.insert",
		"std.array.remove",
		"std.array.concat",
		"std.array.slice",
		"std.file.open",
		"std.file.close",
		"std.file.read",
		"std.file.write",
		"std.file.seek",
		"std.file.tell",
		"std.file.exists",
		"std.file.size",
		"std.time.now",
		"std.time.unix_timestamp",
		"std.time.unix_nano",
		"std.time.sleep_seconds",
		"std.time.sleep_milliseconds",
		"std.time.time_zone_offset",
		"std.time.time_zone_name",
		"std.time.time_from_unix",
		"std.time.time_from_string",
		"std.time.time_to_unix",
		"std.time.time_to_string",
		"std.time.time_to_unix_nano",
		"std.time.duration_to_string":
		return true
	}
	return false
}

func isStubBody(fn *mir.Function) bool {
	if fn == nil || len(fn.Blocks) != 1 {
		return false
	}
	b := fn.Blocks[0]
	if len(b.Instructions) != 1 {
		return false
	}
	if b.Instructions[0].Op != "const" {
		return false
	}
	return b.Terminator.Op == "ret"
}

func tailCallTarget(funcs map[string]*mir.Function, fr *frame, term mir.Terminator, inst mir.Instruction) (*mir.Function, []Result, bool) {
	if term.Op != "ret" || len(term.Operands) != 1 {
		return nil, nil, false
	}
	retOp := term.Operands[0]
	if retOp.Kind != mir.OperandValue || retOp.Value != inst.ID {
		return nil, nil, false
	}
	switch inst.Op {
	case "call", "call.int", "call.string", "call.bool":
		// fine — value-producing call ops
	default:
		return nil, nil, false
	}
	if len(inst.Operands) == 0 || inst.Operands[0].Kind != mir.OperandLiteral {
		return nil, nil, false
	}
	callee := inst.Operands[0].Literal
	fn, ok := funcs[callee]
	if !ok {
		return nil, nil, false
	}
	// Skip tail-call rebind into a callee whose real implementation
	// lives in execIntrinsic. The user-side body is either a stub kept
	// around to satisfy the type checker (`return <literal>` shape) or
	// a partially-correct OmniLang implementation that depends on
	// other stubs — either way, going there bypasses the runtime
	// result we actually want.
	if strings.HasPrefix(callee, "std.") {
		if isStubBody(fn) || isVMIntrinsicOverride(callee) {
			return nil, nil, false
		}
	}
	// Only tail-call functions whose param count matches: a Promise<T>
	// callee would require unwrapping at the call site, which the normal
	// call path handles but the tail-rebind doesn't.
	if strings.HasPrefix(fn.ReturnType, "Promise<") {
		return nil, nil, false
	}
	if len(inst.Operands)-1 != len(fn.Params) {
		return nil, nil, false
	}
	args := make([]Result, 0, len(fn.Params))
	for _, op := range inst.Operands[1:] {
		args = append(args, operandValue(fr, op))
	}
	return fn, args, true
}

func blockByOperand(blocks map[string]*mir.BasicBlock, op mir.Operand) (*mir.BasicBlock, error) {
	if op.Kind != mir.OperandLiteral {
		return nil, fmt.Errorf("branch target must be literal block name")
	}
	block, ok := blocks[op.Literal]
	if !ok {
		return nil, fmt.Errorf("branch target %q not found", op.Literal)
	}
	return block, nil
}

func execInstruction(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	handler, exists := instructionHandlers[inst.Op]
	if !exists {
		return Result{}, fmt.Errorf("unsupported instruction %q", inst.Op)
	}
	return handler(funcs, fr, inst)
}

// execConst handles const instructions
func execConst(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	return literalResult(inst.Operands[0])
}

// execStructInit handles struct initialization
func execStructInit(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Create a struct value from the operands
	// Operands format: [field1Value, field2Value, ...] or [field1Name, field1Value, field2Name, field2Value, ...]
	structValue := make(map[string]interface{})

	// Handle different operand formats
	if len(inst.Operands) == 0 {
		// Empty struct
		return Result{Type: inst.Type, Value: structValue}, nil
	}

	// Check if first operand is a field name (literal) or field value
	if len(inst.Operands) >= 2 && inst.Operands[0].Kind == mir.OperandLiteral {
		// Named field format: [structType, field1Name, field1Value, field2Name, field2Value, ...]
		// or [field1Name, field1Value, field2Name, field2Value, ...]

		// Check if first operand is struct type (skip it) or if it's a field name
		startIndex := 0
		// If we have an odd number of operands, the first one might be a struct type
		// If we have an even number of operands, they should be field-value pairs
		if len(inst.Operands)%2 == 1 {
			// Odd number of operands - first is likely struct type
			startIndex = 1
		}

		// Check if remaining operands form field-value pairs
		remainingOperands := len(inst.Operands) - startIndex
		if remainingOperands%2 != 0 {
			return Result{}, fmt.Errorf("struct.init: named field format requires even number of operands")
		}

		for i := startIndex; i < len(inst.Operands); i += 2 {
			if i+1 >= len(inst.Operands) {
				return Result{}, fmt.Errorf("struct.init: incomplete field pair")
			}

			fieldNameOp := inst.Operands[i]
			fieldValueOp := inst.Operands[i+1]

			if fieldNameOp.Kind != mir.OperandLiteral {
				return Result{}, fmt.Errorf("struct.init: field name must be literal")
			}

			fieldValue := operandValue(fr, fieldValueOp)
			structValue[fieldNameOp.Literal] = fieldValue.Value
		}
	} else {
		// Positional field format: [field1Value, field2Value, ...]
		// Use generic field names
		for i, op := range inst.Operands {
			fieldValue := operandValue(fr, op)
			fieldName := fmt.Sprintf("field_%d", i)
			structValue[fieldName] = fieldValue.Value
		}
	}

	return Result{Type: inst.Type, Value: structValue}, nil
}

// execMember handles struct field access
func execMember(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("member: expected 2 operands, got %d", len(inst.Operands))
	}

	target := operandValue(fr, inst.Operands[0])
	fieldNameOp := inst.Operands[1]

	if fieldNameOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("member: field name must be literal")
	}

	fieldName := fieldNameOp.Literal

	// Handle struct field access
	if structValue, ok := target.Value.(map[string]interface{}); ok {
		if fieldValue, exists := structValue[fieldName]; exists {
			// Determine the field type based on the value
			var fieldType string
			switch fieldValue.(type) {
			case int:
				fieldType = "int"
			case string:
				fieldType = "string"
			case float64:
				fieldType = "float"
			case bool:
				fieldType = "bool"
			default:
				fieldType = "int" // Default fallback
			}
			return Result{Type: fieldType, Value: fieldValue}, nil
		}
		return Result{}, fmt.Errorf("member: field %q not found in struct", fieldName)
	}

	return Result{}, fmt.Errorf("member: target is not a struct")
}

// execPhi handles PHI nodes - values that can come from different control flow paths
func execPhi(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// PHI nodes have operands in pairs: (value, block_name)
	// For now, we'll use a simple approach: return the first value
	// In a more sophisticated implementation, we'd track which block we came from

	if len(inst.Operands) < 2 {
		return Result{}, fmt.Errorf("phi: expected at least 2 operands, got %d", len(inst.Operands))
	}

	if len(inst.Operands)%2 != 0 {
		return Result{}, fmt.Errorf("phi: expected even number of operands (value, block pairs), got %d", len(inst.Operands))
	}

	// For now, we'll return the first value
	// In a real implementation, we'd need to track the incoming control flow
	// and select the appropriate value based on which block we came from
	firstValue := operandValue(fr, inst.Operands[0])
	return firstValue, nil
}

// execSliceAppend handles `append(slice, elem)` by producing a new slice with
// elem appended. The VM carries arrays as typed Go slices (`[]int`,
// `[]string`, `[]float64`, `[]bool`, `[]interface{}` for nested/heterogeneous
// element types), so we dispatch on the underlying value type. The input
// slice is not mutated — Go's builtin append may return a new backing array
// when capacity runs out, and we return that directly.
func execSliceAppend(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("slice.append: expected (slice, elem) operands, got %d", len(inst.Operands))
	}
	slice := operandValue(fr, inst.Operands[0])
	elem := operandValue(fr, inst.Operands[1])

	switch s := slice.Value.(type) {
	case []int:
		n, ok := coerceToInt(elem.Value)
		if !ok {
			return Result{}, fmt.Errorf("slice.append: expected int element, got %T", elem.Value)
		}
		return Result{Type: slice.Type, Value: append(s, n)}, nil
	case []string:
		str, ok := elem.Value.(string)
		if !ok {
			return Result{}, fmt.Errorf("slice.append: expected string element, got %T", elem.Value)
		}
		return Result{Type: slice.Type, Value: append(s, str)}, nil
	case []float64:
		f, ok := coerceToFloat(elem.Value)
		if !ok {
			return Result{}, fmt.Errorf("slice.append: expected float element, got %T", elem.Value)
		}
		return Result{Type: slice.Type, Value: append(s, f)}, nil
	case []bool:
		b, ok := elem.Value.(bool)
		if !ok {
			return Result{}, fmt.Errorf("slice.append: expected bool element, got %T", elem.Value)
		}
		return Result{Type: slice.Type, Value: append(s, b)}, nil
	case []interface{}:
		return Result{Type: slice.Type, Value: append(s, elem.Value)}, nil
	case nil:
		// Appending to a nil slice — start a new []interface{} so subsequent
		// appends through the same binding keep working.
		return Result{Type: slice.Type, Value: []interface{}{elem.Value}}, nil
	default:
		return Result{}, fmt.Errorf("slice.append: unsupported slice backing of Go type %T", slice.Value)
	}
}

// execSpawn launches a function call as a new goroutine. Operand layout
// matches a regular call: [calleeLit | calleeValue, args...]. Argument values
// are captured by the spawning goroutine before the new one starts, mirroring
// Go's `go fn(args)` semantics.
func execSpawn(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("spawn: missing callee operand")
	}
	calleeOp := inst.Operands[0]
	args := make([]Result, 0, len(inst.Operands)-1)
	for _, op := range inst.Operands[1:] {
		args = append(args, operandValue(fr, op))
	}

	switch calleeOp.Kind {
	case mir.OperandLiteral:
		callee := calleeOp.Literal
		fn, ok := funcs[callee]
		if !ok {
			return Result{}, fmt.Errorf("spawn: callee %q not found", callee)
		}
		go func() {
			// We deliberately discard errors and the result: spawned tasks
			// have no direct return path. Programs route results through
			// channels.
			_, _ = execFunction(funcs, fn, args)
		}()
	case mir.OperandValue:
		callable := operandValue(fr, calleeOp)
		go func() {
			_ = invokeDeferredFunc(funcs, fr, callable, args)
		}()
	default:
		return Result{}, fmt.Errorf("spawn: unsupported callee operand kind")
	}
	return Result{Type: "void"}, nil
}

// execChanMake creates a buffered Go channel of `interface{}` typed slots.
// We don't keep per-element typing in the VM channel — the value travels
// through the frame's Result wrapper, which already carries its dynamic
// type tag.
func execChanMake(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("chan.make: missing element-type operand")
	}
	elemOp := inst.Operands[0]
	if elemOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("chan.make: element type must be a literal operand")
	}
	bufSize := 0
	if len(inst.Operands) >= 2 {
		capR := operandValue(fr, inst.Operands[1])
		if n, ok := coerceToInt(capR.Value); ok {
			if n < 0 {
				return Result{}, fmt.Errorf("chan.make: negative capacity %d", n)
			}
			bufSize = n
		}
	}
	ch := make(chan Result, bufSize)
	return Result{Type: inst.Type, Value: ch}, nil
}

// execChanSend pushes a value onto the channel. Blocks if the channel is
// full or unbuffered with no waiting receiver.
func execChanSend(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("chan.send: expected (chan, value) operands, got %d", len(inst.Operands))
	}
	chR := operandValue(fr, inst.Operands[0])
	val := operandValue(fr, inst.Operands[1])
	ch, ok := chR.Value.(chan Result)
	if !ok {
		return Result{}, fmt.Errorf("chan.send: target is not a channel (Go type %T)", chR.Value)
	}
	ch <- val
	return Result{Type: "void"}, nil
}

// execChanRecv pops a value off the channel. Blocks until one is available.
// Result Type is set to the channel's element type so downstream
// instructions see the right type tag.
func execChanRecv(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("chan.recv: expected (chan) operand, got %d", len(inst.Operands))
	}
	chR := operandValue(fr, inst.Operands[0])
	ch, ok := chR.Value.(chan Result)
	if !ok {
		return Result{}, fmt.Errorf("chan.recv: source is not a channel (Go type %T)", chR.Value)
	}
	val := <-ch
	if val.Type == "" {
		val.Type = inst.Type
	}
	return val, nil
}

// execChanRecvOk is the ok-form: `let v, ok = <-c` in source. Returns a
// Result whose Value is a []Result of length 2: [value, ok:bool]. ok is
// false exactly when the channel has been closed AND drained — matching Go
// semantics.
func execChanRecvOk(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("chan.recv.ok: expected (chan) operand, got %d", len(inst.Operands))
	}
	chR := operandValue(fr, inst.Operands[0])
	ch, ok := chR.Value.(chan Result)
	if !ok {
		return Result{}, fmt.Errorf("chan.recv.ok: source is not a channel (Go type %T)", chR.Value)
	}
	val, isOpen := <-ch
	elemType := inst.Type
	// inst.Type might be a tuple like "(int, bool)" — extract the element
	// type from the channel's own type for the inner value. We look it up
	// from the operand's channel type.
	if strings.HasPrefix(chR.Type, "chan<") && strings.HasSuffix(chR.Type, ">") {
		elemType = chR.Type[len("chan<") : len(chR.Type)-1]
	}
	if val.Type == "" {
		val.Type = elemType
	}
	okVal := Result{Type: "bool", Value: isOpen}
	return Result{Type: inst.Type, Value: []Result{val, okVal}}, nil
}

// execTupleNew packs operands into a Go []Result. The MIR builder emits
// this at `return a, b` and anywhere else multi-value results flow
// through a single SSA slot.
func execTupleNew(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	out := make([]Result, 0, len(inst.Operands))
	for _, op := range inst.Operands {
		out = append(out, operandValue(fr, op))
	}
	return Result{Type: inst.Type, Value: out}, nil
}

// execTupleExtract reads the Nth component out of a tuple Result produced
// by tuple.new or chan.recv.ok. Operand layout: [tupleValue, indexLit].
func execTupleExtract(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("tuple.extract: expected (tuple, index) operands, got %d", len(inst.Operands))
	}
	tuple := operandValue(fr, inst.Operands[0])
	idxOp := inst.Operands[1]
	if idxOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("tuple.extract: index must be a literal operand")
	}
	idx := 0
	if _, err := fmt.Sscanf(idxOp.Literal, "%d", &idx); err != nil {
		return Result{}, fmt.Errorf("tuple.extract: invalid index literal %q", idxOp.Literal)
	}
	members, ok := tuple.Value.([]Result)
	if !ok {
		return Result{}, fmt.Errorf("tuple.extract: source is not a tuple (Go type %T)", tuple.Value)
	}
	if idx < 0 || idx >= len(members) {
		return Result{}, fmt.Errorf("tuple.extract: index %d out of bounds for tuple of size %d", idx, len(members))
	}
	r := members[idx]
	if r.Type == "" {
		r.Type = inst.Type
	}
	return r, nil
}

// execSelect picks one ready case out of a set and performs its
// communication, binding recv destinations into the frame. Returns the
// chosen case's index as the instruction's result so the dispatcher
// after it can cbr to the matching body block.
//
// Operand layout matches the MIR builder's lowerSelect encoding:
//
//	operand[0]: literal count
//	Then 6 operands per case:
//	  [0] kind: "send" | "recv" | "recv.ok" | "default"
//	  [1] channel value (or "-" for default)
//	  [2] send value (send only)
//	  [3] recv-dest SSA-id literal "%N" (recv / recv.ok only)
//	  [4] recv-ok-dest SSA-id literal (recv.ok only)
//	  [5] body-block literal
func execSelect(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("select: missing case-count operand")
	}
	countOp := inst.Operands[0]
	if countOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("select: case count must be literal")
	}
	caseCount := 0
	if _, err := fmt.Sscanf(countOp.Literal, "%d", &caseCount); err != nil {
		return Result{}, fmt.Errorf("select: invalid case count %q", countOp.Literal)
	}
	if caseCount == 0 {
		return Result{}, fmt.Errorf("select: no cases")
	}
	if 1+caseCount*6 != len(inst.Operands) {
		return Result{}, fmt.Errorf("select: expected %d operands for %d cases, got %d", 1+caseCount*6, caseCount, len(inst.Operands))
	}

	type localCase struct {
		kind       string
		ch         chan Result
		sendValue  Result
		recvDestID mir.ValueID
		recvOkID   mir.ValueID
	}
	locals := make([]localCase, caseCount)
	parseSSAID := func(s string) (mir.ValueID, error) {
		if s == "-" {
			return mir.InvalidValue, nil
		}
		if !strings.HasPrefix(s, "%") {
			return 0, fmt.Errorf("expected SSA id literal starting with %%, got %q", s)
		}
		var n int
		if _, err := fmt.Sscanf(s[1:], "%d", &n); err != nil {
			return 0, err
		}
		return mir.ValueID(n), nil
	}
	base := 1
	for i := 0; i < caseCount; i++ {
		kindOp := inst.Operands[base+0]
		chOp := inst.Operands[base+1]
		sendOp := inst.Operands[base+2]
		recvDestOp := inst.Operands[base+3]
		recvOkOp := inst.Operands[base+4]
		// body-block operand is base+5; unused here, consumed by the
		// subsequent cbr dispatcher.
		base += 6
		locals[i].kind = kindOp.Literal
		if kindOp.Literal != "default" {
			chR := operandValue(fr, chOp)
			ch, ok := chR.Value.(chan Result)
			if !ok {
				return Result{}, fmt.Errorf("select case %d: operand is not a channel (Go type %T)", i, chR.Value)
			}
			locals[i].ch = ch
		}
		if kindOp.Literal == "send" {
			locals[i].sendValue = operandValue(fr, sendOp)
		}
		if id, err := parseSSAID(recvDestOp.Literal); err == nil {
			locals[i].recvDestID = id
		} else {
			locals[i].recvDestID = mir.InvalidValue
		}
		if id, err := parseSSAID(recvOkOp.Literal); err == nil {
			locals[i].recvOkID = id
		} else {
			locals[i].recvOkID = mir.InvalidValue
		}
	}

	// Build reflect.SelectCase list.
	selCases := make([]reflect.SelectCase, caseCount)
	for i, lc := range locals {
		switch lc.kind {
		case "send":
			selCases[i] = reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(lc.ch),
				Send: reflect.ValueOf(lc.sendValue),
			}
		case "recv", "recv.ok":
			selCases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(lc.ch),
			}
		case "default":
			selCases[i] = reflect.SelectCase{Dir: reflect.SelectDefault}
		default:
			return Result{}, fmt.Errorf("select: unknown case kind %q", lc.kind)
		}
	}

	chosen, recv, recvOK := reflect.Select(selCases)
	lc := locals[chosen]
	switch lc.kind {
	case "recv":
		if lc.recvDestID != mir.InvalidValue {
			if recvOK {
				fr.values[lc.recvDestID] = recv.Interface().(Result)
			} else {
				// Closed + drained: match Go's zero-of-T behavior.
				fr.values[lc.recvDestID] = Result{Type: "void", Value: nil}
			}
		}
	case "recv.ok":
		var val Result
		if recvOK {
			val = recv.Interface().(Result)
		} else {
			val = Result{Type: "void", Value: nil}
		}
		if lc.recvDestID != mir.InvalidValue {
			fr.values[lc.recvDestID] = val
		}
		if lc.recvOkID != mir.InvalidValue {
			fr.values[lc.recvOkID] = Result{Type: "bool", Value: recvOK}
		}
	}
	return Result{Type: "int", Value: chosen}, nil
}

// execChanClose closes the underlying Go channel so waiting receivers
// unblock and subsequent recv.ok returns (zero, false).
func execChanClose(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("chan.close: expected (chan) operand, got %d", len(inst.Operands))
	}
	chR := operandValue(fr, inst.Operands[0])
	ch, ok := chR.Value.(chan Result)
	if !ok {
		return Result{}, fmt.Errorf("chan.close: target is not a channel (Go type %T)", chR.Value)
	}
	// Guard against double-close: Go panics on re-close, so recover and
	// surface it as a VM-level error. Matches what a defensive program
	// expects.
	defer func() {
		_ = recover()
	}()
	close(ch)
	return Result{Type: "void"}, nil
}

// execSliceSlice handles `target[low:high]`. low and high are either Value
// operands (real integers in the frame) or Literal operands carrying "0" or
// "-1" to denote defaults. -1 means "len(target)".
func execSliceSlice(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 3 {
		return Result{}, fmt.Errorf("slice.slice: expected (target, low, high) operands, got %d", len(inst.Operands))
	}
	target := operandValue(fr, inst.Operands[0])
	lowR := operandValue(fr, inst.Operands[1])
	highR := operandValue(fr, inst.Operands[2])
	low, ok := coerceToInt(lowR.Value)
	if !ok {
		return Result{}, fmt.Errorf("slice.slice: low bound not an int (%T)", lowR.Value)
	}
	high, ok := coerceToInt(highR.Value)
	if !ok {
		return Result{}, fmt.Errorf("slice.slice: high bound not an int (%T)", highR.Value)
	}

	n := reflectLen(target.Value)
	if n < 0 {
		return Result{}, fmt.Errorf("slice.slice: target is not a slice (Go type %T)", target.Value)
	}
	if high == -1 {
		high = n
	}
	if low < 0 || high < low || high > n {
		return Result{}, fmt.Errorf("slice.slice: out-of-bounds [%d:%d] for length %d", low, high, n)
	}

	switch s := target.Value.(type) {
	case []int:
		return Result{Type: target.Type, Value: append([]int(nil), s[low:high]...)}, nil
	case []string:
		return Result{Type: target.Type, Value: append([]string(nil), s[low:high]...)}, nil
	case []float64:
		return Result{Type: target.Type, Value: append([]float64(nil), s[low:high]...)}, nil
	case []bool:
		return Result{Type: target.Type, Value: append([]bool(nil), s[low:high]...)}, nil
	case []interface{}:
		return Result{Type: target.Type, Value: append([]interface{}(nil), s[low:high]...)}, nil
	default:
		return Result{}, fmt.Errorf("slice.slice: unsupported slice backing of Go type %T", target.Value)
	}
}

// coerceToInt accepts int and common wider integer forms and unifies them
// to an int so the slice builtins can accept values from any typed context.
func coerceToInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	}
	return 0, false
}

// coerceToFloat mirrors coerceToInt for float-flavored slices.
func coerceToFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	}
	return 0, false
}

// reflectLen returns the length of any supported slice backing, or -1 if v
// is not a recognized slice.
func reflectLen(v interface{}) int {
	switch s := v.(type) {
	case []int:
		return len(s)
	case []string:
		return len(s)
	case []float64:
		return len(s)
	case []bool:
		return len(s)
	case []interface{}:
		return len(s)
	}
	return -1
}

// execArrayInit handles array initialization
func execArrayInit(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Extract element type from array type (e.g., "[]<int>" -> "int", "[]int" -> "int")
	elementType := inst.Type
	if strings.HasPrefix(elementType, "[]<") && strings.HasSuffix(elementType, ">") {
		elementType = elementType[3 : len(elementType)-1]
	} else if strings.HasPrefix(elementType, "array<") && strings.HasSuffix(elementType, ">") {
		elementType = elementType[6 : len(elementType)-1]
	} else if strings.HasPrefix(elementType, "[]") {
		elementType = elementType[2:]
	}

	// Handle nested array types like "[]<[]<int>>"
	if strings.HasPrefix(elementType, "[]<") && strings.HasSuffix(elementType, ">") {
		// This is a nested array type, handle it as a generic array
		elements := make([]interface{}, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			elements[i] = val.Value
		}
		return Result{Type: inst.Type, Value: elements}, nil
	}

	// Create array from operands
	var arrayValue interface{}
	switch elementType {
	case "int":
		elements := make([]int, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "int" {
				return Result{}, fmt.Errorf("array.init: expected int element, got %s", val.Type)
			}
			elements[i] = val.Value.(int)
		}
		arrayValue = elements
	case "string":
		elements := make([]string, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "string" {
				return Result{}, fmt.Errorf("array.init: expected string element, got %s", val.Type)
			}
			elements[i] = val.Value.(string)
		}
		arrayValue = elements
	case "float", "double":
		elements := make([]float64, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "float" && val.Type != "double" {
				return Result{}, fmt.Errorf("array.init: expected float element, got %s", val.Type)
			}
			elements[i] = val.Value.(float64)
		}
		arrayValue = elements
	case "bool":
		elements := make([]bool, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			if val.Type != "bool" {
				return Result{}, fmt.Errorf("array.init: expected bool element, got %s", val.Type)
			}
			elements[i] = val.Value.(bool)
		}
		arrayValue = elements
	default:
		// For unknown types, create a generic interface{} array
		elements := make([]interface{}, len(inst.Operands))
		for i, op := range inst.Operands {
			val := operandValue(fr, op)
			elements[i] = val.Value
		}
		arrayValue = elements
	}

	return Result{Type: inst.Type, Value: arrayValue}, nil
}

// execAssign handles variable assignment
func execAssign(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("assign: expected 2 operands, got %d", len(inst.Operands))
	}

	// Get the target variable ID (operand 0) and the source value (operand 1)
	targetOp := inst.Operands[0]
	if targetOp.Kind != mir.OperandValue {
		return Result{}, fmt.Errorf("assign: target must be a value, got %v", targetOp.Kind)
	}

	value := operandValue(fr, inst.Operands[1])

	// Update the target variable
	fr.values[targetOp.Value] = value

	// Also store in inst.ID for the result
	if inst.ID != mir.InvalidValue {
		fr.values[inst.ID] = value
	}

	return value, nil
}

// execIndex handles array/map indexing
func execIndex(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("index: expected 2 operands, got %d", len(inst.Operands))
	}

	target := operandValue(fr, inst.Operands[0])
	index := operandValue(fr, inst.Operands[1])

	// Handle map indexing first (no need to convert index to int)
	if strings.HasPrefix(target.Type, "map<") {
		mapValue, ok := target.Value.(map[interface{}]interface{})
		if !ok {
			return Result{}, fmt.Errorf("index: target is not a map")
		}

		// Get the value type from the map type
		_, valueType, err := parseMapTypes(target.Type)
		if err != nil {
			return Result{}, fmt.Errorf("index: invalid map type %s: %v", target.Type, err)
		}

		// Look up the key in the map
		keyValue := index.Value
		if foundValue, exists := mapValue[keyValue]; exists {
			// Ensure the found value has the correct type
			return Result{Type: valueType, Value: foundValue}, nil
		}

		// Key not found - return zero value for the value type
		switch valueType {
		case "int":
			return Result{Type: "int", Value: 0}, nil
		case "string":
			return Result{Type: "string", Value: ""}, nil
		case "float", "double":
			return Result{Type: "float", Value: 0.0}, nil
		case "bool":
			return Result{Type: "bool", Value: false}, nil
		default:
			// For unknown types, return nil
			return Result{Type: valueType, Value: nil}, nil
		}
	}

	// For arrays, convert index to int
	indexVal, err := toInt(index)
	if err != nil {
		return Result{}, fmt.Errorf("index: index must be int, got %s", index.Type)
	}

	// Handle array indexing
	if strings.HasPrefix(target.Type, "[]<") || strings.HasPrefix(target.Type, "array<") || strings.HasPrefix(target.Type, "[]") {
		switch arr := target.Value.(type) {
		case []int:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "int", Value: arr[indexVal]}, nil
		case []string:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "string", Value: arr[indexVal]}, nil
		case []float64:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "float", Value: arr[indexVal]}, nil
		case []bool:
			if indexVal < 0 || indexVal >= len(arr) {
				return Result{}, fmt.Errorf("index: array index %d out of bounds [0, %d)", indexVal, len(arr))
			}
			return Result{Type: "bool", Value: arr[indexVal]}, nil
		default:
			return Result{}, fmt.Errorf("index: unsupported array type %T", target.Value)
		}
	}

	return Result{}, fmt.Errorf("index: target type %s does not support indexing", target.Type)
}

// execMapInit handles map initialization
func execMapInit(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Extract key and value types from map type (e.g., "map<string,int>" -> "string", "int")
	keyType, valueType, err := parseMapTypes(inst.Type)
	if err != nil {
		return Result{}, fmt.Errorf("map.init: invalid map type %s: %v", inst.Type, err)
	}

	// Create a map from the operands (key-value pairs)
	mapValue := make(map[interface{}]interface{})

	// Process operands in pairs (key, value, key, value, ...)
	for i := 0; i < len(inst.Operands); i += 2 {
		if i+1 >= len(inst.Operands) {
			return Result{}, fmt.Errorf("map.init: odd number of operands, expected key-value pairs")
		}

		key := operandValue(fr, inst.Operands[i])
		value := operandValue(fr, inst.Operands[i+1])

		// Validate key and value types
		if key.Type != keyType {
			return Result{}, fmt.Errorf("map.init: key type mismatch, expected %s, got %s", keyType, key.Type)
		}
		if value.Type != valueType {
			return Result{}, fmt.Errorf("map.init: value type mismatch, expected %s, got %s", valueType, value.Type)
		}

		// Store in map
		mapValue[key.Value] = value.Value
	}

	return Result{Type: inst.Type, Value: mapValue}, nil
}

// parseMapTypes extracts key and value types from a map type string
func parseMapTypes(mapType string) (string, string, error) {
	if !strings.HasPrefix(mapType, "map<") || !strings.HasSuffix(mapType, ">") {
		return "", "", fmt.Errorf("invalid map type format")
	}

	inner := mapType[4 : len(mapType)-1] // Remove "map<" and ">"
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("map type must have exactly 2 type parameters")
	}

	keyType := strings.TrimSpace(parts[0])
	valueType := strings.TrimSpace(parts[1])

	return keyType, valueType, nil
}

func execArithmetic(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Handle float arithmetic
	if left.Type == "float" || right.Type == "float" || left.Type == "double" || right.Type == "double" {
		lf, err := toFloat(left)
		if err != nil {
			return Result{}, err
		}
		rf, err := toFloat(right)
		if err != nil {
			return Result{}, err
		}

		var res float64
		switch inst.Op {
		case "add":
			res = lf + rf
		case "sub":
			res = lf - rf
		case "mul":
			res = lf * rf
		case "div":
			if rf == 0 {
				return Result{}, fmt.Errorf("division by zero")
			}
			res = lf / rf
		case "mod":
			// For floats, use math.Mod
			if rf == 0 {
				return Result{}, fmt.Errorf("modulo by zero")
			}
			res = float64(int(lf) % int(rf)) // Simple modulo for floats
		default:
			return Result{}, fmt.Errorf("vm: unsupported float arithmetic operator %s", inst.Op)
		}
		return Result{Type: "float", Value: res}, nil
	}

	// Handle integer arithmetic (existing logic)
	li, err := toInt(left)
	if err != nil {
		return Result{}, err
	}
	ri, err := toInt(right)
	if err != nil {
		return Result{}, err
	}
	var res int
	switch inst.Op {
	case "add":
		res = li + ri
	case "sub":
		res = li - ri
	case "mul":
		res = li * ri
	case "div":
		if ri == 0 {
			return Result{}, fmt.Errorf("division by zero")
		}
		res = li / ri
	case "mod":
		if ri == 0 {
			return Result{}, fmt.Errorf("modulo by zero")
		}
		res = li % ri
	}
	return Result{Type: chooseType(inst.Type, left.Type), Value: res}, nil
}

func execComparison(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Handle different types of comparisons
	var res bool

	// String comparisons
	if left.Type == "string" && right.Type == "string" {
		leftStr, ok1 := left.Value.(string)
		rightStr, ok2 := right.Value.(string)
		if !ok1 || !ok2 {
			return Result{}, fmt.Errorf("vm: string comparison failed - invalid string values")
		}

		switch inst.Op {
		case "cmp.eq":
			res = leftStr == rightStr
		case "cmp.neq":
			res = leftStr != rightStr
		case "cmp.lt":
			res = leftStr < rightStr
		case "cmp.lte":
			res = leftStr <= rightStr
		case "cmp.gt":
			res = leftStr > rightStr
		case "cmp.gte":
			res = leftStr >= rightStr
		default:
			return Result{}, fmt.Errorf("vm: unsupported string comparison operator %s", inst.Op)
		}
		return Result{Type: "bool", Value: res}, nil
	}

	// Float comparisons
	if left.Type == "float" || right.Type == "float" || left.Type == "double" || right.Type == "double" {
		leftFloat, err := toFloat(left)
		if err != nil {
			return Result{}, err
		}
		rightFloat, err := toFloat(right)
		if err != nil {
			return Result{}, err
		}

		switch inst.Op {
		case "cmp.eq":
			res = leftFloat == rightFloat
		case "cmp.neq":
			res = leftFloat != rightFloat
		case "cmp.lt":
			res = leftFloat < rightFloat
		case "cmp.lte":
			res = leftFloat <= rightFloat
		case "cmp.gt":
			res = leftFloat > rightFloat
		case "cmp.gte":
			res = leftFloat >= rightFloat
		default:
			return Result{}, fmt.Errorf("vm: unsupported float comparison operator %s", inst.Op)
		}
		return Result{Type: "bool", Value: res}, nil
	}

	// Boolean comparisons
	if left.Type == "bool" && right.Type == "bool" {
		leftBool, ok1 := left.Value.(bool)
		rightBool, ok2 := right.Value.(bool)
		if !ok1 || !ok2 {
			return Result{}, fmt.Errorf("vm: boolean comparison failed - invalid boolean values")
		}

		switch inst.Op {
		case "cmp.eq":
			res = leftBool == rightBool
		case "cmp.neq":
			res = leftBool != rightBool
		default:
			return Result{}, fmt.Errorf("vm: unsupported boolean comparison operator %s", inst.Op)
		}
		return Result{Type: "bool", Value: res}, nil
	}

	// Integer comparisons (existing logic)
	li, err := toInt(left)
	if err != nil {
		return Result{}, err
	}
	ri, err := toInt(right)
	if err != nil {
		return Result{}, err
	}

	switch inst.Op {
	case "cmp.eq":
		res = li == ri
	case "cmp.neq":
		res = li != ri
	case "cmp.lt":
		res = li < ri
	case "cmp.lte":
		res = li <= ri
	case "cmp.gt":
		res = li > ri
	case "cmp.gte":
		res = li >= ri
	}
	return Result{Type: "bool", Value: res}, nil
}

func execLogical(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])
	lb, err := toBool(left)
	if err != nil {
		return Result{}, err
	}
	rb, err := toBool(right)
	if err != nil {
		return Result{}, err
	}
	var res bool
	if inst.Op == "and" {
		res = lb && rb
	} else {
		res = lb || rb
	}
	return Result{Type: "bool", Value: res}, nil
}

func execStringConcat(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Convert operands to strings
	leftStr, err := toString(left)
	if err != nil {
		return Result{}, fmt.Errorf("strcat: left operand: %w", err)
	}
	rightStr, err := toString(right)
	if err != nil {
		return Result{}, fmt.Errorf("strcat: right operand: %w", err)
	}

	// Use pooled strings.Builder for efficient concatenation
	builder := GetStringBuilder()
	defer PutStringBuilder(builder)

	builder.Grow(len(leftStr) + len(rightStr)) // Pre-allocate capacity
	builder.WriteString(leftStr)
	builder.WriteString(rightStr)

	return Result{Type: "string", Value: builder.String()}, nil
}

func execUnary(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	operand := operandValue(fr, inst.Operands[0])

	switch inst.Op {
	case "neg":
		if operand.Type == "int" {
			val := operand.Value.(int)
			return Result{Type: "int", Value: -val}, nil
		} else if operand.Type == "float" || operand.Type == "double" {
			val := operand.Value.(float64)
			return Result{Type: "float", Value: -val}, nil
		}
		return Result{}, fmt.Errorf("neg: operand must be int or float, got %s", operand.Type)
	case "not":
		if operand.Type == "bool" {
			val := operand.Value.(bool)
			return Result{Type: "bool", Value: !val}, nil
		}
		return Result{}, fmt.Errorf("not: operand must be bool, got %s", operand.Type)
	case "bitnot":
		if operand.Type == "int" {
			val := operand.Value.(int)
			return Result{Type: "int", Value: ^val}, nil
		}
		return Result{}, fmt.Errorf("bitnot: operand must be int, got %s", operand.Type)
	default:
		return Result{}, fmt.Errorf("unsupported unary operation %q", inst.Op)
	}
}

// execIfaceCall dispatches a method call on an interface-typed receiver by
// inspecting the receiver's dynamic (concrete) type at runtime, then forwarding
// to the concrete type's method via the normal call machinery.
//
// Operand layout: [ifaceNameLit, methodNameLit, receiverValue, args...].
func execIfaceCall(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 3 {
		return Result{}, fmt.Errorf("iface.call: expected at least ifaceName, methodName, receiver")
	}
	ifaceOp := inst.Operands[0]
	methodOp := inst.Operands[1]
	if ifaceOp.Kind != mir.OperandLiteral || methodOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("iface.call: ifaceName and methodName must be literal operands")
	}
	receiver := operandValue(fr, inst.Operands[2])
	concreteType := receiver.Type
	if concreteType == "" {
		return Result{}, fmt.Errorf("iface.call: receiver is missing a dynamic type tag (interface %s, method %s)", ifaceOp.Literal, methodOp.Literal)
	}
	callee := concreteType + "." + methodOp.Literal
	rewritten := inst
	rewritten.Op = "call"
	rewritten.Operands = append([]mir.Operand{{Kind: mir.OperandLiteral, Literal: callee}}, inst.Operands[2:]...)
	return execCall(funcs, fr, rewritten)
}

// execDeferPush snapshots a named-function defer onto the current frame. The
// MIR builder has already lowered the args as ordinary operands, so by the
// time we get here the argument values exist in the frame — we just read
// them and stash them verbatim. They will be consumed in LIFO order when the
// frame's defer.run instruction fires before `ret`.
func execDeferPush(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("defer.push: missing callee operand")
	}
	calleeOp := inst.Operands[0]
	if calleeOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("defer.push: callee must be literal")
	}
	args := make([]Result, 0, len(inst.Operands)-1)
	for _, op := range inst.Operands[1:] {
		args = append(args, operandValue(fr, op))
	}
	fr.deferStack = append(fr.deferStack, deferredCall{
		kind:   deferKindNamed,
		callee: calleeOp.Literal,
		args:   args,
	})
	return Result{Type: "void"}, nil
}

// execDeferPushFunc snapshots a function-valued defer. Operand layout:
// [funcValue, args...]. The callable value is captured by value just like
// any other arg — subsequent rebindings of the source variable do not
// affect what the defer will call.
func execDeferPushFunc(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("defer.push.func: missing callee value operand")
	}
	recv := operandValue(fr, inst.Operands[0])
	args := make([]Result, 0, len(inst.Operands)-1)
	for _, op := range inst.Operands[1:] {
		args = append(args, operandValue(fr, op))
	}
	fr.deferStack = append(fr.deferStack, deferredCall{
		kind: deferKindFunc,
		recv: recv,
		args: args,
	})
	return Result{Type: "void"}, nil
}

// execDeferPushIface snapshots an interface-dispatched defer. Operand layout:
// [ifaceLit, methodLit, receiver, args...]. The ifaceLit is informational
// (debug only); dispatch is driven entirely by the receiver's runtime type
// tag at defer.run time.
func execDeferPushIface(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 3 {
		return Result{}, fmt.Errorf("defer.push.iface: expected at least ifaceName, methodName, receiver")
	}
	methodOp := inst.Operands[1]
	if methodOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("defer.push.iface: methodName must be literal")
	}
	recv := operandValue(fr, inst.Operands[2])
	args := make([]Result, 0, len(inst.Operands)-3)
	for _, op := range inst.Operands[3:] {
		args = append(args, operandValue(fr, op))
	}
	fr.deferStack = append(fr.deferStack, deferredCall{
		kind:   deferKindIface,
		callee: methodOp.Literal,
		recv:   recv,
		args:   args,
	})
	return Result{Type: "void"}, nil
}

// execDeferRun flushes the current frame's deferred calls in LIFO order,
// invoking each via the normal execFunction machinery. Any value they return
// is discarded: deferred calls cannot contribute to the enclosing function's
// result (matching Go semantics).
func execDeferRun(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	// Walk the stack in reverse and clear it before dispatching so that a
	// deferred call which itself defers on the same frame (unusual but legal)
	// doesn't double-run the existing entries.
	pending := fr.deferStack
	fr.deferStack = nil
	for i := len(pending) - 1; i >= 0; i-- {
		d := pending[i]
		switch d.kind {
		case deferKindNamed:
			if _, handled := execIntrinsicWithArgs(d.callee, d.args, fr); handled {
				continue
			}
			fn, ok := funcs[d.callee]
			if !ok {
				return Result{}, fmt.Errorf("defer.run: callee %q not found", d.callee)
			}
			if _, err := execFunction(funcs, fn, d.args); err != nil {
				return Result{}, err
			}
		case deferKindFunc:
			if err := invokeDeferredFunc(funcs, fr, d.recv, d.args); err != nil {
				return Result{}, err
			}
		case deferKindIface:
			concreteType := d.recv.Type
			if concreteType == "" {
				return Result{}, fmt.Errorf("defer.run (iface): receiver missing dynamic type tag for method %q", d.callee)
			}
			fullName := concreteType + "." + d.callee
			fn, ok := funcs[fullName]
			if !ok {
				return Result{}, fmt.Errorf("defer.run (iface): method %q not found on type %q", d.callee, concreteType)
			}
			callArgs := append([]Result{d.recv}, d.args...)
			if _, err := execFunction(funcs, fn, callArgs); err != nil {
				return Result{}, err
			}
		default:
			return Result{}, fmt.Errorf("defer.run: unknown defer kind %d", d.kind)
		}
	}
	return Result{Type: "void"}, nil
}

// invokeDeferredFunc runs a deferred function-valued call. It mirrors
// execFuncCall's closure vs plain-func dispatch, but with pre-evaluated args
// so there's no operand-lookup step.
func invokeDeferredFunc(funcs map[string]*mir.Function, fr *frame, callable Result, args []Result) error {
	if closure, ok := callable.Value.(map[string]interface{}); ok {
		fnName, ok := closure["function"].(string)
		if !ok {
			return fmt.Errorf("defer.run (func): closure missing function name")
		}
		fn, ok := funcs[fnName]
		if !ok {
			return fmt.Errorf("defer.run (func): closure target %q not found", fnName)
		}
		merged := append([]Result{}, args...)
		if captured, ok := closure["captured"].(map[string]interface{}); ok {
			for _, v := range captured {
				merged = append(merged, Result{Type: "int", Value: v})
			}
		}
		_, err := execFunction(funcs, fn, merged)
		return err
	}
	if name, ok := callable.Value.(string); ok {
		fn, ok := funcs[name]
		if !ok {
			return fmt.Errorf("defer.run (func): named target %q not found", name)
		}
		_, err := execFunction(funcs, fn, args)
		return err
	}
	return fmt.Errorf("defer.run (func): unsupported callable value of Go type %T", callable.Value)
}

// execIntrinsicWithArgs is a thin adapter that lets deferred calls to
// builtin / stdlib intrinsics reuse the existing intrinsic dispatcher. It
// converts pre-evaluated args into temporary literal operands — the cheapest
// way to keep the intrinsic table as the single source of truth for builtin
// behavior without duplicating its call surface.
func execIntrinsicWithArgs(callee string, args []Result, fr *frame) (Result, bool) {
	ops := make([]mir.Operand, 0, len(args))
	// Stash each arg in the frame under a transient negative ID so the
	// intrinsic can read it through operandValue. Negative IDs are outside
	// the range NextValue() issues, so they can't collide with real SSA
	// values.
	base := mir.ValueID(-1000)
	for i, a := range args {
		id := base - mir.ValueID(i)
		fr.values[id] = a
		ops = append(ops, mir.Operand{Kind: mir.OperandValue, Value: id, Type: a.Type})
	}
	res, handled := execIntrinsic(callee, ops, fr)
	for i := range args {
		delete(fr.values, base-mir.ValueID(i))
	}
	return res, handled
}

func execCall(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) == 0 {
		return Result{}, fmt.Errorf("call instruction missing callee operand")
	}
	calleeOp := inst.Operands[0]
	if calleeOp.Kind != mir.OperandLiteral {
		return Result{}, fmt.Errorf("call expects literal callee operand")
	}
	callee := calleeOp.Literal

	// Check if it's an intrinsic function
	if result, handled := execIntrinsic(callee, inst.Operands[1:], fr); handled {
		return result, nil
	}

	fn, ok := funcs[callee]
	if !ok {
		// Check if it's an imported module function (contains a dot but not std.*)
		if strings.Contains(callee, ".") && !strings.HasPrefix(callee, "std.") {
			// For now, return a default value for imported functions
			// In a real implementation, this would need to load and execute the imported module
			return Result{Type: "int", Value: 0}, nil
		}
		// Record coverage for std library functions that aren't intrinsics
		if strings.HasPrefix(callee, "std.") {
			recordCoverage(callee, "", 0)
		}
		return Result{}, fmt.Errorf("callee %q not found", callee)
	}

	// Record coverage for std library functions
	if strings.HasPrefix(callee, "std.") {
		recordCoverage(callee, "", 0)
	}
	args := make([]Result, 0, len(fn.Params))
	for _, op := range inst.Operands[1:] {
		args = append(args, operandValue(fr, op))
	}

	// Check if function returns a Promise (async function)
	if strings.HasPrefix(fn.ReturnType, "Promise<") {
		// Execute async function and return a Promise
		promiseID := newPromise()

		// Execute function asynchronously
		go func() {
			result, err := execFunction(funcs, fn, args)
			if err != nil {
				rejectPromise(promiseID, err)
			} else {
				resolvePromise(promiseID, result)
			}
		}()

		// Return Promise immediately
		return Result{
			Type:  "Promise",
			Value: promiseID,
		}, nil
	}

	// Synchronous execution for non-async functions
	return execFunction(funcs, fn, args)
}

// recordCoverage records a function call for coverage tracking
func recordCoverage(functionName, filePath string, lineNumber int) {
	coverageMu.Lock()
	defer coverageMu.Unlock()

	if !coverageEnabled {
		return
	}

	key := fmt.Sprintf("%s:%s:%d", functionName, filePath, lineNumber)
	if entry, exists := coverageData[key]; exists {
		entry.CallCount++
	} else {
		coverageData[key] = &coverageEntry{
			FunctionName: functionName,
			FilePath:     filePath,
			LineNumber:   lineNumber,
			CallCount:    1,
		}
	}
}

// SetCoverageEnabled enables or disables coverage tracking
func SetCoverageEnabled(enabled bool) {
	coverageMu.Lock()
	defer coverageMu.Unlock()
	coverageEnabled = enabled
}

// IsCoverageEnabled returns whether coverage tracking is enabled
func IsCoverageEnabled() bool {
	coverageMu.RLock()
	defer coverageMu.RUnlock()
	return coverageEnabled
}

// ResetCoverage clears all coverage data
func ResetCoverage() {
	coverageMu.Lock()
	defer coverageMu.Unlock()
	coverageData = make(map[string]*coverageEntry)
}

// ExportCoverage exports coverage data as JSON
func ExportCoverage() ([]byte, error) {
	coverageMu.RLock()
	defer coverageMu.RUnlock()

	entries := make([]*coverageEntry, 0, len(coverageData))
	for _, entry := range coverageData {
		entries = append(entries, entry)
	}

	data := map[string]interface{}{
		"entries": entries,
	}

	return json.Marshal(data)
}

// execIntrinsic handles intrinsic function calls like std.io.println.
// Returns (result, handled) where handled indicates if the function was an intrinsic.
func execIntrinsic(callee string, operands []mir.Operand, fr *frame) (Result, bool) {
	// Record coverage for std library functions
	if strings.HasPrefix(callee, "std.") {
		recordCoverage(callee, "", 0) // File path and line number not available in VM
	}

	switch callee {
	case "len":
		// Builtin len function for arrays
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			// Check if it's an array type
			if strings.HasPrefix(arg.Type, "[]<") || strings.HasPrefix(arg.Type, "array<") {
				switch arr := arg.Value.(type) {
				case []int:
					return Result{Type: "int", Value: len(arr)}, true
				case []string:
					return Result{Type: "int", Value: len(arr)}, true
				case []float64:
					return Result{Type: "int", Value: len(arr)}, true
				case []bool:
					return Result{Type: "int", Value: len(arr)}, true
				default:
					return Result{}, false
				}
			}
		}
		return Result{}, false
	case "std.io.print":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.println":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Println(arg.Value)
			return Result{Type: "void", Value: nil}, true
		} else if len(operands) == 0 {
			fmt.Println()
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.eprint":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Fprint(os.Stderr, arg.Value)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.eprintln":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Fprintln(os.Stderr, arg.Value)
			return Result{Type: "void", Value: nil}, true
		} else if len(operands) == 0 {
			fmt.Fprintln(os.Stderr)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.flush":
		os.Stdout.Sync()
		return Result{Type: "void", Value: nil}, true
	case "std.io.is_terminal":
		stat, err := os.Stdout.Stat()
		if err != nil {
			return Result{Type: "bool", Value: false}, true
		}
		isTTY := (stat.Mode() & os.ModeCharDevice) != 0
		return Result{Type: "bool", Value: isTTY}, true
	case "std.io.sprintf":
		if len(operands) >= 2 {
			fmtArg := operandValue(fr, operands[0])
			argsArg := operandValue(fr, operands[1])
			format, _ := fmtArg.Value.(string)
			argsSlice, _ := argsArg.Value.([]string)
			if argsSlice == nil {
				if as, ok := argsArg.Value.([]interface{}); ok {
					for _, a := range as {
						if s, ok := a.(string); ok {
							argsSlice = append(argsSlice, s)
						} else {
							argsSlice = append(argsSlice, fmt.Sprint(a))
						}
					}
				}
			}
			result := vmSprintf(format, argsSlice)
			return Result{Type: "string", Value: result}, true
		}
	case "std.io.read_line":
		line, err := readLineFromStdin()
		if err != nil && !errors.Is(err, io.EOF) {
			return Result{Type: "string", Value: ""}, true
		}
		return Result{Type: "string", Value: line}, true
	case "std.io.read_all":
		data, err := readAllFromStdin()
		if err != nil && !errors.Is(err, io.EOF) {
			return Result{Type: "string", Value: ""}, true
		}
		return Result{Type: "string", Value: data}, true
	case "std.io.sprint":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			return Result{Type: "string", Value: fmt.Sprint(arg.Value)}, true
		}
	case "std.io.sprintln":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			return Result{Type: "string", Value: fmt.Sprint(arg.Value) + "\n"}, true
		}
	case "std.io.prompt":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			os.Stdout.Sync()
			line, err := readLineFromStdin()
			if err != nil && !errors.Is(err, io.EOF) {
				return Result{Type: "string", Value: ""}, true
			}
			return Result{Type: "string", Value: line}, true
		}
	case "std.io.read_lines":
		data, err := readAllFromStdin()
		if err != nil && !errors.Is(err, io.EOF) {
			return Result{Type: "array<string>", Value: []string{}}, true
		}
		if data == "" {
			return Result{Type: "array<string>", Value: []string{}}, true
		}
		// Split on \n; drop the final empty element if the input ended in \n.
		lines := strings.Split(data, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		return Result{Type: "array<string>", Value: lines}, true
	case "std.io.read_int":
		line, err := readLineFromStdin()
		if err != nil && !errors.Is(err, io.EOF) {
			return Result{Type: "int", Value: int64(0)}, true
		}
		n, ok := parseIntStrict(line)
		if !ok {
			n = 0
		}
		return Result{Type: "int", Value: int64(n)}, true
	case "std.io.read_float":
		line, err := readLineFromStdin()
		if err != nil && !errors.Is(err, io.EOF) {
			return Result{Type: "float", Value: 0.0}, true
		}
		f, ok := parseFloatStrict(line)
		if !ok {
			f = 0.0
		}
		return Result{Type: "float", Value: f}, true
	case "std.io.parse_int":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			s, _ := arg.Value.(string)
			n, ok := parseIntStrict(s)
			if !ok {
				n = 0
			}
			return Result{Type: "int", Value: int64(n)}, true
		}
	case "std.io.parse_float":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			s, _ := arg.Value.(string)
			f, ok := parseFloatStrict(s)
			if !ok {
				f = 0.0
			}
			return Result{Type: "float", Value: f}, true
		}
	case "std.io.is_int":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			s, _ := arg.Value.(string)
			_, ok := parseIntStrict(s)
			return Result{Type: "bool", Value: ok}, true
		}
	case "std.io.is_float":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			s, _ := arg.Value.(string)
			_, ok := parseFloatStrict(s)
			return Result{Type: "bool", Value: ok}, true
		}
	case "std.io.printf":
		if len(operands) >= 2 {
			format, _ := operandValue(fr, operands[0]).Value.(string)
			args := vmStringSlice(operandValue(fr, operands[1]))
			fmt.Print(vmSprintf(format, args))
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.eprintf":
		if len(operands) >= 2 {
			format, _ := operandValue(fr, operands[0]).Value.(string)
			args := vmStringSlice(operandValue(fr, operands[1]))
			fmt.Fprint(os.Stderr, vmSprintf(format, args))
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.print_each":
		if len(operands) == 1 {
			items := vmStringSlice(operandValue(fr, operands[0]))
			for _, item := range items {
				fmt.Println(item)
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.eprint_each":
		if len(operands) == 1 {
			items := vmStringSlice(operandValue(fr, operands[0]))
			for _, item := range items {
				fmt.Fprintln(os.Stderr, item)
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.eprompt":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Fprint(os.Stderr, arg.Value)
			os.Stderr.Sync()
			line, err := readLineFromStdin()
			if err != nil && !errors.Is(err, io.EOF) {
				return Result{Type: "string", Value: ""}, true
			}
			return Result{Type: "string", Value: line}, true
		}
	case "std.io.confirm":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			fmt.Print(arg.Value)
			os.Stdout.Sync()
			line, err := readLineFromStdin()
			yes := false
			if err == nil || errors.Is(err, io.EOF) {
				if len(line) > 0 && (line[0] == 'y' || line[0] == 'Y') {
					yes = true
				}
			}
			return Result{Type: "bool", Value: yes}, true
		}
	case "std.io.flush_stderr":
		os.Stderr.Sync()
		return Result{Type: "void", Value: nil}, true
	case "std.io.style":
		if len(operands) == 2 {
			s, _ := operandValue(fr, operands[0]).Value.(string)
			code, _ := operandValue(fr, operands[1]).Value.(string)
			return Result{Type: "string", Value: ansiWrap(s, code)}, true
		}
	case "std.io.bold":
		return ansiIntrinsic(fr, operands, "1")
	case "std.io.dim":
		return ansiIntrinsic(fr, operands, "2")
	case "std.io.italic":
		return ansiIntrinsic(fr, operands, "3")
	case "std.io.underline":
		return ansiIntrinsic(fr, operands, "4")
	case "std.io.red":
		return ansiIntrinsic(fr, operands, "31")
	case "std.io.green":
		return ansiIntrinsic(fr, operands, "32")
	case "std.io.yellow":
		return ansiIntrinsic(fr, operands, "33")
	case "std.io.blue":
		return ansiIntrinsic(fr, operands, "34")
	case "std.io.magenta":
		return ansiIntrinsic(fr, operands, "35")
	case "std.io.cyan":
		return ansiIntrinsic(fr, operands, "36")
	case "std.io.fprint":
		if len(operands) >= 2 {
			handle, _ := vmIntValue(operandValue(fr, operands[0]))
			arg := operandValue(fr, operands[1])
			vmFprint(handle, fmt.Sprint(arg.Value), false)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.fprintln":
		if len(operands) >= 2 {
			handle, _ := vmIntValue(operandValue(fr, operands[0]))
			arg := operandValue(fr, operands[1])
			vmFprint(handle, fmt.Sprint(arg.Value), true)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.io.fprintf":
		if len(operands) >= 3 {
			handle, _ := vmIntValue(operandValue(fr, operands[0]))
			format, _ := operandValue(fr, operands[1]).Value.(string)
			args := vmStringSlice(operandValue(fr, operands[2]))
			vmFprint(handle, vmSprintf(format, args), false)
			return Result{Type: "void", Value: nil}, true
		}
	case "std.file.read_all":
		if len(operands) == 1 {
			handle, _ := vmIntValue(operandValue(fr, operands[0]))
			return Result{Type: "string", Value: vmFileReadAll(handle)}, true
		}
	case "std.file.read_line":
		if len(operands) == 1 {
			handle, _ := vmIntValue(operandValue(fr, operands[0]))
			return Result{Type: "string", Value: vmFileReadLine(handle)}, true
		}
	case "std.file.write_string":
		if len(operands) == 2 {
			handle, _ := vmIntValue(operandValue(fr, operands[0]))
			s, _ := operandValue(fr, operands[1]).Value.(string)
			n := vmFileWriteString(handle, s)
			return Result{Type: "int", Value: n}, true
		}
	case "std.os.read_file_lines":
		if len(operands) == 1 {
			path, _ := operandValue(fr, operands[0]).Value.(string)
			data, err := vmReadFile(path)
			if err != nil {
				return Result{Type: "array<string>", Value: []string{}}, true
			}
			lines := strings.Split(data, "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}
			return Result{Type: "array<string>", Value: lines}, true
		}
	case "std.os.write_file_lines":
		if len(operands) == 2 {
			path, _ := operandValue(fr, operands[0]).Value.(string)
			lines := vmStringSlice(operandValue(fr, operands[1]))
			ok := vmWriteFileLines(path, lines)
			return Result{Type: "bool", Value: ok}, true
		}
	case "std.io.read_line_async":
		// Async version - execute in goroutine and return Promise
		promiseID := newPromise()
		go func() {
			line, err := readLineFromStdin()
			if err != nil && !errors.Is(err, io.EOF) {
				rejectPromise(promiseID, err)
			} else {
				resolvePromise(promiseID, Result{Type: "string", Value: line})
			}
		}()
		return Result{Type: "Promise", Value: promiseID}, true
	case "std.log.debug":
		return handleLogIntrinsic("debug", operands, fr)
	case "std.log.info":
		return handleLogIntrinsic("info", operands, fr)
	case "std.log.warn":
		return handleLogIntrinsic("warn", operands, fr)
	case "std.log.error":
		return handleLogIntrinsic("error", operands, fr)
	case "std.log.set_level":
		if len(operands) == 1 {
			levelVal := operandValue(fr, operands[0])
			levelStr, err := toString(levelVal)
			if err != nil {
				levelStr = fmt.Sprint(levelVal.Value)
			}
			success := logging.SetLevelByName(strings.TrimSpace(levelStr))
			return Result{Type: "bool", Value: success}, true
		}
		return Result{Type: "bool", Value: false}, true
	case "std.math.max":
		if len(operands) == 2 {
			left := operandValue(fr, operands[0])
			right := operandValue(fr, operands[1])
			if left.Type == "int" && right.Type == "int" {
				leftVal := left.Value.(int)
				rightVal := right.Value.(int)
				if leftVal > rightVal {
					return Result{Type: "int", Value: leftVal}, true
				}
				return Result{Type: "int", Value: rightVal}, true
			}
		}
	case "std.math.min":
		if len(operands) == 2 {
			left := operandValue(fr, operands[0])
			right := operandValue(fr, operands[1])
			if left.Type == "int" && right.Type == "int" {
				leftVal := left.Value.(int)
				rightVal := right.Value.(int)
				if leftVal < rightVal {
					return Result{Type: "int", Value: leftVal}, true
				}
				return Result{Type: "int", Value: rightVal}, true
			}
		}
	case "std.math.abs":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			if arg.Type == "int" {
				val := arg.Value.(int)
				if val < 0 {
					val = -val
				}
				return Result{Type: "int", Value: val}, true
			}
		}
	case "std.math.pow":
		if len(operands) == 2 {
			base := operandValue(fr, operands[0])
			exp := operandValue(fr, operands[1])
			if base.Type == "int" && exp.Type == "int" {
				baseVal := base.Value.(int)
				expVal := exp.Value.(int)
				result := 1
				for i := 0; i < expVal; i++ {
					result *= baseVal
				}
				return Result{Type: "int", Value: result}, true
			}
		}
	case "std.math.gcd":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "int" && b.Type == "int" {
				aVal := a.Value.(int)
				bVal := b.Value.(int)
				if aVal < 0 {
					aVal = -aVal
				}
				if bVal < 0 {
					bVal = -bVal
				}
				for bVal != 0 {
					aVal, bVal = bVal, aVal%bVal
				}
				return Result{Type: "int", Value: aVal}, true
			}
		}
	case "std.math.lcm":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "int" && b.Type == "int" {
				aVal := a.Value.(int)
				bVal := b.Value.(int)
				if aVal < 0 {
					aVal = -aVal
				}
				if bVal < 0 {
					bVal = -bVal
				}
				gcd := aVal
				temp := bVal
				for temp != 0 {
					gcd, temp = temp, gcd%temp
				}
				if gcd == 0 {
					return Result{Type: "int", Value: 0}, true
				}
				lcm := (aVal * bVal) / gcd
				return Result{Type: "int", Value: lcm}, true
			}
		}
	case "std.math.factorial":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				if nVal < 0 {
					return Result{Type: "int", Value: 0}, true
				}
				result := 1
				for i := 1; i <= nVal; i++ {
					result *= i
				}
				return Result{Type: "int", Value: result}, true
			}
		}
	case "std.math.sqrt":
		if len(operands) == 1 {
			x := operandValue(fr, operands[0])
			if x.Type == "int" {
				xVal := x.Value.(int)
				if xVal < 0 {
					return Result{Type: "int", Value: 0}, true
				}
				if xVal == 0 {
					return Result{Type: "int", Value: 0}, true
				}
				result := 1
				for result*result <= xVal {
					result++
				}
				return Result{Type: "int", Value: result - 1}, true
			}
		}
	case "std.math.is_prime":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				if nVal < 2 {
					return Result{Type: "bool", Value: false}, true
				}
				if nVal == 2 {
					return Result{Type: "bool", Value: true}, true
				}
				if nVal%2 == 0 {
					return Result{Type: "bool", Value: false}, true
				}
				for i := 3; i*i <= nVal; i += 2 {
					if nVal%i == 0 {
						return Result{Type: "bool", Value: false}, true
					}
				}
				return Result{Type: "bool", Value: true}, true
			}
		}
	case "std.math.max_float":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "float" && b.Type == "float" {
				aVal := a.Value.(float64)
				bVal := b.Value.(float64)
				if aVal > bVal {
					return Result{Type: "float", Value: aVal}, true
				}
				return Result{Type: "float", Value: bVal}, true
			}
		}
	case "std.math.min_float":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "float" && b.Type == "float" {
				aVal := a.Value.(float64)
				bVal := b.Value.(float64)
				if aVal < bVal {
					return Result{Type: "float", Value: aVal}, true
				}
				return Result{Type: "float", Value: bVal}, true
			}
		}
	case "std.math.abs_float":
		if len(operands) == 1 {
			x := operandValue(fr, operands[0])
			if x.Type == "float" {
				xVal := x.Value.(float64)
				if xVal < 0 {
					return Result{Type: "float", Value: -xVal}, true
				}
				return Result{Type: "float", Value: xVal}, true
			}
		}
	case "std.math.toString":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				return Result{Type: "string", Value: fmt.Sprintf("%d", nVal)}, true
			}
		}
	case "std.int_to_string":
		if len(operands) == 1 {
			n := operandValue(fr, operands[0])
			if n.Type == "int" {
				nVal := n.Value.(int)
				return Result{Type: "string", Value: fmt.Sprintf("%d", nVal)}, true
			}
		}
	case "std.bool_to_string":
		if len(operands) == 1 {
			b := operandValue(fr, operands[0])
			if b.Type == "bool" {
				bVal := b.Value.(bool)
				if bVal {
					return Result{Type: "string", Value: "true"}, true
				}
				return Result{Type: "string", Value: "false"}, true
			}
		}
	case "std.float_to_string":
		if len(operands) == 1 {
			f := operandValue(fr, operands[0])
			if f.Type == "float" || f.Type == "double" {
				fVal := f.Value.(float64)
				return Result{Type: "string", Value: fmt.Sprintf("%g", fVal)}, true
			}
		}
	case "std.string_to_int":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				sVal := s.Value.(string)
				if i, err := strconv.Atoi(sVal); err == nil {
					return Result{Type: "int", Value: int32(i)}, true
				}
			}
		}
	case "std.string_to_float":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				sVal := s.Value.(string)
				if f, err := strconv.ParseFloat(sVal, 64); err == nil {
					return Result{Type: "float", Value: f}, true
				}
			}
		}
	case "std.string_to_bool":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				sVal := s.Value.(string)
				if sVal == "true" {
					return Result{Type: "bool", Value: true}, true
				} else if sVal == "false" {
					return Result{Type: "bool", Value: false}, true
				}
			}
		}
	case "std.char_code":
		if len(operands) == 1 {
			c := operandValue(fr, operands[0])
			// Char values come in as rune (alias for int32). Other
			// integer carrier types are accepted defensively. Result is
			// returned as Go `int` because that's what the rest of the
			// VM's arithmetic ops produce (literalResult for int yields
			// int, not int32).
			switch v := c.Value.(type) {
			case int32:
				return Result{Type: "int", Value: int(v)}, true
			case int:
				return Result{Type: "int", Value: v}, true
			}
		}
	case "std.char_from_code":
		if len(operands) == 1 {
			i := operandValue(fr, operands[0])
			switch v := i.Value.(type) {
			case int:
				return Result{Type: "char", Value: rune(v)}, true
			case int32:
				return Result{Type: "char", Value: rune(v)}, true
			}
		}
	case "std.char_to_string":
		if len(operands) == 1 {
			c := operandValue(fr, operands[0])
			switch v := c.Value.(type) {
			case int32:
				return Result{Type: "string", Value: string(rune(v))}, true
			case int:
				return Result{Type: "string", Value: string(rune(v))}, true
			}
		}
	case "std.algorithms.euclidean_distance":
		if len(operands) == 4 {
			x1, _ := toFloat(operandValue(fr, operands[0]))
			y1, _ := toFloat(operandValue(fr, operands[1]))
			x2, _ := toFloat(operandValue(fr, operands[2]))
			y2, _ := toFloat(operandValue(fr, operands[3]))
			dx := x2 - x1
			dy := y2 - y1
			return Result{Type: "float", Value: math.Sqrt(dx*dx + dy*dy)}, true
		}
	case "std.algorithms.manhattan_distance":
		if len(operands) == 4 {
			x1, _ := toFloat(operandValue(fr, operands[0]))
			y1, _ := toFloat(operandValue(fr, operands[1]))
			x2, _ := toFloat(operandValue(fr, operands[2]))
			y2, _ := toFloat(operandValue(fr, operands[3]))
			return Result{Type: "float", Value: math.Abs(x2-x1) + math.Abs(y2-y1)}, true
		}
	case "std.array.contains":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			target := operandValue(fr, operands[1])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if t, ok := target.Value.(string); ok {
					for _, v := range ss {
						if v == t {
							return Result{Type: "bool", Value: true}, true
						}
					}
					return Result{Type: "bool", Value: false}, true
				}
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				ti, _ := toInt(target)
				for _, v := range xs {
					if v == ti {
						return Result{Type: "bool", Value: true}, true
					}
				}
				return Result{Type: "bool", Value: false}, true
			}
		}
	case "std.array.index_of":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			target := operandValue(fr, operands[1])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if t, ok := target.Value.(string); ok {
					for i, v := range ss {
						if v == t {
							return Result{Type: "int", Value: i}, true
						}
					}
					return Result{Type: "int", Value: -1}, true
				}
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				ti, _ := toInt(target)
				for i, v := range xs {
					if v == ti {
						return Result{Type: "int", Value: i}, true
					}
				}
				return Result{Type: "int", Value: -1}, true
			}
		}
	case "std.array.append":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			val := operandValue(fr, operands[1])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if s, ok := val.Value.(string); ok {
					out := make([]string, len(ss)+1)
					copy(out, ss)
					out[len(ss)] = s
					return Result{Type: "array<string>", Value: out}, true
				}
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				vi, _ := toInt(val)
				out := make([]int, len(xs)+1)
				copy(out, xs)
				out[len(xs)] = vi
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.array.prepend":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			val := operandValue(fr, operands[1])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if s, ok := val.Value.(string); ok {
					out := make([]string, len(ss)+1)
					out[0] = s
					copy(out[1:], ss)
					return Result{Type: "array<string>", Value: out}, true
				}
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				vi, _ := toInt(val)
				out := make([]int, len(xs)+1)
				out[0] = vi
				copy(out[1:], xs)
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.array.insert":
		if len(operands) == 3 {
			arr := operandValue(fr, operands[0])
			idx, _ := toInt(operandValue(fr, operands[1]))
			val := operandValue(fr, operands[2])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if s, ok := val.Value.(string); ok {
					if idx < 0 {
						idx = 0
					}
					if idx > len(ss) {
						idx = len(ss)
					}
					out := make([]string, len(ss)+1)
					copy(out, ss[:idx])
					out[idx] = s
					copy(out[idx+1:], ss[idx:])
					return Result{Type: "array<string>", Value: out}, true
				}
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				vi, _ := toInt(val)
				if idx < 0 {
					idx = 0
				}
				if idx > len(xs) {
					idx = len(xs)
				}
				out := make([]int, len(xs)+1)
				copy(out, xs[:idx])
				out[idx] = vi
				copy(out[idx+1:], xs[idx:])
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.array.remove":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			idx, _ := toInt(operandValue(fr, operands[1]))
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if idx < 0 || idx >= len(ss) {
					out := make([]string, len(ss))
					copy(out, ss)
					return Result{Type: "array<string>", Value: out}, true
				}
				out := make([]string, 0, len(ss)-1)
				out = append(out, ss[:idx]...)
				out = append(out, ss[idx+1:]...)
				return Result{Type: "array<string>", Value: out}, true
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				if idx < 0 || idx >= len(xs) {
					out := make([]int, len(xs))
					copy(out, xs)
					return Result{Type: "array<int>", Value: out}, true
				}
				out := make([]int, 0, len(xs)-1)
				out = append(out, xs[:idx]...)
				out = append(out, xs[idx+1:]...)
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.array.concat":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if as := vmArrayAsStrings(a.Value); as != nil {
				if bs := vmArrayAsStrings(b.Value); bs != nil {
					out := make([]string, 0, len(as)+len(bs))
					out = append(out, as...)
					out = append(out, bs...)
					return Result{Type: "array<string>", Value: out}, true
				}
			}
			as := vmArrayAsInts(a.Value)
			bs := vmArrayAsInts(b.Value)
			if as != nil && bs != nil {
				out := make([]int, 0, len(as)+len(bs))
				out = append(out, as...)
				out = append(out, bs...)
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.array.slice":
		if len(operands) == 3 {
			arr := operandValue(fr, operands[0])
			lo, _ := toInt(operandValue(fr, operands[1]))
			hi, _ := toInt(operandValue(fr, operands[2]))
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				if lo < 0 {
					lo = 0
				}
				if hi > len(ss) {
					hi = len(ss)
				}
				if hi < lo {
					hi = lo
				}
				out := make([]string, hi-lo)
				copy(out, ss[lo:hi])
				return Result{Type: "array<string>", Value: out}, true
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				if lo < 0 {
					lo = 0
				}
				if hi > len(xs) {
					hi = len(xs)
				}
				if hi < lo {
					hi = lo
				}
				out := make([]int, hi-lo)
				copy(out, xs[lo:hi])
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.algorithms.find_max":
		if len(operands) == 1 {
			arr := operandValue(fr, operands[0])
			xs := vmArrayAsInts(arr.Value)
			if len(xs) > 0 {
				m := xs[0]
				for _, v := range xs[1:] {
					if v > m {
						m = v
					}
				}
				return Result{Type: "int", Value: m}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.algorithms.find_min":
		if len(operands) == 1 {
			arr := operandValue(fr, operands[0])
			xs := vmArrayAsInts(arr.Value)
			if len(xs) > 0 {
				m := xs[0]
				for _, v := range xs[1:] {
					if v < m {
						m = v
					}
				}
				return Result{Type: "int", Value: m}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.algorithms.linear_search":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			target, _ := toInt(operandValue(fr, operands[1]))
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				for i, v := range xs {
					if v == target {
						return Result{Type: "int", Value: i}, true
					}
				}
			}
			return Result{Type: "int", Value: -1}, true
		}
	case "std.algorithms.binary_search":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			target, _ := toInt(operandValue(fr, operands[1]))
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				lo, hi := 0, len(xs)-1
				for lo <= hi {
					mid := lo + (hi-lo)/2
					if xs[mid] == target {
						return Result{Type: "int", Value: mid}, true
					} else if xs[mid] < target {
						lo = mid + 1
					} else {
						hi = mid - 1
					}
				}
			}
			return Result{Type: "int", Value: -1}, true
		}
	case "std.algorithms.count_occurrences":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			target, _ := toInt(operandValue(fr, operands[1]))
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				count := 0
				for _, v := range xs {
					if v == target {
						count++
					}
				}
				return Result{Type: "int", Value: count}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.algorithms.bubble_sort", "std.algorithms.selection_sort", "std.algorithms.insertion_sort":
		if len(operands) == 1 {
			arr := operandValue(fr, operands[0])
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				out := make([]int, len(xs))
				copy(out, xs)
				// Simple insertion sort — VM doesn't need to honor the
				// algorithm's name, only its sorted-output contract.
				for i := 1; i < len(out); i++ {
					key := out[i]
					j := i - 1
					for j >= 0 && out[j] > key {
						out[j+1] = out[j]
						j--
					}
					out[j+1] = key
				}
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.algorithms.reverse":
		if len(operands) == 1 {
			arr := operandValue(fr, operands[0])
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				out := make([]int, len(xs))
				for i, v := range xs {
					out[len(xs)-1-i] = v
				}
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.collections.queue_create":
		// VM carries the queue as a *[]int — pointer so set/destruct
		// flows update the same underlying slice across calls.
		q := &[]int{}
		return Result{Type: "queue<int>", Value: q}, true
	case "std.collections.queue_enqueue":
		if len(operands) == 2 {
			q := operandValue(fr, operands[0])
			el, _ := toInt(operandValue(fr, operands[1]))
			if qp, ok := q.Value.(*[]int); ok {
				*qp = append(*qp, el)
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.collections.queue_dequeue":
		if len(operands) == 1 {
			q := operandValue(fr, operands[0])
			if qp, ok := q.Value.(*[]int); ok && len(*qp) > 0 {
				v := (*qp)[0]
				*qp = (*qp)[1:]
				return Result{Type: "int", Value: v}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.queue_peek":
		if len(operands) == 1 {
			q := operandValue(fr, operands[0])
			if qp, ok := q.Value.(*[]int); ok && len(*qp) > 0 {
				return Result{Type: "int", Value: (*qp)[0]}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.queue_is_empty":
		if len(operands) == 1 {
			q := operandValue(fr, operands[0])
			if qp, ok := q.Value.(*[]int); ok {
				return Result{Type: "bool", Value: len(*qp) == 0}, true
			}
			return Result{Type: "bool", Value: true}, true
		}
	case "std.collections.queue_size":
		if len(operands) == 1 {
			q := operandValue(fr, operands[0])
			if qp, ok := q.Value.(*[]int); ok {
				return Result{Type: "int", Value: len(*qp)}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.queue_clear":
		if len(operands) == 1 {
			q := operandValue(fr, operands[0])
			if qp, ok := q.Value.(*[]int); ok {
				*qp = (*qp)[:0]
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.collections.stack_create":
		s := &[]int{}
		return Result{Type: "stack<int>", Value: s}, true
	case "std.collections.stack_push":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			el, _ := toInt(operandValue(fr, operands[1]))
			if sp, ok := s.Value.(*[]int); ok {
				*sp = append(*sp, el)
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.collections.stack_pop":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(*[]int); ok && len(*sp) > 0 {
				v := (*sp)[len(*sp)-1]
				*sp = (*sp)[:len(*sp)-1]
				return Result{Type: "int", Value: v}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.stack_peek":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(*[]int); ok && len(*sp) > 0 {
				return Result{Type: "int", Value: (*sp)[len(*sp)-1]}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.stack_is_empty":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(*[]int); ok {
				return Result{Type: "bool", Value: len(*sp) == 0}, true
			}
			return Result{Type: "bool", Value: true}, true
		}
	case "std.collections.stack_size":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(*[]int); ok {
				return Result{Type: "int", Value: len(*sp)}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.stack_clear":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(*[]int); ok {
				*sp = (*sp)[:0]
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.collections.set_create":
		s := map[int]bool{}
		return Result{Type: "set<int>", Value: s}, true
	case "std.collections.set_add":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			el, _ := toInt(operandValue(fr, operands[1]))
			if sp, ok := s.Value.(map[int]bool); ok {
				_, present := sp[el]
				sp[el] = true
				return Result{Type: "bool", Value: !present}, true
			}
			return Result{Type: "bool", Value: false}, true
		}
	case "std.collections.set_remove":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			el, _ := toInt(operandValue(fr, operands[1]))
			if sp, ok := s.Value.(map[int]bool); ok {
				_, present := sp[el]
				delete(sp, el)
				return Result{Type: "bool", Value: present}, true
			}
			return Result{Type: "bool", Value: false}, true
		}
	case "std.collections.set_contains":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			el, _ := toInt(operandValue(fr, operands[1]))
			if sp, ok := s.Value.(map[int]bool); ok {
				_, present := sp[el]
				return Result{Type: "bool", Value: present}, true
			}
			return Result{Type: "bool", Value: false}, true
		}
	case "std.collections.set_size":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(map[int]bool); ok {
				return Result{Type: "int", Value: len(sp)}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.set_clear":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if sp, ok := s.Value.(map[int]bool); ok {
				for k := range sp {
					delete(sp, k)
				}
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.collections.set_union":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			out := map[int]bool{}
			if ap, ok := a.Value.(map[int]bool); ok {
				for k := range ap {
					out[k] = true
				}
			}
			if bp, ok := b.Value.(map[int]bool); ok {
				for k := range bp {
					out[k] = true
				}
			}
			return Result{Type: "set<int>", Value: out}, true
		}
	case "std.collections.set_intersection":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			out := map[int]bool{}
			ap, _ := a.Value.(map[int]bool)
			bp, _ := b.Value.(map[int]bool)
			if ap != nil && bp != nil {
				for k := range ap {
					if bp[k] {
						out[k] = true
					}
				}
			}
			return Result{Type: "set<int>", Value: out}, true
		}
	case "std.collections.set_difference":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			out := map[int]bool{}
			ap, _ := a.Value.(map[int]bool)
			bp, _ := b.Value.(map[int]bool)
			if ap != nil {
				for k := range ap {
					if bp == nil || !bp[k] {
						out[k] = true
					}
				}
			}
			return Result{Type: "set<int>", Value: out}, true
		}
	case "std.collections.size":
		if len(operands) == 1 {
			m := operandValue(fr, operands[0])
			if mp, ok := m.Value.(map[interface{}]interface{}); ok {
				return Result{Type: "int", Value: len(mp)}, true
			}
		}
	case "std.collections.has":
		if len(operands) == 2 {
			m := operandValue(fr, operands[0])
			k := operandValue(fr, operands[1])
			if mp, ok := m.Value.(map[interface{}]interface{}); ok {
				_, present := mp[k.Value]
				return Result{Type: "bool", Value: present}, true
			}
		}
	case "std.collections.get":
		if len(operands) == 2 {
			m := operandValue(fr, operands[0])
			k := operandValue(fr, operands[1])
			if mp, ok := m.Value.(map[interface{}]interface{}); ok {
				if v, present := mp[k.Value]; present {
					switch x := v.(type) {
					case int:
						return Result{Type: "int", Value: x}, true
					case int32:
						return Result{Type: "int", Value: int(x)}, true
					case string:
						return Result{Type: "string", Value: x}, true
					default:
						return Result{Type: "int", Value: v}, true
					}
				}
			}
			return Result{Type: "int", Value: 0}, true
		}
	case "std.collections.set":
		if len(operands) == 3 {
			m := operandValue(fr, operands[0])
			k := operandValue(fr, operands[1])
			v := operandValue(fr, operands[2])
			if mp, ok := m.Value.(map[interface{}]interface{}); ok {
				mp[k.Value] = v.Value
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.collections.remove":
		if len(operands) == 2 {
			m := operandValue(fr, operands[0])
			k := operandValue(fr, operands[1])
			if mp, ok := m.Value.(map[interface{}]interface{}); ok {
				_, present := mp[k.Value]
				delete(mp, k.Value)
				return Result{Type: "bool", Value: present}, true
			}
			return Result{Type: "bool", Value: false}, true
		}
	case "std.collections.clear":
		if len(operands) == 1 {
			m := operandValue(fr, operands[0])
			if mp, ok := m.Value.(map[interface{}]interface{}); ok {
				for k := range mp {
					delete(mp, k)
				}
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.string.split":
		if len(operands) == 2 {
			s, _ := toString(operandValue(fr, operands[0]))
			d, _ := toString(operandValue(fr, operands[1]))
			var parts []string
			if d == "" {
				for _, r := range s {
					parts = append(parts, string(r))
				}
			} else {
				parts = strings.Split(s, d)
			}
			return Result{Type: "array<string>", Value: parts}, true
		}
	case "std.string.split_lines":
		if len(operands) == 1 {
			s, _ := toString(operandValue(fr, operands[0]))
			parts := strings.Split(s, "\n")
			return Result{Type: "array<string>", Value: parts}, true
		}
	case "std.string.split_words":
		if len(operands) == 1 {
			s, _ := toString(operandValue(fr, operands[0]))
			parts := strings.Fields(s)
			return Result{Type: "array<string>", Value: parts}, true
		}
	case "std.string.join":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			sep, _ := toString(operandValue(fr, operands[1]))
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				return Result{Type: "string", Value: strings.Join(ss, sep)}, true
			}
		}
	case "std.string.replace", "std.string.replace_all":
		if len(operands) == 3 {
			s, _ := toString(operandValue(fr, operands[0]))
			old, _ := toString(operandValue(fr, operands[1]))
			repl, _ := toString(operandValue(fr, operands[2]))
			if old == "" {
				return Result{Type: "string", Value: s}, true
			}
			return Result{Type: "string", Value: strings.ReplaceAll(s, old, repl)}, true
		}
	case "std.string.replace_first":
		if len(operands) == 3 {
			s, _ := toString(operandValue(fr, operands[0]))
			old, _ := toString(operandValue(fr, operands[1]))
			repl, _ := toString(operandValue(fr, operands[2]))
			return Result{Type: "string", Value: strings.Replace(s, old, repl, 1)}, true
		}
	case "std.string.replace_last":
		if len(operands) == 3 {
			s, _ := toString(operandValue(fr, operands[0]))
			old, _ := toString(operandValue(fr, operands[1]))
			repl, _ := toString(operandValue(fr, operands[2]))
			if old == "" {
				return Result{Type: "string", Value: s}, true
			}
			idx := strings.LastIndex(s, old)
			if idx < 0 {
				return Result{Type: "string", Value: s}, true
			}
			return Result{Type: "string", Value: s[:idx] + repl + s[idx+len(old):]}, true
		}
	case "std.string.find_all":
		if len(operands) == 2 {
			s, _ := toString(operandValue(fr, operands[0]))
			sub, _ := toString(operandValue(fr, operands[1]))
			var hits []int
			if sub != "" {
				start := 0
				for {
					idx := strings.Index(s[start:], sub)
					if idx < 0 {
						break
					}
					hits = append(hits, start+idx)
					start = start + idx + len(sub)
				}
			}
			return Result{Type: "array<int>", Value: hits}, true
		}
	case "std.math.random_seed":
		if len(operands) == 1 {
			seed, _ := toInt(operandValue(fr, operands[0]))
			vmRandomSeed(uint32(seed))
			return Result{Type: "void", Value: nil}, true
		}
	case "std.math.random_int":
		if len(operands) == 1 {
			bound, _ := toInt(operandValue(fr, operands[0]))
			return Result{Type: "int", Value: vmRandomInt(bound)}, true
		}
	case "std.algorithms.shuffle":
		if len(operands) == 1 {
			arr := operandValue(fr, operands[0])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				out := make([]string, len(ss))
				copy(out, ss)
				for i := len(out) - 1; i > 0; i-- {
					j := vmRandomInt(i + 1)
					out[i], out[j] = out[j], out[i]
				}
				return Result{Type: "array<string>", Value: out}, true
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				out := make([]int, len(xs))
				copy(out, xs)
				for i := len(out) - 1; i > 0; i-- {
					j := vmRandomInt(i + 1)
					out[i], out[j] = out[j], out[i]
				}
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.algorithms.unique":
		if len(operands) == 1 {
			arr := operandValue(fr, operands[0])
			if ss := vmArrayAsStrings(arr.Value); ss != nil {
				out := make([]string, 0, len(ss))
				seen := map[string]bool{}
				for _, v := range ss {
					if !seen[v] {
						seen[v] = true
						out = append(out, v)
					}
				}
				return Result{Type: "array<string>", Value: out}, true
			}
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				out := make([]int, 0, len(xs))
				seen := map[int]bool{}
				for _, v := range xs {
					if !seen[v] {
						seen[v] = true
						out = append(out, v)
					}
				}
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.algorithms.rotate":
		if len(operands) == 2 {
			arr := operandValue(fr, operands[0])
			k, _ := toInt(operandValue(fr, operands[1]))
			xs := vmArrayAsInts(arr.Value)
			if xs != nil {
				n := len(xs)
				out := make([]int, n)
				if n <= 1 {
					copy(out, xs)
					return Result{Type: "array<int>", Value: out}, true
				}
				shift := k % n
				if shift < 0 {
					shift += n
				}
				for i, v := range xs {
					out[(i+shift)%n] = v
				}
				return Result{Type: "array<int>", Value: out}, true
			}
		}
	case "std.algorithms.levenshtein_distance":
		if len(operands) == 2 {
			a, _ := toString(operandValue(fr, operands[0]))
			b, _ := toString(operandValue(fr, operands[1]))
			ar := []rune(a)
			br := []rune(b)
			m := len(ar)
			n := len(br)
			if m == 0 {
				return Result{Type: "int", Value: n}, true
			}
			if n == 0 {
				return Result{Type: "int", Value: m}, true
			}
			prev := make([]int, n+1)
			curr := make([]int, n+1)
			for j := 0; j <= n; j++ {
				prev[j] = j
			}
			for i := 1; i <= m; i++ {
				curr[0] = i
				for j := 1; j <= n; j++ {
					cost := 0
					if ar[i-1] != br[j-1] {
						cost = 1
					}
					del := prev[j] + 1
					ins := curr[j-1] + 1
					sub := prev[j-1] + cost
					min1 := del
					if ins < min1 {
						min1 = ins
					}
					if sub < min1 {
						min1 = sub
					}
					curr[j] = min1
				}
				prev, curr = curr, prev
			}
			return Result{Type: "int", Value: prev[n]}, true
		}
	case "std.time.now":
		return Result{Type: "std.time.Time", Value: buildTimeStruct(time.Now().UTC())}, true
	case "std.time.unix_timestamp":
		return Result{Type: "int", Value: int(time.Now().Unix())}, true
	case "std.time.unix_nano":
		return Result{Type: "int", Value: int(time.Now().UnixNano())}, true
	case "std.time.sleep_seconds":
		if len(operands) == 1 {
			seconds, err := toFloat(operandValue(fr, operands[0]))
			if err == nil && seconds > 0 {
				time.Sleep(time.Duration(seconds * float64(time.Second)))
			}
		}
		return Result{Type: "void", Value: nil}, true
	case "std.time.sleep_milliseconds":
		if len(operands) == 1 {
			milliseconds, err := toInt(operandValue(fr, operands[0]))
			if err == nil && milliseconds > 0 {
				time.Sleep(time.Duration(milliseconds) * time.Millisecond)
			}
		}
		return Result{Type: "void", Value: nil}, true
	case "std.time.time_zone_offset":
		_, offset := time.Now().Zone()
		return Result{Type: "int", Value: offset}, true
	case "std.time.time_zone_name":
		name, _ := time.Now().Zone()
		return Result{Type: "string", Value: name}, true
	case "std.time.time_from_unix":
		if len(operands) == 1 {
			timestamp, err := toInt(operandValue(fr, operands[0]))
			if err == nil {
				return Result{Type: "std.time.Time", Value: buildTimeStruct(time.Unix(int64(timestamp), 0).UTC())}, true
			}
		}
		return Result{Type: "std.time.Time", Value: buildTimeStruct(time.Unix(0, 0).UTC())}, true
	case "std.time.time_from_string":
		if len(operands) == 1 {
			s, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				if t, parseErr := time.Parse(time.RFC3339Nano, s); parseErr == nil {
					return Result{Type: "std.time.Time", Value: buildTimeStruct(t.UTC())}, true
				}
			}
		}
		return Result{Type: "std.time.Time", Value: buildTimeStruct(time.Unix(0, 0).UTC())}, true
	case "std.time.time_to_unix":
		if len(operands) == 1 {
			if t, ok := timeFromStruct(operandValue(fr, operands[0])); ok {
				return Result{Type: "int", Value: int(t.Unix())}, true
			}
		}
		return Result{Type: "int", Value: 0}, true
	case "std.time.time_to_string":
		if len(operands) == 1 {
			if t, ok := timeFromStruct(operandValue(fr, operands[0])); ok {
				return Result{Type: "string", Value: t.Format(time.RFC3339Nano)}, true
			}
		}
		return Result{Type: "string", Value: "1970-01-01T00:00:00Z"}, true
	case "std.time.time_to_unix_nano":
		if len(operands) == 1 {
			if t, ok := timeFromStruct(operandValue(fr, operands[0])); ok {
				return Result{Type: "int", Value: int(t.UnixNano())}, true
			}
		}
		return Result{Type: "int", Value: 0}, true
	case "std.time.duration_to_string":
		if len(operands) == 1 {
			if seconds, nanoseconds, ok := durationFields(operandValue(fr, operands[0])); ok {
				if nanoseconds == 0 {
					return Result{Type: "string", Value: fmt.Sprintf("%ds", seconds)}, true
				}
				return Result{Type: "string", Value: fmt.Sprintf("%d.%09ds", seconds, nanoseconds)}, true
			}
		}
		return Result{Type: "string", Value: "0s"}, true
	// File operations
	case "std.file.exists":
		if len(operands) == 1 {
			if name, err := toString(operandValue(fr, operands[0])); err == nil {
				exists, _ := vmFileExists(name)
				return Result{Type: "int", Value: boolToInt(exists)}, true
			}
		}
		return Result{Type: "int", Value: 0}, true
	case "std.file.size":
		if len(operands) == 1 {
			if name, err := toString(operandValue(fr, operands[0])); err == nil {
				statSize, sizeErr := vmFileSize(name)
				if sizeErr != nil {
					return Result{Type: "int", Value: -1}, true
				}
				return Result{Type: "int", Value: statSize}, true
			}
		}
		return Result{Type: "int", Value: -1}, true
	case "std.file.open":
		if len(operands) == 2 {
			name, err1 := toString(operandValue(fr, operands[0]))
			mode, err2 := toString(operandValue(fr, operands[1]))
			if err1 != nil || err2 != nil {
				return Result{Type: "int", Value: -1}, true
			}
			handle, err := vmFileOpen(name, mode)
			if err != nil {
				return Result{Type: "int", Value: -1}, true
			}
			return Result{Type: "int", Value: handle}, true
		}
		return Result{Type: "int", Value: -1}, true
	case "std.file.write":
		if len(operands) == 3 {
			handleVal := operandValue(fr, operands[0])
			data, errData := toString(operandValue(fr, operands[1]))
			sizeVal, errSize := toInt(operandValue(fr, operands[2]))
			if errData == nil && errSize == nil {
				handle, errHandle := toInt(handleVal)
				if errHandle == nil {
					written, writeErr := vmFileWrite(handle, data, sizeVal)
					if writeErr != nil {
						return Result{Type: "int", Value: -1}, true
					}
					return Result{Type: "int", Value: written}, true
				}
			}
		}
		return Result{Type: "int", Value: -1}, true
	case "std.file.close":
		if len(operands) == 1 {
			fileHandle, err := toInt(operandValue(fr, operands[0]))
			if err != nil {
				return Result{Type: "int", Value: -1}, true
			}
			if err := vmFileClose(fileHandle); err != nil {
				return Result{Type: "int", Value: -1}, true
			}
			return Result{Type: "int", Value: 0}, true
		}
		return Result{Type: "int", Value: -1}, true
	case "std.string.length":
		if len(operands) == 1 {
			arg := operandValue(fr, operands[0])
			if arg.Type == "string" {
				str := arg.Value.(string)
				return Result{Type: "int", Value: len(str)}, true
			}
		}
	case "std.string.concat":
		if len(operands) == 2 {
			left := operandValue(fr, operands[0])
			right := operandValue(fr, operands[1])
			if left.Type == "string" && right.Type == "string" {
				leftStr := left.Value.(string)
				rightStr := right.Value.(string)
				return Result{Type: "string", Value: leftStr + rightStr}, true
			}
		}
	case "std.string.substring":
		if len(operands) == 3 {
			s := operandValue(fr, operands[0])
			start := operandValue(fr, operands[1])
			end := operandValue(fr, operands[2])
			if s.Type == "string" && start.Type == "int" && end.Type == "int" {
				str := s.Value.(string)
				startIdx := start.Value.(int)
				endIdx := end.Value.(int)
				if startIdx < 0 || endIdx > len(str) || startIdx > endIdx {
					return Result{Type: "string", Value: ""}, true
				}
				return Result{Type: "string", Value: str[startIdx:endIdx]}, true
			}
		}
	case "std.string.char_at":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			idx := operandValue(fr, operands[1])
			if s.Type == "string" && idx.Type == "int" {
				str := s.Value.(string)
				index := idx.Value.(int)
				if index < 0 || index >= len(str) {
					return Result{Type: "char", Value: ' '}, true
				}
				return Result{Type: "char", Value: rune(str[index])}, true
			}
		}
	case "std.string.starts_with":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			prefix := operandValue(fr, operands[1])
			if s.Type == "string" && prefix.Type == "string" {
				str := s.Value.(string)
				pre := prefix.Value.(string)
				if len(pre) > len(str) {
					return Result{Type: "bool", Value: false}, true
				}
				return Result{Type: "bool", Value: str[:len(pre)] == pre}, true
			}
		}
	case "std.string.ends_with":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			suffix := operandValue(fr, operands[1])
			if s.Type == "string" && suffix.Type == "string" {
				str := s.Value.(string)
				suf := suffix.Value.(string)
				if len(suf) > len(str) {
					return Result{Type: "bool", Value: false}, true
				}
				return Result{Type: "bool", Value: str[len(str)-len(suf):] == suf}, true
			}
		}
	case "std.string.contains":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			substr := operandValue(fr, operands[1])
			if s.Type == "string" && substr.Type == "string" {
				str := s.Value.(string)
				sub := substr.Value.(string)
				return Result{Type: "bool", Value: strings.Contains(str, sub)}, true
			}
		}
	case "std.string.index_of":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			substr := operandValue(fr, operands[1])
			if s.Type == "string" && substr.Type == "string" {
				str := s.Value.(string)
				sub := substr.Value.(string)
				idx := strings.Index(str, sub)
				return Result{Type: "int", Value: idx}, true
			}
		}
	case "std.string.last_index_of":
		if len(operands) == 2 {
			s := operandValue(fr, operands[0])
			substr := operandValue(fr, operands[1])
			if s.Type == "string" && substr.Type == "string" {
				str := s.Value.(string)
				sub := substr.Value.(string)
				idx := strings.LastIndex(str, sub)
				return Result{Type: "int", Value: idx}, true
			}
		}
	case "std.string.trim":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				str := s.Value.(string)
				return Result{Type: "string", Value: strings.TrimSpace(str)}, true
			}
		}
	case "std.string.to_upper":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				str := s.Value.(string)
				return Result{Type: "string", Value: strings.ToUpper(str)}, true
			}
		}
	case "std.string.to_lower":
		if len(operands) == 1 {
			s := operandValue(fr, operands[0])
			if s.Type == "string" {
				str := s.Value.(string)
				return Result{Type: "string", Value: strings.ToLower(str)}, true
			}
		}
	case "std.string.equals":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "string" && b.Type == "string" {
				aStr := a.Value.(string)
				bStr := b.Value.(string)
				return Result{Type: "bool", Value: aStr == bStr}, true
			}
		}
	case "std.string.compare":
		if len(operands) == 2 {
			a := operandValue(fr, operands[0])
			b := operandValue(fr, operands[1])
			if a.Type == "string" && b.Type == "string" {
				aStr := a.Value.(string)
				bStr := b.Value.(string)
				return Result{Type: "int", Value: strings.Compare(aStr, bStr)}, true
			}
		}
	case "std.testing.suite":
		if len(operands) == 0 {
			id := newTestingSuiteID()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.expect":
		if len(operands) >= 3 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
			}
			name := stringFromResult(operandValue(fr, operands[1]))
			passed := boolFromResult(operandValue(fr, operands[2]))
			message := ""
			if len(operands) >= 4 {
				message = stringFromResult(operandValue(fr, operands[3]))
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			recordTestingResultLocked(suite, name, passed, message)
			testingSuitesMu.Unlock()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.pass":
		if len(operands) >= 2 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
			}
			name := stringFromResult(operandValue(fr, operands[1]))
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			recordTestingResultLocked(suite, name, true, "")
			testingSuitesMu.Unlock()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.fail":
		if len(operands) >= 3 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
			}
			name := stringFromResult(operandValue(fr, operands[1]))
			message := stringFromResult(operandValue(fr, operands[2]))
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			recordTestingResultLocked(suite, name, false, message)
			testingSuitesMu.Unlock()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.equal_int":
		if len(operands) >= 4 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
			}
			name := stringFromResult(operandValue(fr, operands[1]))
			expectedRes := operandValue(fr, operands[2])
			actualRes := operandValue(fr, operands[3])
			expectedVal, expErr := toInt(expectedRes)
			actualVal, actErr := toInt(actualRes)
			passed := expErr == nil && actErr == nil && expectedVal == actualVal
			message := fmt.Sprintf("expected %v, got %v", expectedRes.Value, actualRes.Value)
			if expErr == nil && actErr == nil {
				message = fmt.Sprintf("expected %d, got %d", expectedVal, actualVal)
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			recordTestingResultLocked(suite, name, passed, message)
			testingSuitesMu.Unlock()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.equal_bool":
		if len(operands) >= 4 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
			}
			name := stringFromResult(operandValue(fr, operands[1]))
			expectedRes := operandValue(fr, operands[2])
			actualRes := operandValue(fr, operands[3])
			expectedVal, expErr := toBool(expectedRes)
			actualVal, actErr := toBool(actualRes)
			passed := expErr == nil && actErr == nil && expectedVal == actualVal
			expStr := stringFromResult(expectedRes)
			actStr := stringFromResult(actualRes)
			message := fmt.Sprintf("expected %s, got %s", expStr, actStr)
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			recordTestingResultLocked(suite, name, passed, message)
			testingSuitesMu.Unlock()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.equal_string":
		if len(operands) >= 4 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
			}
			name := stringFromResult(operandValue(fr, operands[1]))
			expectedStr := stringFromResult(operandValue(fr, operands[2]))
			actualStr := stringFromResult(operandValue(fr, operands[3]))
			passed := expectedStr == actualStr
			message := fmt.Sprintf("expected %q, got %q", expectedStr, actualStr)
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			recordTestingResultLocked(suite, name, passed, message)
			testingSuitesMu.Unlock()
			return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(id)}, true
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.equal_float":
		if len(operands) >= 4 {
			return execTestingEqualFloat(fr, operands, 6)
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.equal_float_precision":
		if len(operands) >= 5 {
			precisionRes := operandValue(fr, operands[4])
			precision, err := toInt(precisionRes)
			if err != nil || precision < 0 {
				precision = 6
			}
			return execTestingEqualFloat(fr, operands[:4], precision)
		}
		return Result{Type: "std.testing.Suite", Value: buildTestingSuiteStruct(0)}, true
	case "std.testing.total":
		if len(operands) >= 1 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "int", Value: 0}, true
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			total := suite.total
			testingSuitesMu.Unlock()
			return Result{Type: "int", Value: total}, true
		}
		return Result{Type: "int", Value: 0}, true
	case "std.testing.failures":
		if len(operands) >= 1 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "int", Value: 0}, true
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			failures := suite.failed
			testingSuitesMu.Unlock()
			return Result{Type: "int", Value: failures}, true
		}
		return Result{Type: "int", Value: 0}, true
	case "std.testing.summary":
		if len(operands) >= 1 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "int", Value: 0}, true
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			total := suite.total
			failures := suite.failed
			testingSuitesMu.Unlock()
			passedCount := total - failures
			fmt.Printf("\nTest Summary: %d total, %d passed, %d failed\n", total, passedCount, failures)
			return Result{Type: "int", Value: failures}, true
		}
		return Result{Type: "int", Value: 0}, true
	case "std.testing.passed":
		if len(operands) >= 1 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				return Result{Type: "bool", Value: false}, true
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			passed := suite.failed == 0
			testingSuitesMu.Unlock()
			return Result{Type: "bool", Value: passed}, true
		}
		return Result{Type: "bool", Value: false}, true
	case "std.testing.exit":
		if len(operands) >= 1 {
			suiteIDRes := operandValue(fr, operands[0])
			id, ok := suiteIDFromResult(suiteIDRes)
			if !ok {
				panic(exitSignal{code: 0})
			}
			testingSuitesMu.Lock()
			suite := ensureTestingSuiteLocked(id)
			failures := suite.failed
			testingSuitesMu.Unlock()
			panic(exitSignal{code: failures})
		}
		panic(exitSignal{code: 0})
	case "std.test.start":
		if len(operands) == 1 {
			testName := operandValue(fr, operands[0])
			if testName.Type == "string" {
				fmt.Printf("Running test: %s\n", testName.Value)
				return Result{Type: "void", Value: nil}, true
			}
		}
	case "std.test.end":
		if len(operands) == 2 {
			testName := operandValue(fr, operands[0])
			passed := operandValue(fr, operands[1])
			if testName.Type != "string" || passed.Type != "bool" {
				return Result{Type: "void", Value: nil}, true
			}

			passedVal, _ := toBool(passed)
			if passedVal {
				fmt.Printf("✓ %s PASSED\n", testName.Value)
			} else {
				fmt.Printf("✗ %s FAILED\n", testName.Value)
			}
			return Result{Type: "void", Value: nil}, true
		}
	case "std.assert":
		if len(operands) == 2 {
			condition := operandValue(fr, operands[0])
			message := operandValue(fr, operands[1])
			if condition.Type != "bool" || message.Type != "string" {
				return Result{Type: "void", Value: nil}, true
			}

			conditionVal, _ := toBool(condition)
			if !conditionVal {
				fmt.Printf("  ASSERTION FAILED: %s\n", message.Value)
			}

			return Result{Type: "void", Value: nil}, true
		}
	case "std.assert.eq":
		if len(operands) == 3 {
			expected := operandValue(fr, operands[0])
			actual := operandValue(fr, operands[1])
			message := operandValue(fr, operands[2])
			if message.Type != "string" {
				return Result{Type: "void", Value: nil}, true
			}

			var equal bool
			if expected.Type == actual.Type {
				switch expected.Type {
				case "int":
					expectedVal, _ := toInt(expected)
					actualVal, _ := toInt(actual)
					equal = expectedVal == actualVal
				case "string":
					expectedStr, _ := toString(expected)
					actualStr, _ := toString(actual)
					equal = expectedStr == actualStr
				case "bool":
					expectedBool, _ := toBool(expected)
					actualBool, _ := toBool(actual)
					equal = expectedBool == actualBool
				case "float", "double":
					expectedFloat, _ := toFloat(expected)
					actualFloat, _ := toFloat(actual)
					equal = expectedFloat == actualFloat
				default:
					equal = expected.Value == actual.Value
				}
			}

			if !equal {
				fmt.Printf("  ASSERTION FAILED: %s (expected: %v, actual: %v)\n", message.Value, expected.Value, actual.Value)
			}

			return Result{Type: "void", Value: nil}, true
		}
	case "std.assert.true":
		if len(operands) == 2 {
			condition := operandValue(fr, operands[0])
			message := operandValue(fr, operands[1])
			if condition.Type != "bool" || message.Type != "string" {
				return Result{Type: "void", Value: nil}, true
			}

			conditionVal, _ := toBool(condition)
			if !conditionVal {
				fmt.Printf("  ASSERTION FAILED: %s (expected: true, actual: false)\n", message.Value)
			}
		}

		return Result{Type: "void", Value: nil}, true
	case "std.assert.false":
		if len(operands) == 2 {
			condition := operandValue(fr, operands[0])
			message := operandValue(fr, operands[1])
			if condition.Type != "bool" || message.Type != "string" {
				return Result{Type: "void", Value: nil}, true
			}

			conditionVal, _ := toBool(condition)
			if conditionVal {
				fmt.Printf("  ASSERTION FAILED: %s (expected: false, actual: true)\n", message.Value)
			}
		}

		return Result{Type: "void", Value: nil}, true
	case "std.test.summary":
		if len(operands) == 0 {
			fmt.Printf("\nTest Summary: All tests completed\n")
			return Result{Type: "int", Value: 0}, true
		}
	case "std.os.args":
		args := cloneCLIArgs()
		if args == nil {
			args = []string{}
		}
		return Result{Type: "array<string>", Value: args}, true
	case "std.os.has_flag":
		if len(operands) == 1 {
			name, err := toString(operandValue(fr, operands[0]))
			if err != nil {
				return Result{Type: "bool", Value: false}, true
			}
			return Result{Type: "bool", Value: hasFlag(name)}, true
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.get_flag":
		if len(operands) == 2 {
			name, err1 := toString(operandValue(fr, operands[0]))
			if err1 != nil {
				return Result{Type: "string", Value: ""}, true
			}
			def, err2 := toString(operandValue(fr, operands[1]))
			if err2 != nil {
				return Result{Type: "string", Value: ""}, true
			}
			return Result{Type: "string", Value: getFlagValue(name, def)}, true
		}
		return Result{Type: "string", Value: ""}, true
	case "std.os.positional_arg":
		if len(operands) == 2 {
			index, err1 := toInt(operandValue(fr, operands[0]))
			if err1 != nil {
				return Result{Type: "string", Value: ""}, true
			}
			def, err2 := toString(operandValue(fr, operands[1]))
			if err2 != nil {
				return Result{Type: "string", Value: ""}, true
			}
			return Result{Type: "string", Value: positionalArgValue(index, def)}, true
		}
		return Result{Type: "string", Value: ""}, true
	case "std.os.args_count":
		cliArgsMu.RLock()
		count := len(cliArgs)
		cliArgsMu.RUnlock()
		return Result{Type: "int", Value: count}, true
	case "std.os.exit":
		code := 0
		if len(operands) >= 1 {
			if val, err := toInt(operandValue(fr, operands[0])); err == nil {
				code = val
			}
		}
		panic(exitSignal{code: code})
	case "std.os.getenv":
		if len(operands) == 1 {
			name, err := toString(operandValue(fr, operands[0]))
			if err != nil {
				return Result{Type: "string", Value: ""}, true
			}
			return Result{Type: "string", Value: os.Getenv(name)}, true
		}
		return Result{Type: "string", Value: ""}, true
	case "std.os.setenv":
		if len(operands) == 2 {
			name, err1 := toString(operandValue(fr, operands[0]))
			value, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				return Result{Type: "bool", Value: os.Setenv(name, value) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.unsetenv":
		if len(operands) == 1 {
			name, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				return Result{Type: "bool", Value: os.Unsetenv(name) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.getpid":
		return Result{Type: "int", Value: os.Getpid()}, true
	case "std.os.getppid":
		return Result{Type: "int", Value: os.Getppid()}, true
	case "std.os.getcwd":
		cwd, err := os.Getwd()
		if err != nil {
			return Result{Type: "string", Value: ""}, true
		}
		return Result{Type: "string", Value: cwd}, true
	case "std.os.chdir":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				return Result{Type: "bool", Value: os.Chdir(path) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.mkdir":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				if path == "" {
					return Result{Type: "bool", Value: false}, true
				}
				if err := os.MkdirAll(path, 0o755); err == nil {
					return Result{Type: "bool", Value: true}, true
				}
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.rmdir":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				if path == "" {
					return Result{Type: "bool", Value: false}, true
				}
				return Result{Type: "bool", Value: os.Remove(path) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.exists":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				exists, _ := vmPathExists(path)
				return Result{Type: "bool", Value: exists}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.is_file":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				isFile, _ := vmIsFile(path)
				return Result{Type: "bool", Value: isFile}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.is_dir":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				isDir, _ := vmIsDir(path)
				return Result{Type: "bool", Value: isDir}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.remove":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				return Result{Type: "bool", Value: os.Remove(path) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.rename":
		if len(operands) == 2 {
			src, err1 := toString(operandValue(fr, operands[0]))
			dst, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				return Result{Type: "bool", Value: os.Rename(src, dst) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.copy":
		if len(operands) == 2 {
			src, err1 := toString(operandValue(fr, operands[0]))
			dst, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				return Result{Type: "bool", Value: vmCopyFile(src, dst) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.read_file":
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					return Result{Type: "string", Value: ""}, true
				}
				return Result{Type: "string", Value: string(data)}, true
			}
		}
		return Result{Type: "string", Value: ""}, true
	case "std.os.write_file":
		if len(operands) == 2 {
			path, err1 := toString(operandValue(fr, operands[0]))
			content, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				return Result{Type: "bool", Value: os.WriteFile(path, []byte(content), 0o644) == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.append_file":
		if len(operands) == 2 {
			path, err1 := toString(operandValue(fr, operands[0]))
			content, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				if openErr != nil {
					return Result{Type: "bool", Value: false}, true
				}
				_, writeErr := io.WriteString(f, content)
				closeErr := f.Close()
				return Result{Type: "bool", Value: writeErr == nil && closeErr == nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.os.read_file_async":
		// Async version - execute in goroutine and return Promise
		if len(operands) == 1 {
			path, err := toString(operandValue(fr, operands[0]))
			if err == nil {
				promiseID := newPromise()
				go func() {
					data, readErr := os.ReadFile(path)
					if readErr != nil {
						rejectPromise(promiseID, readErr)
					} else {
						resolvePromise(promiseID, Result{Type: "string", Value: string(data)})
					}
				}()
				return Result{Type: "Promise", Value: promiseID}, true
			}
		}
		// Return rejected promise on error
		promiseID := newPromise()
		rejectPromise(promiseID, fmt.Errorf("invalid arguments"))
		return Result{Type: "Promise", Value: promiseID}, true
	case "std.os.write_file_async":
		// Async version - execute in goroutine and return Promise
		if len(operands) == 2 {
			path, err1 := toString(operandValue(fr, operands[0]))
			content, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				promiseID := newPromise()
				go func() {
					writeErr := os.WriteFile(path, []byte(content), 0o644)
					if writeErr != nil {
						rejectPromise(promiseID, writeErr)
					} else {
						resolvePromise(promiseID, Result{Type: "bool", Value: true})
					}
				}()
				return Result{Type: "Promise", Value: promiseID}, true
			}
		}
		// Return rejected promise on error
		promiseID := newPromise()
		rejectPromise(promiseID, fmt.Errorf("invalid arguments"))
		return Result{Type: "Promise", Value: promiseID}, true
	case "std.os.append_file_async":
		// Async version - execute in goroutine and return Promise
		if len(operands) == 2 {
			path, err1 := toString(operandValue(fr, operands[0]))
			content, err2 := toString(operandValue(fr, operands[1]))
			if err1 == nil && err2 == nil {
				promiseID := newPromise()
				go func() {
					f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
					if openErr != nil {
						rejectPromise(promiseID, openErr)
						return
					}
					_, writeErr := io.WriteString(f, content)
					closeErr := f.Close()
					if writeErr != nil || closeErr != nil {
						rejectPromise(promiseID, fmt.Errorf("write error: %v, close error: %v", writeErr, closeErr))
					} else {
						resolvePromise(promiseID, Result{Type: "bool", Value: true})
					}
				}()
				return Result{Type: "Promise", Value: promiseID}, true
			}
		}
		// Return rejected promise on error
		promiseID := newPromise()
		rejectPromise(promiseID, fmt.Errorf("invalid arguments"))
		return Result{Type: "Promise", Value: promiseID}, true
	case "std.file.read":
		if len(operands) == 3 {
			handle, errHandle := toInt(operandValue(fr, operands[0]))
			// buffer operand ignored in VM; present for parity
			_ = operandValue(fr, operands[1])
			size, errSize := toInt(operandValue(fr, operands[2]))
			if errHandle == nil && errSize == nil {
				read, readErr := vmFileRead(handle, size)
				if readErr != nil {
					return Result{Type: "int", Value: -1}, true
				}
				return Result{Type: "int", Value: read}, true
			}
		}
		return Result{Type: "int", Value: -1}, true
	case "std.file.seek":
		if len(operands) == 3 {
			handle, errHandle := toInt(operandValue(fr, operands[0]))
			offset, errOffset := toInt(operandValue(fr, operands[1]))
			whence, errWhence := toInt(operandValue(fr, operands[2]))
			if errHandle == nil && errOffset == nil && errWhence == nil {
				result, seekErr := vmFileSeek(handle, offset, whence)
				if seekErr != nil {
					return Result{Type: "int", Value: -1}, true
				}
				return Result{Type: "int", Value: result}, true
			}
		}
		return Result{Type: "int", Value: -1}, true
	case "std.file.tell":
		if len(operands) == 1 {
			handle, err := toInt(operandValue(fr, operands[0]))
			if err == nil {
				pos, tellErr := vmFileTell(handle)
				if tellErr != nil {
					return Result{Type: "int", Value: -1}, true
				}
				return Result{Type: "int", Value: pos}, true
			}
		}
		return Result{Type: "int", Value: -1}, true
	case "std.network.ip_parse":
		if len(operands) == 1 {
			if ipStr, err := toString(operandValue(fr, operands[0])); err == nil {
				return Result{Type: "std.network.IPAddress", Value: buildIPAddressStruct(ipStr)}, true
			}
		}
		return Result{Type: "std.network.IPAddress", Value: buildIPAddressStruct("")}, true
	case "std.network.ip_is_valid":
		if len(operands) == 1 {
			if ipStr, err := toString(operandValue(fr, operands[0])); err == nil {
				return Result{Type: "bool", Value: net.ParseIP(ipStr) != nil}, true
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.network.ip_is_private":
		if len(operands) == 1 {
			if ipStr, ok := ipValueToString(operandValue(fr, operands[0])); ok {
				if ip := net.ParseIP(ipStr); ip != nil {
					return Result{Type: "bool", Value: ip.IsPrivate()}, true
				}
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.network.ip_is_loopback":
		if len(operands) == 1 {
			if ipStr, ok := ipValueToString(operandValue(fr, operands[0])); ok {
				if ip := net.ParseIP(ipStr); ip != nil {
					return Result{Type: "bool", Value: ip.IsLoopback()}, true
				}
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.network.ip_to_string":
		if len(operands) == 1 {
			if ipStr, ok := ipValueToString(operandValue(fr, operands[0])); ok {
				return Result{Type: "string", Value: ipStr}, true
			}
		}
		return Result{Type: "string", Value: ""}, true
	case "std.network.url_parse":
		if len(operands) == 1 {
			if urlStr, err := toString(operandValue(fr, operands[0])); err == nil {
				if parsed, urlErr := urlpkg.Parse(urlStr); urlErr == nil {
					return Result{Type: "std.network.URL", Value: buildURLStruct(parsed)}, true
				}
			}
		}
		return Result{Type: "std.network.URL", Value: buildURLStruct(nil)}, true
	case "std.network.url_to_string":
		if len(operands) == 1 {
			val := operandValue(fr, operands[0])
			if val.Type == "string" {
				return val, true
			}
			if urlStr, ok := urlFromStruct(val); ok {
				return Result{Type: "string", Value: urlStr}, true
			}
		}
		return Result{Type: "string", Value: ""}, true
	case "std.network.url_is_valid":
		if len(operands) == 1 {
			if urlStr, err := toString(operandValue(fr, operands[0])); err == nil {
				if parsed, parseErr := urlpkg.Parse(urlStr); parseErr == nil && parsed.Scheme != "" && parsed.Host != "" {
					return Result{Type: "bool", Value: true}, true
				}
			}
		}
		return Result{Type: "bool", Value: false}, true
	case "std.network.dns_lookup":
		return Result{Type: "array<std.network.IPAddress>", Value: []map[string]interface{}{}}, true
	case "std.network.dns_reverse_lookup":
		return Result{Type: "string", Value: ""}, true
	case "std.network.http_get":
		return Result{Type: "std.network.HTTPResponse", Value: buildHTTPPlaceholderResponse()}, true
	case "std.network.http_post":
		return Result{Type: "std.network.HTTPResponse", Value: buildHTTPPlaceholderResponse()}, true
	case "std.network.http_put":
		return Result{Type: "std.network.HTTPResponse", Value: buildHTTPPlaceholderResponse()}, true
	case "std.network.http_delete":
		return Result{Type: "std.network.HTTPResponse", Value: buildHTTPPlaceholderResponse()}, true
	}

	return Result{}, false
}

func operandValue(fr *frame, op mir.Operand) Result {
	switch op.Kind {
	case mir.OperandValue:
		return fr.values[op.Value]
	case mir.OperandLiteral:
		res, err := literalResult(op)
		if err != nil {
			return Result{Type: op.Type, Value: nil}
		}
		return res
	default:
		return Result{Type: "<unknown>", Value: nil}
	}
}

func literalResult(op mir.Operand) (Result, error) {
	typ := op.Type
	if typ == "" {
		typ = inferLiteralType(op.Literal)
	}
	switch typ {
	case "int", "long", "byte":
		// Handle hex and binary literals
		convertedLiteral := convertLiteralToDecimal(op.Literal)
		v, err := strconv.Atoi(convertedLiteral)
		if err != nil {
			return Result{}, fmt.Errorf("invalid int literal %q", op.Literal)
		}
		return Result{Type: typ, Value: v}, nil
	case "float", "double":
		v, err := strconv.ParseFloat(op.Literal, 64)
		if err != nil {
			return Result{}, fmt.Errorf("invalid float literal %q", op.Literal)
		}
		return Result{Type: typ, Value: v}, nil
	case "bool":
		// Handle boolean literals properly
		if op.Literal == "true" {
			return Result{Type: typ, Value: true}, nil
		} else if op.Literal == "false" {
			return Result{Type: typ, Value: false}, nil
		}
		return Result{}, fmt.Errorf("invalid bool literal %q", op.Literal)
	case "string":
		// Strip the wrapping quotes and decode escape sequences. The C
		// backend gets the same behavior for free because it forwards
		// the raw source lexeme to the C compiler, which decodes escapes
		// when it parses the C string literal.
		return Result{Type: typ, Value: decodeStringLiteral(op.Literal)}, nil
	case "char":
		// char literals come through as 'a' / '\n' — strip the wrapping
		// single quotes and decode common escapes. Stored as rune so
		// std.char_code can return its int value.
		body := op.Literal
		if len(body) >= 2 && body[0] == '\'' && body[len(body)-1] == '\'' {
			body = body[1 : len(body)-1]
		}
		var r rune
		switch body {
		case `\n`:
			r = '\n'
		case `\t`:
			r = '\t'
		case `\r`:
			r = '\r'
		case `\\`:
			r = '\\'
		case `\'`:
			r = '\''
		case `\"`:
			r = '"'
		case `\0`:
			r = 0
		default:
			rs := []rune(body)
			if len(rs) > 0 {
				r = rs[0]
			}
		}
		return Result{Type: "char", Value: r}, nil
	case "null":
		return Result{Type: typ, Value: nil}, nil
	default:
		return Result{Type: typ, Value: nil}, nil
	}
}

func inferLiteralType(lit string) string {
	if lit == "true" || lit == "false" {
		return "bool"
	}
	if lit == "null" {
		return "null"
	}
	if _, err := strconv.Atoi(lit); err == nil {
		return "int"
	}
	if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
		return "string"
	}
	return "<unknown>"
}

// convertLiteralToDecimal converts hex and binary literals to decimal
func convertLiteralToDecimal(literal string) string {
	if strings.HasPrefix(literal, "0x") || strings.HasPrefix(literal, "0X") {
		// Hex literal - convert to decimal
		hexStr := literal[2:]
		// Remove underscores
		hexStr = strings.ReplaceAll(hexStr, "_", "")
		// Convert to int64 and back to string
		if val, err := strconv.ParseInt(hexStr, 16, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	} else if strings.HasPrefix(literal, "0b") || strings.HasPrefix(literal, "0B") {
		// Binary literal - convert to decimal
		binaryStr := literal[2:]
		// Remove underscores
		binaryStr = strings.ReplaceAll(binaryStr, "_", "")
		// Convert to int64 and back to string
		if val, err := strconv.ParseInt(binaryStr, 2, 64); err == nil {
			return strconv.FormatInt(val, 10)
		}
	}
	// Return as-is for regular decimal literals
	return literal
}

// VM-side xorshift32 mirroring the runtime PRNG. Sharing the same
// algorithm keeps tests that pin the seed reproducible across both
// backends.
var (
	vmRngMu    sync.Mutex
	vmRngState uint32 = 0x9E3779B9
)

func vmRandomSeed(seed uint32) {
	vmRngMu.Lock()
	defer vmRngMu.Unlock()
	if seed == 0 {
		vmRngState = 0x9E3779B9
	} else {
		vmRngState = seed
	}
}

func vmRandomInt(bound int) int {
	if bound <= 0 {
		return 0
	}
	vmRngMu.Lock()
	defer vmRngMu.Unlock()
	x := vmRngState
	if x == 0 {
		x = 0x9E3779B9
	}
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	vmRngState = x
	return int(x % uint32(bound))
}

// vmArrayAsStrings coerces a VM array carrier into `[]string`. Same
// shape as vmArrayAsInts: returns nil when the value isn't array-
// shaped or contains a non-string element.
func vmArrayAsStrings(v interface{}) []string {
	switch xs := v.(type) {
	case []string:
		return xs
	case []interface{}:
		out := make([]string, 0, len(xs))
		for _, e := range xs {
			s, ok := e.(string)
			if !ok {
				return nil
			}
			out = append(out, s)
		}
		return out
	}
	return nil
}

// vmArrayAsInts coerces a VM array carrier (Go `[]int` for typed
// integer arrays, `[]interface{}` for heterogeneous or generic ones)
// into a `[]int` for the std.algorithms intrinsics. Returns nil when
// the value isn't array-shaped or contains a non-integer element.
func vmArrayAsInts(v interface{}) []int {
	switch xs := v.(type) {
	case []int:
		return xs
	case []interface{}:
		out := make([]int, 0, len(xs))
		for _, e := range xs {
			switch n := e.(type) {
			case int:
				out = append(out, n)
			case int32:
				out = append(out, int(n))
			case int64:
				out = append(out, int(n))
			default:
				return nil
			}
		}
		return out
	}
	return nil
}

func toInt(value Result) (int, error) {
	switch v := value.Value.(type) {
	case int:
		return v, nil
	case int32:
		// rune values (chars) and any callee that returns int32 land here.
		return int(v), nil
	case int64:
		return int(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, fmt.Errorf("vm: expected int, got nil")
	default:
		return 0, fmt.Errorf("vm: expected int value, got %T", v)
	}
}

func toBool(value Result) (bool, error) {
	switch v := value.Value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("vm: expected bool value, got %T", v)
	}
}

func toFloat(value Result) (float64, error) {
	switch v := value.Value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case nil:
		return 0, fmt.Errorf("vm: expected float, got nil")
	default:
		return 0, fmt.Errorf("vm: expected float value, got %T", v)
	}
}

func toString(value Result) (string, error) {
	switch v := value.Value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int32:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case float32:
		return fmt.Sprintf("%g", v), nil
	case float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("vm: cannot convert %T to string", v)
	}
}

func chooseType(preferred, fallback string) string {
	if preferred != "" && preferred != inferTypePlaceholder {
		return preferred
	}
	if fallback != "" && fallback != inferTypePlaceholder {
		return fallback
	}
	return "int"
}

// execMalloc handles dynamic memory allocation
func execMalloc(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("malloc: expected 1 operand (size), got %d", len(inst.Operands))
	}

	size := operandValue(fr, inst.Operands[0])
	sizeVal, err := toInt(size)
	if err != nil {
		return Result{}, fmt.Errorf("malloc: size must be int, got %s", size.Type)
	}

	if sizeVal <= 0 {
		return Result{}, fmt.Errorf("malloc: size must be positive, got %d", sizeVal)
	}

	// In the VM, we simulate memory allocation by creating a byte slice
	// In a real implementation, this would allocate actual memory
	allocatedMemory := make([]byte, sizeVal)

	// Return a pointer-like value (in VM, we use the slice directly)
	return Result{Type: "ptr", Value: allocatedMemory}, nil
}

// execFree handles dynamic memory deallocation
func execFree(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("free: expected 1 operand (pointer), got %d", len(inst.Operands))
	}

	_ = operandValue(fr, inst.Operands[0]) // ptr - not used in VM simulation

	// In the VM, we simulate memory deallocation
	// In a real implementation, this would free actual memory
	// For now, we just return void
	return Result{Type: "void", Value: nil}, nil
}

// execRealloc handles dynamic memory reallocation
func execRealloc(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("realloc: expected 2 operands (pointer, new_size), got %d", len(inst.Operands))
	}

	ptr := operandValue(fr, inst.Operands[0])
	newSize := operandValue(fr, inst.Operands[1])

	newSizeVal, err := toInt(newSize)
	if err != nil {
		return Result{}, fmt.Errorf("realloc: new_size must be int, got %s", newSize.Type)
	}

	if newSizeVal <= 0 {
		return Result{}, fmt.Errorf("realloc: new_size must be positive, got %d", newSizeVal)
	}

	// In the VM, we simulate memory reallocation by creating a new slice
	// In a real implementation, this would reallocate actual memory
	newMemory := make([]byte, newSizeVal)

	// Copy old data if the pointer is valid
	if ptr.Value != nil {
		if oldMemory, ok := ptr.Value.([]byte); ok {
			copySize := newSizeVal
			if len(oldMemory) < copySize {
				copySize = len(oldMemory)
			}
			copy(newMemory, oldMemory[:copySize])
		}
	}

	return Result{Type: "ptr", Value: newMemory}, nil
}

// execFileOpen handles file opening
func execFileOpen(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("file.open: expected 2 operands (filename, mode), got %d", len(inst.Operands))
	}

	name, err1 := toString(operandValue(fr, inst.Operands[0]))
	mode, err2 := toString(operandValue(fr, inst.Operands[1]))
	if err1 != nil || err2 != nil {
		return Result{}, fmt.Errorf("file.open: filename and mode must be strings")
	}
	handle, err := vmFileOpen(name, mode)
	if err != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: handle}, nil
}

// execFileClose handles file closing
func execFileClose(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("file.close: expected 1 operand (file_handle), got %d", len(inst.Operands))
	}

	fileHandle, err := toInt(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.close: file_handle must be int")
	}
	if err := vmFileClose(fileHandle); err != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: 0}, nil
}

// execFileRead handles file reading
func execFileRead(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 3 {
		return Result{}, fmt.Errorf("file.read: expected 3 operands (file_handle, buffer, size), got %d", len(inst.Operands))
	}

	fileHandle, err := toInt(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.read: file_handle must be int")
	}
	_ = operandValue(fr, inst.Operands[1])
	size, errSize := toInt(operandValue(fr, inst.Operands[2]))
	if errSize != nil {
		return Result{}, fmt.Errorf("file.read: size must be int")
	}
	read, readErr := vmFileRead(fileHandle, size)
	if readErr != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: read}, nil
}

// execFileWrite handles file writing
func execFileWrite(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 3 {
		return Result{}, fmt.Errorf("file.write: expected 3 operands (file_handle, buffer, size), got %d", len(inst.Operands))
	}

	fileHandle, err := toInt(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.write: file_handle must be int")
	}
	data, errData := toString(operandValue(fr, inst.Operands[1]))
	if errData != nil {
		return Result{}, fmt.Errorf("file.write: buffer must be string")
	}
	size, errSize := toInt(operandValue(fr, inst.Operands[2]))
	if errSize != nil {
		return Result{}, fmt.Errorf("file.write: size must be int")
	}
	written, writeErr := vmFileWrite(fileHandle, data, size)
	if writeErr != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: written}, nil
}

// execFileSeek handles file seeking
func execFileSeek(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 3 {
		return Result{}, fmt.Errorf("file.seek: expected 3 operands (file_handle, offset, whence), got %d", len(inst.Operands))
	}

	fileHandle, err := toInt(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.seek: file_handle must be int")
	}
	offset, errOff := toInt(operandValue(fr, inst.Operands[1]))
	if errOff != nil {
		return Result{}, fmt.Errorf("file.seek: offset must be int")
	}
	whence, errWhence := toInt(operandValue(fr, inst.Operands[2]))
	if errWhence != nil {
		return Result{}, fmt.Errorf("file.seek: whence must be int")
	}
	result, seekErr := vmFileSeek(fileHandle, offset, whence)
	if seekErr != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: result}, nil
}

// execFileTell handles file position querying
func execFileTell(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("file.tell: expected 1 operand (file_handle), got %d", len(inst.Operands))
	}

	fileHandle, err := toInt(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.tell: file_handle must be int")
	}
	pos, tellErr := vmFileTell(fileHandle)
	if tellErr != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: pos}, nil
}

// execFileExists handles file existence checking
func execFileExists(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("file.exists: expected 1 operand (filename), got %d", len(inst.Operands))
	}

	filename, err := toString(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.exists: filename must be string")
	}
	exists, existsErr := vmFileExists(filename)
	if existsErr != nil {
		return Result{Type: "bool", Value: false}, nil
	}
	return Result{Type: "bool", Value: exists}, nil
}

// execFileSize handles file size querying
func execFileSize(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("file.size: expected 1 operand (filename), got %d", len(inst.Operands))
	}

	filename, err := toString(operandValue(fr, inst.Operands[0]))
	if err != nil {
		return Result{}, fmt.Errorf("file.size: filename must be string")
	}
	statSize, sizeErr := vmFileSize(filename)
	if sizeErr != nil {
		return Result{Type: "int", Value: -1}, nil
	}
	return Result{Type: "int", Value: statSize}, nil
}

// execTestStart handles test start
func execTestStart(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 1 {
		return Result{}, fmt.Errorf("test.start: expected 1 operand (test_name), got %d", len(inst.Operands))
	}

	testName := operandValue(fr, inst.Operands[0])
	if testName.Type != "string" {
		return Result{}, fmt.Errorf("test.start: test_name must be string")
	}

	// In the VM, we simulate test start
	fmt.Printf("Running test: %s\n", testName.Value)
	return Result{Type: "void", Value: nil}, nil
}

// execTestEnd handles test end
func execTestEnd(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("test.end: expected 2 operands (test_name, passed), got %d", len(inst.Operands))
	}

	testName := operandValue(fr, inst.Operands[0])
	passed := operandValue(fr, inst.Operands[1])

	if testName.Type != "string" || passed.Type != "bool" {
		return Result{}, fmt.Errorf("test.end: test_name must be string, passed must be bool")
	}

	// In the VM, we simulate test end
	passedVal, _ := toBool(passed)
	if passedVal {
		fmt.Printf("✓ %s PASSED\n", testName.Value)
	} else {
		fmt.Printf("✗ %s FAILED\n", testName.Value)
	}
	return Result{Type: "void", Value: nil}, nil
}

// execAssert handles basic assertion
func execAssert(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("assert: expected 2 operands (condition, message), got %d", len(inst.Operands))
	}

	condition := operandValue(fr, inst.Operands[0])
	message := operandValue(fr, inst.Operands[1])

	if condition.Type != "bool" || message.Type != "string" {
		return Result{}, fmt.Errorf("assert: condition must be bool, message must be string")
	}

	conditionVal, _ := toBool(condition)
	if !conditionVal {
		fmt.Printf("  ASSERTION FAILED: %s\n", message.Value)
		return Result{Type: "void", Value: nil}, fmt.Errorf("assertion failed: %s", message.Value)
	}

	return Result{Type: "void", Value: nil}, nil
}

// execAssertEq handles equality assertion
func execAssertEq(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 3 {
		return Result{}, fmt.Errorf("assert.eq: expected 3 operands (expected, actual, message), got %d", len(inst.Operands))
	}

	expected := operandValue(fr, inst.Operands[0])
	actual := operandValue(fr, inst.Operands[1])
	message := operandValue(fr, inst.Operands[2])

	if message.Type != "string" {
		return Result{}, fmt.Errorf("assert.eq: message must be string")
	}

	// Check if values are equal based on their types
	var equal bool
	if expected.Type == actual.Type {
		switch expected.Type {
		case "int":
			expectedVal, _ := toInt(expected)
			actualVal, _ := toInt(actual)
			equal = expectedVal == actualVal
		case "string":
			expectedStr, _ := toString(expected)
			actualStr, _ := toString(actual)
			equal = expectedStr == actualStr
		case "bool":
			expectedBool, _ := toBool(expected)
			actualBool, _ := toBool(actual)
			equal = expectedBool == actualBool
		case "float", "double":
			expectedFloat, _ := toFloat(expected)
			actualFloat, _ := toFloat(actual)
			equal = expectedFloat == actualFloat
		default:
			equal = expected.Value == actual.Value
		}
	}

	if !equal {
		fmt.Printf("  ASSERTION FAILED: %s (expected: %v, actual: %v)\n", message.Value, expected.Value, actual.Value)
		return Result{Type: "void", Value: nil}, fmt.Errorf("assertion failed: %s", message.Value)
	}

	return Result{Type: "void", Value: nil}, nil
}

// execAssertTrue handles true assertion
func execAssertTrue(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("assert.true: expected 2 operands (condition, message), got %d", len(inst.Operands))
	}

	condition := operandValue(fr, inst.Operands[0])
	message := operandValue(fr, inst.Operands[1])

	if condition.Type != "bool" || message.Type != "string" {
		return Result{}, fmt.Errorf("assert.true: condition must be bool, message must be string")
	}

	conditionVal, _ := toBool(condition)
	if !conditionVal {
		fmt.Printf("  ASSERTION FAILED: %s (expected: true, actual: false)\n", message.Value)
		return Result{Type: "void", Value: nil}, fmt.Errorf("assertion failed: %s", message.Value)
	}

	return Result{Type: "void", Value: nil}, nil
}

// execAssertFalse handles false assertion
func execAssertFalse(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) != 2 {
		return Result{}, fmt.Errorf("assert.false: expected 2 operands (condition, message), got %d", len(inst.Operands))
	}

	condition := operandValue(fr, inst.Operands[0])
	message := operandValue(fr, inst.Operands[1])

	if condition.Type != "bool" || message.Type != "string" {
		return Result{}, fmt.Errorf("assert.false: condition must be bool, message must be string")
	}

	conditionVal, _ := toBool(condition)
	if conditionVal {
		fmt.Printf("  ASSERTION FAILED: %s (expected: false, actual: true)\n", message.Value)
		return Result{Type: "void", Value: nil}, fmt.Errorf("assertion failed: %s", message.Value)
	}

	return Result{Type: "void", Value: nil}, nil
}

// execFuncRef handles function reference (getting a function pointer)
func execFuncRef(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 1 {
		return Result{}, fmt.Errorf("func.ref: expected at least 1 operand")
	}

	funcName := inst.Operands[0].Literal

	// Store the function name as the value
	// The type will be the function type from the instruction
	return Result{Type: inst.Type, Value: funcName}, nil
}

// execFuncAssign handles function assignment
func execFuncAssign(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 1 {
		return Result{}, fmt.Errorf("func.assign: expected at least 1 operand")
	}

	funcValue := operandValue(fr, inst.Operands[0])
	return funcValue, nil
}

// execFuncCall handles function call through function pointer or closure
func execFuncCall(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 1 {
		return Result{}, fmt.Errorf("func.call: expected at least 1 operand")
	}

	funcValue := operandValue(fr, inst.Operands[0])

	// Check if this is a closure or a simple function reference
	if closure, ok := funcValue.Value.(map[string]interface{}); ok {
		// This is a closure - extract the function name and captured variables
		funcName, ok := closure["function"].(string)
		if !ok {
			return Result{}, fmt.Errorf("func.call: closure function name is not a string")
		}

		// Get the function from the map
		fn, exists := funcs[funcName]
		if !exists {
			return Result{}, fmt.Errorf("func.call: function %q not found", funcName)
		}

		// Prepare arguments: first the lambda parameters, then the captured variables
		args := make([]Result, len(inst.Operands)-1)
		for i, op := range inst.Operands[1:] {
			args[i] = operandValue(fr, op)
		}

		// Add captured variables as additional arguments
		if captured, ok := closure["captured"].(map[string]interface{}); ok {
			for _, capturedValue := range captured {
				args = append(args, Result{Type: "int", Value: capturedValue}) // Default to int type
			}
		}

		// Call the function with all arguments
		return execFunction(funcs, fn, args)
	} else {
		// This is a simple function reference
		funcName, ok := funcValue.Value.(string)
		if !ok {
			return Result{}, fmt.Errorf("func.call: function value is not a string or closure")
		}

		// Get the function from the map
		fn, exists := funcs[funcName]
		if !exists {
			return Result{}, fmt.Errorf("func.call: function %q not found", funcName)
		}

		// Prepare arguments
		args := make([]Result, len(inst.Operands)-1)
		for i, op := range inst.Operands[1:] {
			args[i] = operandValue(fr, op)
		}

		// Call the function
		return execFunction(funcs, fn, args)
	}
}

// execClosureCreate handles closure creation
func execClosureCreate(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 1 {
		return Result{}, fmt.Errorf("closure.create: expected at least 1 operand")
	}

	funcName := inst.Operands[0].Literal

	// Create a closure structure
	closure := map[string]interface{}{
		"function": funcName,
		"captured": make(map[string]interface{}),
	}

	return Result{Type: inst.Type, Value: closure}, nil
}

// execClosureCapture handles capturing variables in a closure
func execClosureCapture(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 3 {
		return Result{}, fmt.Errorf("closure.capture: expected at least 3 operands")
	}

	closureValue := operandValue(fr, inst.Operands[0])
	varName := inst.Operands[1].Literal
	varValue := operandValue(fr, inst.Operands[2])

	// Get the closure map
	closure, ok := closureValue.Value.(map[string]interface{})
	if !ok {
		return Result{}, fmt.Errorf("closure.capture: closure value is not a map")
	}

	// Get the captured variables map
	captured, ok := closure["captured"].(map[string]interface{})
	if !ok {
		return Result{}, fmt.Errorf("closure.capture: captured variables is not a map")
	}

	// Store the captured variable
	captured[varName] = varValue.Value

	return Result{Type: "void", Value: nil}, nil
}

// execClosureBind handles binding a closure to create a function reference
func execClosureBind(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 1 {
		return Result{}, fmt.Errorf("closure.bind: expected at least 1 operand")
	}

	closureValue := operandValue(fr, inst.Operands[0])

	// Return the closure as a function reference
	// The closure contains both the function name and captured variables
	return Result{Type: inst.Type, Value: closureValue.Value}, nil
}

// execBitwise handles bitwise operations
func execBitwise(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 2 {
		return Result{}, fmt.Errorf("bitwise operation: expected at least 2 operands")
	}

	left := operandValue(fr, inst.Operands[0])
	right := operandValue(fr, inst.Operands[1])

	// Convert operands to integers
	li, err := toInt(left)
	if err != nil {
		return Result{}, fmt.Errorf("bitwise operation: left operand must be int, got %s", left.Type)
	}
	ri, err := toInt(right)
	if err != nil {
		return Result{}, fmt.Errorf("bitwise operation: right operand must be int, got %s", right.Type)
	}

	var res int
	switch inst.Op {
	case "bitand":
		res = li & ri
	case "bitor":
		res = li | ri
	case "bitxor":
		res = li ^ ri
	case "lshift":
		res = li << ri
	case "rshift":
		res = li >> ri
	default:
		return Result{}, fmt.Errorf("unsupported bitwise operation %q", inst.Op)
	}

	return Result{Type: "int", Value: res}, nil
}

// execCast handles type casting
func execCast(funcs map[string]*mir.Function, fr *frame, inst mir.Instruction) (Result, error) {
	if len(inst.Operands) < 1 {
		return Result{}, fmt.Errorf("cast: expected at least 1 operand")
	}

	operand := operandValue(fr, inst.Operands[0])
	targetType := inst.Type

	// Handle type conversions
	switch targetType {
	case "int":
		if operand.Type == "float" || operand.Type == "double" {
			f, err := toFloat(operand)
			if err != nil {
				return Result{}, err
			}
			return Result{Type: "int", Value: int(f)}, nil
		} else if operand.Type == "bool" {
			b, err := toBool(operand)
			if err != nil {
				return Result{}, err
			}
			if b {
				return Result{Type: "int", Value: 1}, nil
			}
			return Result{Type: "int", Value: 0}, nil
		} else if operand.Type == "string" {
			// String to int conversion
			s, err := toString(operand)
			if err != nil {
				return Result{}, err
			}
			if val, err := strconv.Atoi(s); err == nil {
				return Result{Type: "int", Value: val}, nil
			}
			return Result{}, fmt.Errorf("cast: cannot convert string %q to int", s)
		}
		return Result{Type: "int", Value: operand.Value}, nil

	case "float", "double":
		if operand.Type == "int" {
			i, err := toInt(operand)
			if err != nil {
				return Result{}, err
			}
			return Result{Type: "float", Value: float64(i)}, nil
		} else if operand.Type == "bool" {
			b, err := toBool(operand)
			if err != nil {
				return Result{}, err
			}
			if b {
				return Result{Type: "float", Value: 1.0}, nil
			}
			return Result{Type: "float", Value: 0.0}, nil
		} else if operand.Type == "string" {
			// String to float conversion
			s, err := toString(operand)
			if err != nil {
				return Result{}, err
			}
			if val, err := strconv.ParseFloat(s, 64); err == nil {
				return Result{Type: "float", Value: val}, nil
			}
			return Result{}, fmt.Errorf("cast: cannot convert string %q to float", s)
		}
		return Result{Type: "float", Value: operand.Value}, nil

	case "bool":
		if operand.Type == "int" {
			i, err := toInt(operand)
			if err != nil {
				return Result{}, err
			}
			return Result{Type: "bool", Value: i != 0}, nil
		} else if operand.Type == "float" || operand.Type == "double" {
			f, err := toFloat(operand)
			if err != nil {
				return Result{}, err
			}
			return Result{Type: "bool", Value: f != 0.0}, nil
		} else if operand.Type == "string" {
			s, err := toString(operand)
			if err != nil {
				return Result{}, err
			}
			return Result{Type: "bool", Value: s != ""}, nil
		}
		return Result{Type: "bool", Value: operand.Value}, nil

	case "string":
		s, err := toString(operand)
		if err != nil {
			return Result{}, err
		}
		return Result{Type: "string", Value: s}, nil

	default:
		// For other types, just return the operand with the new type
		return Result{Type: targetType, Value: operand.Value}, nil
	}
}

func handleLogIntrinsic(level string, operands []mir.Operand, fr *frame) (Result, bool) {
	message := buildLogMessage(operands, fr)
	logger := logging.Logger()
	switch level {
	case "debug":
		logger.DebugString(message)
	case "warn":
		logger.WarnString(message)
	case "error":
		logger.ErrorString(message)
	default:
		logger.InfoString(message)
	}
	return Result{Type: "void", Value: nil}, true
}

func buildLogMessage(operands []mir.Operand, fr *frame) string {
	if len(operands) == 0 {
		return ""
	}
	parts := make([]string, 0, len(operands))
	for _, op := range operands {
		val := operandValue(fr, op)
		str, err := toString(val)
		if err != nil {
			str = fmt.Sprint(val.Value)
		}
		parts = append(parts, str)
	}
	return strings.Join(parts, " ")
}

// vmIntValue extracts an int from a Result. Helps when an operand
// could be int64 (from the VM's int convention) or int (from file
// handles).
func vmIntValue(r Result) (int, bool) {
	switch v := r.Value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	}
	return 0, false
}

func vmFprint(handle int, s string, newline bool) {
	file, err := getFileHandle(handle)
	if err != nil {
		return
	}
	io.WriteString(file, s)
	if newline {
		io.WriteString(file, "\n")
	}
}

func vmFileReadAll(handle int) string {
	file, err := getFileHandle(handle)
	if err != nil {
		return ""
	}
	data, err := io.ReadAll(file)
	if err != nil && !errors.Is(err, io.EOF) {
		return ""
	}
	return string(data)
}

func vmFileReadLine(handle int) string {
	file, err := getFileHandle(handle)
	if err != nil {
		return ""
	}
	var buf []byte
	one := make([]byte, 1)
	for {
		n, err := file.Read(one)
		if n == 0 || err != nil {
			break
		}
		if one[0] == '\r' {
			// Try to consume a paired \n. If next byte isn't \n, leave
			// it in the file (we can't unread on os.File without seeking).
			// Best effort: peek via Seek+Read.
			pos, _ := file.Seek(0, io.SeekCurrent)
			if _, err := file.Read(one); err == nil && one[0] != '\n' {
				file.Seek(pos, io.SeekStart)
			}
			break
		}
		if one[0] == '\n' {
			break
		}
		buf = append(buf, one[0])
	}
	return string(buf)
}

func vmFileWriteString(handle int, s string) int {
	file, err := getFileHandle(handle)
	if err != nil {
		return -1
	}
	n, err := io.WriteString(file, s)
	if err != nil {
		return -1
	}
	return n
}

func vmReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func vmWriteFileLines(path string, lines []string) bool {
	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()
	for _, line := range lines {
		if _, err := io.WriteString(f, line+"\n"); err != nil {
			return false
		}
	}
	return true
}

func vmFileOpen(filename, mode string) (int, error) {
	flags, perm, err := parseFileMode(mode)
	if err != nil {
		return -1, err
	}
	if perm == 0 {
		perm = 0o644
	}
	file, err := os.OpenFile(filename, flags, perm)
	if err != nil {
		return -1, err
	}
	fileHandleMu.Lock()
	handle := fileHandleCounter
	fileHandleCounter++
	fileHandleTable[handle] = file
	fileHandleMu.Unlock()
	return handle, nil
}

func vmFileClose(handle int) error {
	fileHandleMu.Lock()
	file := fileHandleTable[handle]
	if file != nil {
		delete(fileHandleTable, handle)
	}
	fileHandleMu.Unlock()
	if file == nil {
		return fmt.Errorf("invalid file handle %d", handle)
	}
	return file.Close()
}

func vmFileRead(handle int, size int) (int, error) {
	if size <= 0 {
		return 0, nil
	}
	file, err := getFileHandle(handle)
	if err != nil {
		return -1, err
	}
	buf := make([]byte, size)
	read, readErr := file.Read(buf)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return read, readErr
	}
	return read, nil
}

func vmFileWrite(handle int, data string, size int) (int, error) {
	file, err := getFileHandle(handle)
	if err != nil {
		return -1, err
	}
	if size < 0 {
		size = 0
	}
	if size > len(data) {
		size = len(data)
	}
	written, writeErr := file.Write([]byte(data)[:size])
	return written, writeErr
}

func vmFileSeek(handle int, offset int, whence int) (int, error) {
	file, err := getFileHandle(handle)
	if err != nil {
		return -1, err
	}
	var seekWhence int
	switch whence {
	case 0:
		seekWhence = io.SeekStart
	case 1:
		seekWhence = io.SeekCurrent
	case 2:
		seekWhence = io.SeekEnd
	default:
		return -1, fmt.Errorf("file.seek: invalid whence %d", whence)
	}
	if _, err := file.Seek(int64(offset), seekWhence); err != nil {
		return -1, err
	}
	return 0, nil
}

func vmFileTell(handle int) (int, error) {
	file, err := getFileHandle(handle)
	if err != nil {
		return -1, err
	}
	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1, err
	}
	return int(pos), nil
}

func vmFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func vmFileSize(path string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	return int(info.Size()), nil
}

func vmPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func vmIsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.Mode().IsRegular(), nil
}

func vmIsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func vmCopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, sourceFile)
	return err
}

func getFileHandle(handle int) (*os.File, error) {
	fileHandleMu.Lock()
	file := fileHandleTable[handle]
	fileHandleMu.Unlock()
	if file == nil {
		return nil, fmt.Errorf("invalid file handle %d", handle)
	}
	return file, nil
}

func parseFileMode(mode string) (int, os.FileMode, error) {
	m := strings.TrimSpace(mode)
	m = strings.ReplaceAll(m, "b", "")
	switch m {
	case "r":
		return os.O_RDONLY, 0, nil
	case "r+":
		return os.O_RDWR, 0, nil
	case "w":
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o644, nil
	case "w+":
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0o644, nil
	case "a":
		return os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0o644, nil
	case "a+":
		return os.O_APPEND | os.O_CREATE | os.O_RDWR, 0o644, nil
	default:
		return 0, 0, fmt.Errorf("unsupported file mode %q", mode)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func buildIPAddressStruct(address string) map[string]interface{} {
	info := map[string]interface{}{
		"address": address,
		"is_ipv4": false,
		"is_ipv6": false,
	}
	if address == "" {
		return info
	}
	ip := net.ParseIP(address)
	if ip == nil {
		return info
	}
	if ip.To4() != nil {
		info["is_ipv4"] = true
	} else if ip.To16() != nil {
		info["is_ipv6"] = true
	}
	return info
}

func buildTimeStruct(t time.Time) map[string]interface{} {
	t = t.UTC()
	return map[string]interface{}{
		"year":       t.Year(),
		"month":      int(t.Month()),
		"day":        t.Day(),
		"hour":       t.Hour(),
		"minute":     t.Minute(),
		"second":     t.Second(),
		"nanosecond": t.Nanosecond(),
	}
}

func structIntField(fields map[string]interface{}, name string) (int, bool) {
	switch v := fields[name].(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

func timeFromStruct(val Result) (time.Time, bool) {
	fields, ok := val.Value.(map[string]interface{})
	if !ok {
		return time.Time{}, false
	}
	year, ok := structIntField(fields, "year")
	if !ok {
		return time.Time{}, false
	}
	month, ok := structIntField(fields, "month")
	if !ok {
		return time.Time{}, false
	}
	day, ok := structIntField(fields, "day")
	if !ok {
		return time.Time{}, false
	}
	hour, ok := structIntField(fields, "hour")
	if !ok {
		return time.Time{}, false
	}
	minute, ok := structIntField(fields, "minute")
	if !ok {
		return time.Time{}, false
	}
	second, ok := structIntField(fields, "second")
	if !ok {
		return time.Time{}, false
	}
	nanosecond, ok := structIntField(fields, "nanosecond")
	if !ok {
		return time.Time{}, false
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC), true
}

func durationFields(val Result) (int, int, bool) {
	fields, ok := val.Value.(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	seconds, ok := structIntField(fields, "seconds")
	if !ok {
		return 0, 0, false
	}
	nanoseconds, ok := structIntField(fields, "nanoseconds")
	if !ok {
		return 0, 0, false
	}
	return seconds, nanoseconds, true
}

func buildURLStruct(u *urlpkg.URL) map[string]interface{} {
	result := map[string]interface{}{
		"scheme":   "",
		"host":     "",
		"port":     0,
		"path":     "",
		"query":    "",
		"fragment": "",
	}
	if u == nil {
		return result
	}
	result["scheme"] = u.Scheme
	result["host"] = u.Hostname()
	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			result["port"] = p
		}
	}
	result["path"] = u.EscapedPath()
	result["query"] = u.RawQuery
	result["fragment"] = u.Fragment
	return result
}

func urlFromStruct(val Result) (string, bool) {
	data, ok := val.Value.(map[string]interface{})
	if !ok {
		return "", false
	}
	scheme, _ := data["scheme"].(string)
	host, _ := data["host"].(string)
	var port int
	switch v := data["port"].(type) {
	case int:
		port = v
	case float64:
		port = int(v)
	}
	path, _ := data["path"].(string)
	query, _ := data["query"].(string)
	fragment, _ := data["fragment"].(string)
	if scheme == "" || host == "" {
		return "", false
	}
	assembled := &urlpkg.URL{
		Scheme:   scheme,
		Path:     path,
		RawQuery: strings.TrimPrefix(query, "?"),
		Fragment: strings.TrimPrefix(fragment, "#"),
	}
	// Default-port normalization mirrors the OmniLang `url_to_string`
	// implementation in std/network/network.omni: a port equal to the
	// scheme's default is dropped on output so that round-trips of
	// "http://h/" and "https://h/" stay stable. Without this the VM
	// re-emits an explicit ":80" / ":443" the C backend wouldn't.
	emitPort := port > 0
	if emitPort && ((scheme == "http" && port == 80) || (scheme == "https" && port == 443)) {
		emitPort = false
	}
	if emitPort {
		assembled.Host = fmt.Sprintf("%s:%d", host, port)
	} else {
		assembled.Host = host
	}
	return assembled.String(), true
}

func buildHTTPPlaceholderResponse() map[string]interface{} {
	return map[string]interface{}{
		"status_code": 501,
		"status_text": "Not Implemented",
		"headers":     map[string]string{},
		"body":        "",
	}
}

func ipValueToString(val Result) (string, bool) {
	if val.Type == "string" {
		if s, ok := val.Value.(string); ok {
			return s, true
		}
	}
	if m, ok := val.Value.(map[string]interface{}); ok {
		if address, ok := m["address"].(string); ok {
			return address, true
		}
	}
	return "", false
}
