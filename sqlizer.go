package sqlizer

import "C"
import "github.com/Supersonido/sqlizer/drivers"

type ConnectionConfig struct {
}

type Config struct {
	Connection        drivers.Config
	ModelsInit        []func(drivers.Driver)
	ModelsAssociation []func()
}

var Conn drivers.Driver

func (c Config) Init() drivers.Driver {
	switch c.Connection.Dialect {
	case "postgres":
		Conn = &drivers.Postgres{}
	default:
		panic("Invalid dialect")
	}

	err := Conn.Connect(c.Connection)
	panicOnError(err)

	for _, Init := range c.ModelsInit {
		Init(Conn)
	}

	for _, Association := range c.ModelsAssociation {
		Association()
	}

	return Conn
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
