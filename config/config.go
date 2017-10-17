package config

// Configuration is the exportable type of the node configuration
type Configuration struct {
	Logger struct {
		Format string `default:"default"`
		Debug  bool   `default:"false"`
	}
	Global struct {
		SSLCert string
		SSLKey  string
	}
	NodeNetwork struct {
		Port      int    `default:"6969" env:"NODE_PORT"`
		Interface string `default:"127.0.0.1"`
	}
	Web struct {
		Static struct {
			Port      int    `default:"4000"`
			Interface string `default:"127.0.0.1"`
			Directory string `default:"portal/dist" env:"STATIC_DIR"`
		}
		API struct {
			Port      int    `default:"3000"`
			Interface string `default:"127.0.0.1"`
		}
	}
}
