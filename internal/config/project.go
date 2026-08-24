package config

// Project is a sshmark project bookmark.
type Project struct {
	Name      string   `toml:"name"`
	Host      string   `toml:"host"`
	User      string   `toml:"user"`
	Port      int      `toml:"port,omitempty"`
	Directory string   `toml:"directory"`
	Tunnels   []Tunnel `toml:"tunnels,omitempty"`
}

// Tunnel is a local-to-remote port forward.
type Tunnel struct {
	LocalPort   int    `toml:"local_port"`
	RemoteHost  string `toml:"remote_host,omitempty"`
	RemotePort  int    `toml:"remote_port"`
	Description string `toml:"description,omitempty"`
}
