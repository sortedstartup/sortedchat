package dao

type Dao interface {
}

type DaoImpl struct {
}

func NewDao() Dao {
	return &DaoImpl{}
}

func LoadConfig() (*Config, error) {
	return &Config{}, nil
}

type Config struct {
}

func NewDAOFactory(config *Config) (DAOFactory, error) {
	return &DAOFactoryImpl{}, nil
}

type DAOFactory interface {
	Close() error
}

type DAOFactoryImpl struct {
}

func (f *DAOFactoryImpl) Close() error {
	return nil
}
