package main

import (
	"os"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/getsentry/sentry-go"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"github.com/apexinfosysindia/os-agent/apparmor"
	"github.com/apexinfosysindia/os-agent/boards"
	"github.com/apexinfosysindia/os-agent/cgroup"
	"github.com/apexinfosysindia/os-agent/config/swap"
	"github.com/apexinfosysindia/os-agent/config/timesyncd"
	"github.com/apexinfosysindia/os-agent/config/usbip"
	"github.com/apexinfosysindia/os-agent/datadisk"
	"github.com/apexinfosysindia/os-agent/system"
	logging "github.com/apexinfosysindia/os-agent/utils/log"
)

const (
	busName    = "io.apexos.os"
	objectPath = "/io/apexos/os"
)

var (
	// version and baord are set at link time via -X flag
	version       string = "dev"
	board         string = "unknown"
	enableCapture bool   = false
)

func main() {
	logging.Info.Printf("Start OS-Agent %s", version)

	// Sentry: crash reporting is DISABLED unless the OS_AGENT_SENTRY_DSN
	// environment variable carries a DSN — the upstream hardcoded endpoint
	// (third-party account coupling, initialized unconditionally at every
	// start) was removed; pointing at an own Sentry instance later is a pure
	// config change, no rebuild.
	if dsn := os.Getenv("OS_AGENT_SENTRY_DSN"); dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:        dsn,
			Release:    version,
			BeforeSend: filterSentry,
		})
		if err != nil {
			logging.Critical.Fatalf("Sentry init: %s", err)
		}

		defer sentry.Flush(2 * time.Second)
		defer sentry.Recover()
	} else {
		logging.Info.Printf("Sentry disabled: OS_AGENT_SENTRY_DSN is not set")
	}

	// Connect DBus
	conn, err := dbus.SystemBus()
	if err != nil {
		logging.Critical.Fatalf("DBus connection: %s", err)
	}

	// Init Dbus io.apexos.os
	reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil {
		logging.Critical.Panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		logging.Critical.Fatalf("name already taken")
	}

	// Set base Property / functionality
	InitializeDBus(conn)

	logging.Info.Printf("Listening on service %s ...", busName)
	datadisk.InitializeDBus(conn)
	system.InitializeDBus(conn)
	apparmor.InitializeDBus(conn)
	cgroup.InitializeDBus(conn)
	boards.InitializeDBus(conn, board)
	swap.InitializeDBus(conn)
	timesyncd.InitializeDBus(conn)
	usbip.InitializeDBus(conn)

	_, err = daemon.SdNotify(false, daemon.SdNotifyReady)
	if err != nil {
		logging.Critical.Panic(err)
	}
	select {}
}

func InitializeDBus(conn *dbus.Conn) {
	propsSpec := map[string]map[string]*prop.Prop{
		busName: {
			"Version": {
				Value:    version,
				Writable: false,
				Emit:     prop.EmitInvalidates,
				Callback: nil,
			},
			"Diagnostics": {
				Value:    enableCapture,
				Writable: true,
				Emit:     prop.EmitTrue,
				Callback: func(c *prop.Change) *dbus.Error {
					logging.Info.Printf("Diagnostics is now %t", c.Value)
					enableCapture = c.Value.(bool)
					return nil
				},
			},
		},
	}

	props, err := prop.Export(conn, objectPath, propsSpec)
	if err != nil {
		logging.Critical.Panic(err)
	}

	node := &introspect.Node{
		Name: objectPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:       busName,
				Properties: props.Introspection(busName),
			},
		},
	}
	err = conn.Export(introspect.NewIntrospectable(node), objectPath, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		logging.Critical.Panic(err)
	}

	logging.Info.Printf("Exposing object %s with interface %s ...", objectPath, busName)
}
