package monitor

import "strings"

// shellComms is the set of comm names we recognise as interactive shells.
var shellComms = map[string]bool{
	"bash": true, "sh": true, "zsh": true, "fish": true,
	"dash": true, "ksh": true, "csh": true, "tcsh": true,
	"ash": true, "mksh": true, "elvish": true, "nu": true,
	"xonsh": true, "oil": true,
}

// busyComms are processes that are inherently "interesting" work.
var busyComms = map[string]bool{
	// compilers / build systems
	"gcc": true, "g++": true, "cc1": true, "cc1plus": true,
	"clang": true, "clang++": true, "lld": true, "ld": true,
	"rustc": true, "go": true, "javac": true, "kotlinc": true,
	"scalac": true, "swiftc": true,
	"make": true, "ninja": true, "cmake": true, "bazel": true,
	"buck": true, "ant": true, "maven": true, "gradle": true,
	// test runners
	"pytest": true, "jest": true, "cargo": true,
	"go-test": true, "mvn": true, "phpunit": true,
	// package managers / installers that do heavy work
	"npm": true, "yarn": true, "pnpm": true, "pip": true,
	"pip3": true, "apt": true, "apt-get": true, "dpkg": true,
}

// shellSession groups a shell process with the other processes in its session.
type shellSession struct {
	Shell     *procInfo   // the shell itself
	OtherPIDs []*procInfo // other processes in the same session (bg jobs etc.)
}

// buildSessionMap scans the snapshot and, for each shell, collects all
// processes that share the same session ID (SID).
func buildSessionMap(snap *procSnapshot) map[int]*shellSession {
	sessions := make(map[int]*shellSession)

	// First pass: find every shell.
	for _, p := range snap.procs {
		if isShell(p.Comm) {
			// Use the shell's SID as key (a shell is normally the session leader,
			// so SID == PID, but we store by SID to handle edge cases).
			sid := p.SID
			if _, ok := sessions[sid]; !ok {
				sessions[sid] = &shellSession{}
			}
			// If multiple shells share a SID (nested), prefer the one whose
			// PID equals the SID (the session leader).
			if p.PID == sid || sessions[sid].Shell == nil {
				sessions[sid].Shell = p
			}
		}
	}

	// Second pass: assign all non-shell processes to their shell's session.
	for _, p := range snap.procs {
		sess, ok := sessions[p.SID]
		if !ok {
			continue // not a shell session
		}
		if sess.Shell != nil && p.PID == sess.Shell.PID {
			continue // skip the shell itself
		}
		sess.OtherPIDs = append(sess.OtherPIDs, p)
	}

	return sessions
}

// isShell returns true if comm matches a known shell name.
func isShell(comm string, extras ...string) bool {
	comm = strings.ToLower(comm)
	if shellComms[comm] {
		return true
	}
	for _, e := range extras {
		if strings.ToLower(e) == comm {
			return true
		}
	}
	return false
}

// isBusyComm returns true if comm is an inherently busy process.
func isBusyComm(comm string, extras []string) bool {
	comm = strings.ToLower(comm)
	if busyComms[comm] {
		return true
	}
	for _, e := range extras {
		if strings.ToLower(e) == comm {
			return true
		}
	}
	return false
}

// isIdleState returns true for process states we consider non-busy.
// S = interruptible sleep (idle shell waiting for input)
// Z = zombie (already dead)
func isIdleState(state string) bool {
	return state == "S" || state == "Z"
}
