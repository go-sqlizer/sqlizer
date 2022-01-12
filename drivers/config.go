package drivers

type Config struct {
	Dialect         string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSl             string
	ConnectionPool  int
	StartPoolOnBoot bool
}