package actions

import (
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
)

const (
	dbusDest      = "org.freedesktop.login1"
	dbusPath      = "/org/freedesktop/login1"
	dbusInterface = "org.freedesktop.login1.Manager"
)

// callLogind is a helper function that calls a method on the logind D-Bus service.
func CallLogind(method string) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		slog.Error("failed to connect to system bus", "error", err)
		return fmt.Errorf("failed to connect to system bus: %w", err)
	}
	slog.Debug("connected to system bus")
	defer conn.Close()

	obj := conn.Object(dbusDest, dbus.ObjectPath(dbusPath))

	// the boolean argument is for "interactive" (polkit dialog)
	if call := obj.Call(dbusInterface+"."+method, 0, true); call.Err != nil {
		slog.Error("dbus call failed", "method", method, "error", call.Err)
		return fmt.Errorf("dbus call to %s failed: %w", method, call.Err)
	}
	slog.Debug("dbus call successful")
	return nil
}
