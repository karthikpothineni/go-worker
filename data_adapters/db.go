package dataAdapters

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // Blank import for gorm-mysql

	"go-worker/config"
)

var mysqlDB *gorm.DB

// Init  - initializes the database connection
func Init() {

	var err error
	c := config.GetConfig()

	// create aurora-mysql connection for data-team's datastore
	mysqlConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.GetString("mysql.db_username"), c.GetString("mysql.db_password"),
		c.GetString("mysql.db_host"), c.GetString("mysql.db_port"),
		c.GetString("mysql.db_name"))

	mysqlDB, err = gorm.Open(c.GetString("mysql.db_type"), mysqlConnectionString)
	if err != nil {
		panic("Can't connect to mysql database, check config!" + err.Error())
	}
	mysqlConnLifeTime := c.GetInt("mysql.conn_life_time")
	mysqlDB.DB().SetConnMaxLifetime(time.Minute * time.Duration(mysqlConnLifeTime))
}
