package drivers

type Config struct {
	Dialect         string
	Url             string
	ConnectionPool  int
	StartPoolOnBoot bool
}
