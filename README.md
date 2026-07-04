# Agent for ApexOS

This is the OS Agent for ApexOS. It is used for ApexOS
OS and ApexOS Supervised installation types and it allows the
ApexOS Supervisor to communicate with the host operating system.

## Installation & Update

### Using the ApexOS Operating System

The OS Agent is pre-installed with the ApexOS Operating System.

Updates are part of the ApexOS Operating System updates, which
the ApexOS UI will offer to upgrade to when there is a new version
available.

### Using ApexOS Supervised on Debian

Download the latest Debian package from OS Agent GitHub release page at:

<https://github.com/apexinfosysindia/os-agent/releases/latest>

Next, install (or update) the downloaded Debian package using:

```shell
sudo dpkg -i os-agent_1.0.0_linux_x86_64.deb
```

Note: Replace the `deb` file in the above example with the file you
have downloaded from the releases page.

You can test if the installation was successful by running:

```bash
busctl introspect --system io.apexos.os /io/apexos/os
```

This should **not** return an error. If you get an object introspection
with `io.apexos.os`, `interface` etc. OS Agent is working as expected.

## Uninstall

To remove OS Agent from your system use the Debian packaging system:

```shell
sudo dpkg -r os-agent
```

## Development

### Compile

```shell
go build -ldflags "-X main.version="
```

### Tests

```shell
gdbus introspect --system --dest io.apexos.os --object-path /io/apexos/os
gdbus call --system --dest io.apexos.os --object-path /io/apexos/os/Boards/Yellow --method org.freedesktop.DBus.Properties.Set io.apexos.os.Boards.Yellow PowerLED "<false>"
```
