package config

// Configuration is the exportable type of the node configuration
type Configuration struct {
	Version string
	Logger  struct {
		Format string `default:"default"`
		Debug  bool   `default:"false"`
	}
	Global struct {
		SSLCert string
		SSLKey  string
		Message string `default:"a nice person"`
	}
	Storage struct {
		DiskStore struct {
			ImageDir string `default:"/var/lib/uspeak/data/images" env:"IMAGE_PATH"`
			KeyDir   string `default:"/var/lib/uspeak/data/keys" env:"KEY_PATH"`
			PostDir  string `default:"/var/lib/uspeak/data/posts" env:"POST_PATH"`
		}
		BoltStore struct {
			ImagePath string `default:"/var/lib/uspeak/data/image.db" env:"IMAGE_PATH"`
			KeyPath   string `default:"/var/lib/uspeak/data/key.db" env:"KEY_PATH"`
			PostPath  string `default:"/var/lib/uspeak/data/post.db" env:"POST_PATH"`
		}
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
			Port          int    `default:"3000" env:"API_PORT"`
			Interface     string `default:"127.0.0.1"`
			AdminEnabled  bool   `default:"false"`
			AdminUser     string `default:"admin"`
			AdminPassword string `default:"admin"`
		}
	}
}
