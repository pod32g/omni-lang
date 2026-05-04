package vm

import (
	"net"
	"strconv"
	"sync"
)

// sockEntry holds either the listener side or the connected side of a TCP
// socket. The OmniLang surface (socket_create / bind / listen / accept /
// connect / send / receive / close) is shaped for Posix FDs; Go's net
// package gives us net.Listener and net.Conn instead. We bridge by:
//
//   - socket_create reserves an id with an empty entry.
//   - socket_bind stashes the "host:port" string but doesn't open yet,
//     so it can fail before commit (matches the Posix bind/listen split).
//   - socket_listen actually calls net.Listen and stores the listener.
//   - socket_connect calls net.Dial and stores the conn.
//   - socket_accept blocks on listener.Accept and allocates a fresh id
//     for the resulting conn.
//
// All access goes through sockMu so the registry is safe across goroutines
// (spawn-blocks may share a server socket between accept loops).
type sockEntry struct {
	listener net.Listener
	conn     net.Conn
	bindAddr string // "host:port", set by socket_bind before socket_listen opens it
}

var (
	sockMu     sync.Mutex
	sockTable  = map[int]*sockEntry{}
	sockNextID = 1 // 0 is reserved as an "invalid" sentinel
)

// sockAlloc reserves a fresh id under the lock. Caller fills the entry.
func sockAlloc() (int, *sockEntry) {
	sockMu.Lock()
	defer sockMu.Unlock()
	id := sockNextID
	sockNextID++
	e := &sockEntry{}
	sockTable[id] = e
	return id, e
}

// sockGet returns the entry for id, or nil if the id is unknown.
func sockGet(id int) *sockEntry {
	sockMu.Lock()
	defer sockMu.Unlock()
	return sockTable[id]
}

// sockDelete removes an entry from the registry and returns whatever was
// there so the caller can close it outside the lock.
func sockDelete(id int) *sockEntry {
	sockMu.Lock()
	defer sockMu.Unlock()
	e := sockTable[id]
	delete(sockTable, id)
	return e
}

// vmSocketCreate reserves a fresh id; the actual TCP work happens in
// bind/listen/connect. Always succeeds.
func vmSocketCreate() int {
	id, _ := sockAlloc()
	return id
}

// vmSocketBind stashes the target address. We can't call net.Listen yet
// because socket_listen takes the backlog and is the natural commit
// point — opening here would mean ignoring the backlog parameter.
func vmSocketBind(id int, addr string, port int) bool {
	e := sockGet(id)
	if e == nil {
		return false
	}
	sockMu.Lock()
	defer sockMu.Unlock()
	e.bindAddr = net.JoinHostPort(addr, strconv.Itoa(port))
	return true
}

// vmSocketListen actually opens the listener using the address stashed by
// socket_bind. backlog is accepted by the OmniLang signature but Go's
// runtime handles the kernel listen queue itself, so it's informational.
func vmSocketListen(id int, _backlog int) bool {
	e := sockGet(id)
	if e == nil || e.bindAddr == "" {
		return false
	}
	ln, err := net.Listen("tcp", e.bindAddr)
	if err != nil {
		return false
	}
	sockMu.Lock()
	e.listener = ln
	sockMu.Unlock()
	return true
}

// vmSocketAccept blocks until a peer connects, then registers the conn
// under a fresh id so subsequent send/receive/close calls find it.
func vmSocketAccept(id int) int {
	e := sockGet(id)
	if e == nil || e.listener == nil {
		return -1
	}
	conn, err := e.listener.Accept()
	if err != nil {
		return -1
	}
	connID, connEntry := sockAlloc()
	sockMu.Lock()
	connEntry.conn = conn
	sockMu.Unlock()
	return connID
}

// vmSocketConnect dials a TCP peer and stores the conn under the existing
// id. The id was minted by socket_create — Posix-style flow — so we reuse
// its slot rather than allocating a new one.
func vmSocketConnect(id int, addr string, port int) bool {
	e := sockGet(id)
	if e == nil {
		return false
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	sockMu.Lock()
	e.conn = conn
	sockMu.Unlock()
	return true
}

// vmSocketSend writes data and returns bytes written, or -1 on error /
// missing conn (matches the C-runtime's contract).
func vmSocketSend(id int, data string) int {
	e := sockGet(id)
	if e == nil || e.conn == nil {
		return -1
	}
	n, err := e.conn.Write([]byte(data))
	if err != nil {
		return -1
	}
	return n
}

// vmSocketReceive reads up to bufferSize bytes and returns them as a
// string. Empty string on EOF, error, or missing conn — same shape the
// C runtime surfaces.
func vmSocketReceive(id int, bufferSize int) string {
	e := sockGet(id)
	if e == nil || e.conn == nil || bufferSize <= 0 {
		return ""
	}
	buf := make([]byte, bufferSize)
	n, err := e.conn.Read(buf)
	if err != nil && n == 0 {
		return ""
	}
	return string(buf[:n])
}

// vmSocketClose closes whichever side of the socket is open and drops it
// from the registry. Returns true when something was actually closed.
func vmSocketClose(id int) bool {
	e := sockDelete(id)
	if e == nil {
		return false
	}
	closed := false
	if e.conn != nil {
		_ = e.conn.Close()
		closed = true
	}
	if e.listener != nil {
		_ = e.listener.Close()
		closed = true
	}
	return closed
}
