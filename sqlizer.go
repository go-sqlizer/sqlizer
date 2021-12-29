package sqlizer

import "github.com/Supersonido/sqlizer/drivers"

type ConnectionConfig struct {
}

type Config struct {
	Connection        ConnectionConfig
	ModelsInit        []func()
	ModelsAssociation []func()
}

var Conn drivers.Driver

func (c Config) Init() {
	Conn = drivers.Postgres{}
	Conn.Connect(drivers.Config{})

	for _, Init := range c.ModelsInit {
		Init()
	}

	for _, Association := range c.ModelsAssociation {
		Association()
	}
}
