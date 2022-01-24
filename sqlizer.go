package sqlizer

import "C"
import (
	"github.com/go-sqlizer/sqlizer/drivers"
	"github.com/go-sqlizer/sqlizer/model"
)

type ConnectionConfig struct {
}

type ModelsInit []func(drivers.Driver) *model.Model
type Models map[string]*model.Model
type ModelsAssociation []func(Models)

type Config struct {
	Connection        drivers.Config
	ModelsInit        ModelsInit
	ModelsAssociation ModelsAssociation
}

func (c Config) Init() drivers.Driver {
	var Conn drivers.Driver

	switch c.Connection.Dialect {
	case "postgres":
		Conn = &drivers.Postgres{}
	default:
		panic("Invalid dialect")
	}

	err := Conn.Connect(c.Connection)
	panicOnError(err)

	models := make(Models)
	for _, Init := range c.ModelsInit {
		modelInstance := Init(Conn)
		models[modelInstance.Name] = modelInstance
	}

	for _, Association := range c.ModelsAssociation {
		Association(models)
	}

	return Conn
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
