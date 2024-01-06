package queryplan

import "database/sql/driver"

type queryPlanDriver struct {
	sqlDriver driver.Driver
}

func (qd *queryPlanDriver) Open(name string) (driver.Conn, error) {
	conn, err := qd.sqlDriver.Open(name)
	if err != nil {
		return nil, err
	}

	return &queryPlanConn{conn: conn}, nil
}
